package repository

import (
	"backapp/internal/models"
	"database/sql"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
)

func BenchmarkCachedReadHits(b *testing.B) {
	for _, name := range []string{"active-event", "user-roles"} {
		b.Run(name, func(b *testing.B) {
			previous := GlobalCache
			GlobalCache = cache.New(time.Minute, 0)
			defer func() { GlobalCache = previous }()
			db := &sql.DB{}
			var read func()
			if name == "active-event" {
				load := func() (int, error) { return 7, nil }
				read = func() {
					if _, err := cachedRead(db, "events", "active", load); err != nil {
						b.Fatal(err)
					}
				}
			} else {
				displayName, classID := "Student", 1
				user := &models.User{ID: "alice", DisplayName: &displayName, ClassID: &classID, NotificationFilters: []string{"general"}, Roles: []models.Role{{ID: 1, Name: "student"}}}
				load := func() (*models.User, error) { return user, nil }
				read = func() {
					if _, err := cachedRead(db, "user:alice", "7", load); err != nil {
						b.Fatal(err)
					}
				}
			}
			read()
			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					read()
				}
			})
		})
	}
}
