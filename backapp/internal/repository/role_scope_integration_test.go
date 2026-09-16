package repository

import (
	"database/sql"
	"os"
	"testing"

	"backapp/internal/models"

	"github.com/stretchr/testify/require"
)

func eventRoleDB(t *testing.T) *sql.DB {
	t.Helper()
	db := seasonalMembershipDB(t)
	_, err := db.Exec(`
		ALTER TABLE users ADD notification_filters JSON DEFAULT ('["general"]');
		CREATE TABLE roles (id INT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(50) NOT NULL UNIQUE);
		CREATE TABLE user_roles (
			user_id VARCHAR(36) NOT NULL, role_id INT NOT NULL, event_id INT NULL,
			PRIMARY KEY (user_id, role_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
			FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
		);
		CREATE TABLE sports (id INT PRIMARY KEY, name VARCHAR(255) NOT NULL);
		CREATE TABLE teams (
			id INT PRIMARY KEY, class_id INT NOT NULL, sport_id INT NOT NULL,
			FOREIGN KEY (class_id) REFERENCES classes(id),
			FOREIGN KEY (sport_id) REFERENCES sports(id)
		);
		CREATE TABLE team_members (
			team_id INT NOT NULL, user_id VARCHAR(36) NOT NULL,
			PRIMARY KEY (team_id, user_id),
			FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
		CREATE TABLE push_subscriptions (
			id INT PRIMARY KEY AUTO_INCREMENT, user_id VARCHAR(36), endpoint VARCHAR(255)
		);
		CREATE TABLE notifications (
			id INT PRIMARY KEY AUTO_INCREMENT, title VARCHAR(100), body TEXT, type VARCHAR(50),
			created_by VARCHAR(36), event_id INT NULL, created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE notification_targets (
			notification_id INT, role_name VARCHAR(50), PRIMARY KEY (notification_id, role_name)
		);
		CREATE TABLE notification_recipients (
			notification_id INT, user_id VARCHAR(36), PRIMARY KEY (notification_id, user_id)
		);
		INSERT INTO roles VALUES (1, 'root'), (2, 'admin'), (3, 'student'),
			(4, 'IS4_Basketball'), (5, 'IS4_Tennis'), (6, 'IS4_Volleyball'), (7, 'scorekeeper');
		INSERT INTO sports VALUES (1, 'Basketball'), (2, 'Tennis'), (3, 'Volleyball');
		INSERT INTO teams VALUES (100, 11, 1), (101, 11, 2), (200, 21, 1), (201, 21, 3);
		INSERT INTO team_members VALUES ('100', 'spring'), ('101', 'spring'),
			('200', 'autumn'), ('201', 'autumn'), ('100', 'incomplete'), ('200', 'incomplete');
		INSERT INTO user_roles (user_id, role_id, event_id) SELECT id, 3, NULL FROM users;
		INSERT INTO user_roles VALUES ('spring', 4, 1), ('spring', 5, NULL),
			('spring', 7, NULL), ('autumn', 4, 2), ('autumn', 6, 2), ('incomplete', 4, 2);
		INSERT INTO active_event VALUES (1, 1);
		INSERT INTO push_subscriptions (user_id, endpoint) VALUES
			('spring', 'https://push.example/spring'), ('autumn', 'https://push.example/autumn'),
			('incomplete', 'https://push.example/both');
	`)
	require.NoError(t, err)
	migration, err := os.ReadFile("../../db/migrations/0020_scope_user_roles_by_event.up.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(migration))
	require.NoError(t, err)
	return db
}

func roleNamesForTest(roles []models.Role) []string {
	names := make([]string, 0, len(roles))
	for _, role := range roles {
		names = append(names, role.Name)
	}
	return names
}

func TestEventRolesMigrationAndIndependentAssignmentsMySQL(t *testing.T) {
	db := eventRoleDB(t)
	repo := NewUserRepository(db)
	springID, autumnID := 1, 2
	var count int
	// 同じロールで上書きされていた春の割り当てをチーム登録から復元する。
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM user_roles WHERE user_id = 'incomplete' AND role_id = 4").Scan(&count))
	require.Equal(t, 2, count)
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM user_roles WHERE role_id = 5 AND event_id IS NULL").Scan(&count))
	require.Zero(t, count)
	user, err := repo.GetUserWithRoles("spring")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"student", "IS4_Basketball", "IS4_Tennis", "scorekeeper"}, roleNamesForTest(user.Roles))

	// 秋に同じロールを繰り返し付与しても、春の記録と共通ロールを維持する。
	for range 2 {
		require.NoError(t, repo.UpdateUserRole("spring", "IS4_Basketball", &autumnID))
		require.NoError(t, repo.UpdateUserRole("spring", "scorekeeper", nil))
	}
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM user_roles WHERE user_id = 'spring' AND role_id = 4").Scan(&count))
	require.Equal(t, 2, count)
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM user_roles WHERE user_id = 'spring' AND role_id = 7").Scan(&count))
	require.Equal(t, 1, count)
	require.NoError(t, repo.DeleteUserRole("spring", "IS4_Basketball", &autumnID))
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM user_roles WHERE user_id = 'spring' AND role_id = 4 AND event_id = ?", springID).Scan(&count))
	require.Equal(t, 1, count)

	// NULL指定の解除は同名の大会別割り当てを巻き込まない。
	require.NoError(t, repo.UpdateUserRole("spring", "scorekeeper", &autumnID))
	require.NoError(t, repo.DeleteUserRole("spring", "scorekeeper", nil))
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM user_roles WHERE user_id = 'spring' AND role_id = 7 AND event_id = 2").Scan(&count))
	require.Equal(t, 1, count)
	require.NoError(t, repo.AddUserRoleIfNotExists("spring", "scorekeeper"))
	require.NoError(t, repo.AddUserRoleIfNotExists("spring", "scorekeeper"))
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM user_roles WHERE user_id = 'spring' AND role_id = 7").Scan(&count))
	require.Equal(t, 2, count)
	_, err = db.Exec("UPDATE active_event SET event_id = 2")
	require.NoError(t, err)
	user, err = repo.GetUserWithRoles("spring")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"student", "scorekeeper"}, roleNamesForTest(user.Roles))

	// 一意制約がNULLの共通割り当ても守り、外部キーのCASCADEも動く。
	_, err = db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES ('spring', 7)")
	require.Error(t, err)
	_, err = db.Exec("DELETE FROM users WHERE id = 'incomplete'")
	require.NoError(t, err)
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM user_roles WHERE user_id = 'incomplete'").Scan(&count))
	require.Zero(t, count)
	_, err = db.Exec("INSERT INTO events VALUES (99, 2030, 'upcoming'); INSERT INTO user_roles (user_id, role_id, event_id) VALUES ('spring', 7, 99); DELETE FROM events WHERE id = 99")
	require.NoError(t, err)
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM user_roles WHERE event_id = 99").Scan(&count))
	require.Zero(t, count)
}

