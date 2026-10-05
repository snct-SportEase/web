package repository

import (
	"backapp/internal/models"
	"database/sql"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
)

func isolateAdversarialReadCache(t *testing.T) {
	t.Helper()
	previous := GlobalCache
	GlobalCache = cache.New(time.Minute, 0)
	t.Cleanup(func() { GlobalCache = previous })
}

func TestCachedReadOverlappingRevisionsKeepNewestSnapshot(t *testing.T) {
	isolateAdversarialReadCache(t)
	db := &sql.DB{}
	type result struct {
		value int
		err   error
	}
	var releases, completions []chan struct{}
	t.Cleanup(func() {
		for _, release := range releases {
			select {
			case <-release:
			default:
				close(release)
			}
		}
		for _, completed := range completions {
			<-completed
		}
	})
	startRead := func(value int) (chan struct{}, <-chan result) {
		entered, release := make(chan struct{}), make(chan struct{})
		completed := make(chan struct{})
		releases = append(releases, release)
		completions = append(completions, completed)
		done := make(chan result, 1)
		go func() {
			defer close(completed)
			read, err := cachedRead(db, "events", "active", func() (int, error) {
				close(entered)
				<-release
				return value, nil
			})
			done <- result{read, err}
		}()
		select {
		case <-entered:
		case <-time.After(2 * time.Second):
			t.Fatal("new revision joined an older in-flight read")
		}
		return release, done
	}
	oldRelease, oldDone := startRead(1)
	invalidateReads(db, "events")
	middleRelease, middleDone := startRead(2)
	invalidateReads(db, "events")
	newRelease, newDone := startRead(3)
	// Finish in reverse order: both obsolete reads complete after the new one.
	for _, read := range []struct {
		release chan struct{}
		done    <-chan result
		want    int
	}{{newRelease, newDone, 3}, {middleRelease, middleDone, 2}, {oldRelease, oldDone, 1}} {
		close(read.release)
		got := <-read.done
		if got.err != nil || got.value != read.want {
			t.Fatalf("in-flight read = %+v, want %d", got, read.want)
		}
	}
	got, err := cachedRead(db, "events", "active", func() (int, error) {
		t.Error("obsolete completion discarded the newest cached snapshot")
		return 99, nil
	})
	if err != nil || got != 3 {
		t.Fatalf("current snapshot = %d, %v; want 3", got, err)
	}
}

func TestCachedReadIsolationAcrossDatabasesAndViewers(t *testing.T) {
	isolateAdversarialReadCache(t)
	dbA, dbB := &sql.DB{}, &sql.DB{}
	for _, read := range []struct {
		db        *sql.DB
		namespace string
		key       string
		role      string
	}{
		{dbA, "user:alice", "1:alice", "root"},
		{dbB, "user:alice", "1:alice", "student"},
		{dbA, "user:bob", "1:bob", "student"},
		{dbA, "user:alice", "2:alice", "admin"},
	} {
		for i := 0; i < 2; i++ {
			got, err := cachedRead(read.db, read.namespace, read.key, func() (string, error) {
				if i != 0 {
					t.Error("cache miss on repeated isolated read")
				}
				return read.role, nil
			})
			if err != nil || got != read.role {
				t.Fatalf("viewer/DB/event snapshot crossed a boundary: got %q, %v; want %q", got, err, read.role)
			}
		}
	}
}

func TestCachedReadConcurrentCallerMutationsDoNotGrantRoles(t *testing.T) {
	isolateAdversarialReadCache(t)
	db := &sql.DB{}
	load := func() (*models.User, error) {
		name := "Alice"
		return &models.User{DisplayName: &name, Roles: []models.Role{{Name: "student"}}, NotificationFilters: []string{"general"}}, nil
	}
	if _, err := cachedRead(db, "user:alice", "1", load); err != nil {
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	for i := 0; i < 64; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			user, err := cachedRead(db, "user:alice", "1", load)
			if err != nil {
				t.Error(err)
				return
			}
			if user.Roles[0].Name != "student" || *user.DisplayName != "Alice" || user.NotificationFilters[0] != "general" {
				t.Error("another caller mutated this snapshot")
			}
			user.Roles[0].Name = "root"
			*user.DisplayName = "changed"
			user.NotificationFilters[0] = "changed"
		}()
	}
	workers.Wait()
	user, err := cachedRead(db, "user:alice", "1", load)
	if err != nil || user.Roles[0].Name != "student" || *user.DisplayName != "Alice" || user.NotificationFilters[0] != "general" {
		t.Fatal("caller mutation poisoned the shared authorization snapshot")
	}
}

func TestCachedReadEncodingFailureDoesNotPoisonRetry(t *testing.T) {
	isolateAdversarialReadCache(t)
	db := &sql.DB{}
	if _, err := cachedRead(db, "scores", "1", func() (float64, error) { return math.NaN(), nil }); err == nil {
		t.Fatal("non-JSON value unexpectedly succeeded")
	}
	got, err := cachedRead(db, "scores", "1", func() (float64, error) { return 42, nil })
	if err != nil || got != 42 {
		t.Fatalf("encoding failure blocked a valid retry: %v, %v", got, err)
	}
}
