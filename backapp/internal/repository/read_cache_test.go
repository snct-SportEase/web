package repository

import (
	"backapp/internal/models"
	"database/sql"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
)

func TestCachedReadIsolationAndInvalidation(t *testing.T) {
	previous := GlobalCache
	GlobalCache = cache.New(time.Minute, time.Minute)
	defer func() { GlobalCache = previous }()
	db := &sql.DB{}
	var calls atomic.Int32
	load := func() (*models.User, error) {
		calls.Add(1)
		name := "original"
		return &models.User{DisplayName: &name, Roles: []models.Role{{Name: "student"}}}, nil
	}
	first, err := cachedRead(db, "users", "1:alice", load)
	if err != nil {
		t.Fatal(err)
	}
	*first.DisplayName = "modified"
	first.Roles[0].Name = "root"
	second, err := cachedRead(db, "users", "1:alice", load)
	if err != nil {
		t.Fatal(err)
	}
	if *second.DisplayName != "original" || second.Roles[0].Name != "student" || calls.Load() != 1 {
		t.Fatal("cached values leaked caller mutations")
	}
	invalidateReads(db, "users")
	if _, err := cachedRead(db, "users", "1:alice", load); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatal("write did not invalidate cache")
	}
	if _, err := cachedRead(db, "users", "2:alice", load); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 3 {
		t.Fatal("event keys were not isolated")
	}
}

func TestCachedReadConcurrentMissAndInvalidation(t *testing.T) {
	previous := GlobalCache
	GlobalCache = cache.New(time.Minute, time.Minute)
	defer func() { GlobalCache = previous }()
	db := &sql.DB{}
	entered, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	load := func() (int, error) {
		if calls.Add(1) == 1 {
			close(entered)
		}
		<-release
		return 1, nil
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := cachedRead(db, "events", "active", load); err != nil {
				t.Error(err)
			}
		}()
	}
	<-entered
	close(release)
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatalf("concurrent miss loaded %d times", calls.Load())
	}

	invalidateReads(db, "events")
	entered, release = make(chan struct{}), make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = cachedRead(db, "events", "active", func() (int, error) { close(entered); <-release; return 2, nil })
	}()
	<-entered
	invalidateReads(db, "events")
	close(release)
	<-done
	value, err := cachedRead(db, "events", "active", func() (int, error) { return 3, nil })
	if err != nil || value != 3 {
		t.Fatal("in-flight read repopulated invalidated cache")
	}
}

func TestCachedReadDoesNotCacheErrors(t *testing.T) {
	previous := GlobalCache
	GlobalCache = cache.New(time.Minute, time.Minute)
	defer func() { GlobalCache = previous }()
	db := &sql.DB{}
	if _, err := cachedRead(db, "events", "active", func() (int, error) { return 0, errors.New("unavailable") }); err == nil {
		t.Fatal("expected error")
	}
	value, err := cachedRead(db, "events", "active", func() (int, error) { return 7, nil })
	if err != nil || value != 7 {
		t.Fatal("error was cached")
	}
}