func TestEventRolesFollowCurrentEventMySQL(t *testing.T) {
	db := eventRoleDB(t)
	users := NewUserRepository(db)
	events := NewEventRepository(db)
	roles := NewRoleRepository(db)
	notifications := NewNotificationRepository(db)
	require.NoError(t, users.ReplaceMasterRole("unassigned", "root"))
	autumnID := 2
	require.NoError(t, events.SetActiveEvent(&autumnID))
	user, err := users.GetUserWithRoles("spring")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"student", "scorekeeper"}, roleNamesForTest(user.Roles))
	require.Equal(t, 21, *user.ClassID)
	root, err := users.GetUserWithRoles("unassigned")
	require.NoError(t, err)
	require.Equal(t, []string{"root"}, roleNamesForTest(root.Roles))
	found, err := users.FindUsers("spring@example.com", "email")
	require.NoError(t, err)
	require.Len(t, found, 1)
	require.ElementsMatch(t, []string{"student", "scorekeeper"}, roleNamesForTest(found[0].Roles))
	available, err := roles.GetAllRoles()
	require.NoError(t, err)
	require.NotContains(t, roleNamesForTest(available), "IS4_Tennis")
	require.Contains(t, roleNamesForTest(available), "IS4_Volleyball")
	require.Contains(t, roleNamesForTest(available), "root")
	require.Contains(t, roleNamesForTest(available), "admin")

	ids, err := notifications.GetUserIDsByRoles([]string{"IS4_Basketball"}, &autumnID)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"autumn", "incomplete"}, ids)
	stats, err := notifications.GetPushSubscriptionStatsByTargets([]string{"IS4_Basketball"}, nil)
	require.NoError(t, err)
	require.EqualValues(t, 2, stats.TargetUserCount)
	require.EqualValues(t, 2, stats.SubscribedUserCount)
	require.EqualValues(t, 2, stats.SubscriptionEndpointCount)
	stats, err = notifications.GetPushSubscriptionStatsByTargets([]string{"IS4_Basketball"}, []string{"spring"})
	require.NoError(t, err)
	require.EqualValues(t, 3, stats.TargetUserCount)

	// 春に作成した通知は、切り替え後も春のロール所有者へ送る。
	springID := 1
	ids, err = notifications.GetUserIDsByRoles([]string{"IS4_Basketball"}, &springID)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"spring", "incomplete"}, ids)
	require.NoError(t, events.SetActiveEvent(&springID))
	user, err = users.GetUserWithRoles("spring")
	require.NoError(t, err)
	require.Contains(t, roleNamesForTest(user.Roles), "IS4_Tennis")

	// 有効大会がないときは共通ロールだけを返す。
	require.NoError(t, events.SetActiveEvent(nil))
	user, err = users.GetUserWithRoles("spring")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"student", "scorekeeper"}, roleNamesForTest(user.Roles))
	available, err = roles.GetAllRoles()
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"root", "admin", "student", "scorekeeper"}, roleNamesForTest(available))
	ids, err = notifications.GetUserIDsByRoles([]string{"IS4_Basketball"}, nil)
	require.NoError(t, err)
	require.Empty(t, ids)
	ids, err = notifications.GetUserIDsByRoles([]string{"root"}, nil)
	require.NoError(t, err)
	require.Equal(t, []string{"unassigned"}, ids)
}

