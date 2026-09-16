package repository

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

// SQLの結合条件を実際のMySQLで検証する。テスト専用DBは毎回作成・削除する。
func seasonalMembershipDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("SPORTEASE_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("MySQL統合テストには SPORTEASE_TEST_MYSQL_DSN が必要です")
	}
	cfg, err := mysql.ParseDSN(dsn)
	require.NoError(t, err)
	cfg.DBName = ""
	cfg.MultiStatements = true
	cfg.ParseTime = true
	admin, err := sql.Open("mysql", cfg.FormatDSN())
	require.NoError(t, err)
	t.Cleanup(func() { admin.Close() })
	dbName := fmt.Sprintf("sportease_transition_test_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE DATABASE " + dbName)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = admin.Exec("DROP DATABASE " + dbName) })
	cfg.DBName = dbName
	db, err := sql.Open("mysql", cfg.FormatDSN())
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	_, err = db.Exec(`
		CREATE TABLE events (id INT PRIMARY KEY, year INT NOT NULL, status VARCHAR(20) NOT NULL);
		CREATE TABLE active_event (id INT PRIMARY KEY, event_id INT NULL);
		CREATE TABLE classes (
			id INT PRIMARY KEY, event_id INT NOT NULL, name VARCHAR(50) NOT NULL,
			UNIQUE KEY (event_id, name), FOREIGN KEY (event_id) REFERENCES events(id)
		);
		CREATE TABLE users (
			id VARCHAR(36) PRIMARY KEY, email VARCHAR(255) NOT NULL,
			display_name VARCHAR(50) NULL, class_id INT NULL,
			is_profile_complete BOOLEAN NOT NULL DEFAULT FALSE,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (class_id) REFERENCES classes(id)
		);
		INSERT INTO events VALUES (1, 2026, 'archived'), (2, 2026, 'upcoming'),
			(3, 2025, 'archived'), (4, 2027, 'upcoming');
		INSERT INTO classes VALUES (11, 1, 'IS4'), (12, 1, 'IS5'),
			(21, 2, 'IS4'), (22, 2, 'IS5'), (31, 3, 'IS3'), (32, 3, 'IS5'),
			(41, 4, 'IS4'), (42, 4, 'IS5');
		INSERT INTO users (id, email, display_name, class_id, is_profile_complete) VALUES
			('spring', 'spring@example.com', '春のメンバー', 11, TRUE),
			('fifth', 'fifth@example.com', '同年度の５年生', 12, TRUE),
			('autumn', 'autumn@example.com', '秋のメンバー', 21, TRUE),
			('incomplete', 'incomplete@example.com', '設定途中', 11, FALSE),
			('old', 'old@example.com', '前年度のメンバー', 31, TRUE),
			('graduate', 'graduate@example.com', '前年度の５年生', 32, TRUE),
			('unassigned', 'unassigned@example.com', NULL, NULL, FALSE);
	`)
	require.NoError(t, err)
	return db
}

func assertSeasonalProfile(t *testing.T, db *sql.DB, userID string, classID any, name any, complete bool) {
	t.Helper()
	var gotClass sql.NullInt64
	var gotName sql.NullString
	var gotComplete bool
	err := db.QueryRow("SELECT class_id, display_name, is_profile_complete FROM users WHERE id = ?", userID).
		Scan(&gotClass, &gotName, &gotComplete)
	require.NoError(t, err)
	if classID == nil {
		require.False(t, gotClass.Valid)
	} else {
		require.True(t, gotClass.Valid)
		require.EqualValues(t, classID, gotClass.Int64)
	}
	if name == nil {
		require.False(t, gotName.Valid)
	} else {
		require.True(t, gotName.Valid)
		require.Equal(t, name, gotName.String)
	}
	require.Equal(t, complete, gotComplete)
}

func TestSeasonalMembershipTransitionMySQL(t *testing.T) {
	for _, tc := range []struct {
		name   string
		active string
	}{
		{"春から秋へ直接切り替える", "INSERT INTO active_event VALUES (1, 1)"},
		{"春を終了してから秋を選択する", "INSERT INTO active_event VALUES (1, NULL)"},
		{"有効大会のレコードがなくても引き継ぐ", ""},
		{"秋の再選択で春に残った所属を修復する", "INSERT INTO active_event VALUES (1, 2)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := seasonalMembershipDB(t)
			if tc.active != "" {
				_, err := db.Exec(tc.active)
				require.NoError(t, err)
			}
			repo := NewEventRepository(db)
			autumnID := 2
			require.NoError(t, repo.SetActiveEvent(&autumnID))
			assertSeasonalProfile(t, db, "spring", 21, "春のメンバー", true)
			assertSeasonalProfile(t, db, "fifth", 22, "同年度の５年生", true)
			assertSeasonalProfile(t, db, "autumn", 21, "秋のメンバー", true)
			assertSeasonalProfile(t, db, "incomplete", 21, "設定途中", false)
			assertSeasonalProfile(t, db, "old", nil, nil, false)
			assertSeasonalProfile(t, db, "unassigned", nil, nil, false)
			var count int
			require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM users WHERE id = 'graduate'").Scan(&count))
			require.Zero(t, count)
			members, err := NewClassRepository(db).GetClassMembers(21)
			require.NoError(t, err)
			require.Len(t, members, 3)

			// 同年度の春へ戻しても、春秋両方で登録したメンバーを維持する。
			springID := 1
			require.NoError(t, repo.SetActiveEvent(&springID))
			assertSeasonalProfile(t, db, "spring", 11, "春のメンバー", true)
			assertSeasonalProfile(t, db, "autumn", 11, "秋のメンバー", true)
			assertSeasonalProfile(t, db, "fifth", 12, "同年度の５年生", true)

			// 有効大会を解除した状態で年度が変わっても、再登録・卒業処理を行う。
			require.NoError(t, repo.SetActiveEvent(nil))
			nextYearID := 4
			require.NoError(t, repo.SetActiveEvent(&nextYearID))
			assertSeasonalProfile(t, db, "spring", nil, nil, false)
			require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM users WHERE id = 'fifth'").Scan(&count))
			require.Zero(t, count)
		})
	}
}

func TestSeasonalMembershipRepairMigrationMySQL(t *testing.T) {
	db := seasonalMembershipDB(t)
	migration, err := os.ReadFile("../../db/migrations/0019_repair_same_year_class_memberships.up.sql")
	require.NoError(t, err)
	// 有効大会がない場合は既存の所属を変更しない。
	_, err = db.Exec(string(migration))
	require.NoError(t, err)
	assertSeasonalProfile(t, db, "spring", 11, "春のメンバー", true)
	_, err = db.Exec("INSERT INTO active_event VALUES (1, 2)")
	require.NoError(t, err)
	for range 2 {
		_, err = db.Exec(string(migration))
		require.NoError(t, err)
		assertSeasonalProfile(t, db, "spring", 21, "春のメンバー", true)
		assertSeasonalProfile(t, db, "fifth", 22, "同年度の５年生", true)
		assertSeasonalProfile(t, db, "incomplete", 21, "設定途中", false)
		assertSeasonalProfile(t, db, "old", 31, "前年度のメンバー", true)
		assertSeasonalProfile(t, db, "graduate", 32, "前年度の５年生", true)
	}
}