func TestEventRoleNotificationAccessMySQL(t *testing.T) {
	db := eventRoleDB(t)
	_, err := db.Exec(`
		UPDATE active_event SET event_id = 2;
		INSERT INTO notifications (id, title, body, type, created_by, event_id) VALUES
			(1, '春の競技連絡', '本文', 'general', 'unassigned', 1),
			(2, '秋の競技連絡', '本文', 'general', 'unassigned', 2),
			(3, '全体連絡', '本文', 'general', 'unassigned', NULL),
			(4, '個人連絡', '本文', 'general', 'unassigned', 1);
		INSERT INTO notification_targets VALUES (1, 'IS4_Basketball'), (2, 'IS4_Basketball'), (3, 'student');
		INSERT INTO notification_recipients VALUES (4, 'autumn');
	`)
	require.NoError(t, err)
	repo := NewNotificationRepository(db)
	list, err := repo.GetNotificationsForAccess([]string{"student", "IS4_Basketball"}, "autumn", false, 50)
	require.NoError(t, err)
	var ids []int
	for _, notification := range list {
		ids = append(ids, notification.ID)
	}
	// 秋だけの参加者には春の同名ロール宛の通知を公開しない。個人宛は維持する。
	require.ElementsMatch(t, []int{2, 3, 4}, ids)
	list, err = repo.GetNotificationsForAccess([]string{"student", "IS4_Basketball"}, "incomplete", false, 50)
	require.NoError(t, err)
	ids = nil
	for _, notification := range list {
		ids = append(ids, notification.ID)
	}
	require.ElementsMatch(t, []int{1, 2, 3}, ids)
	list, err = repo.GetNotificationsForAccess([]string{"student"}, "unassigned", true, 50)
	require.NoError(t, err)
	require.Len(t, list, 4)
}
