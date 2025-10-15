package repository

import (
	"backapp/internal/models"
	"database/sql"
	"errors"
	"strings"
)

type UserRepository interface {
	FindUsers(query string, searchType string) ([]*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	UpdateUserDisplayName(userID string, displayName string) error
	GetUserWithRoles(userID string) (*models.User, error)
	IsEmailWhitelisted(email string) (bool, error)
	GetRoleByEmail(email string) (string, error)
	AddUserRoleIfNotExists(userID string, roleName string) error
	UpdateUserRole(userID string, roleName string, eventID *int) error
	DeleteUserRole(userID string, roleName string) error
}

func (r *userRepository) DeleteUserRole(userID string, roleName string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// ロール名からロールIDを取得
	var roleID int
	err = tx.QueryRow("SELECT id FROM roles WHERE name = ?", roleName).Scan(&roleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// ロールが存在しない場合は、何もせずに成功として扱う
			return nil
		}
		return err
	}

	// user_rolesテーブルからエントリを削除
	_, err = tx.Exec("DELETE FROM user_roles WHERE user_id = ? AND role_id = ?", userID, roleID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetRoleByEmail(email string) (string, error) {
	var roleName string
	err := r.db.QueryRow("SELECT role FROM whitelisted_emails WHERE email = ? AND event_id IS NULL", email).Scan(&roleName)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("email not in whitelist") // Not in whitelist, return error
		}
		return "", err // Other DB error
	}
	return roleName, nil
}

func (r *userRepository) AddUserRoleIfNotExists(userID string, roleName string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// role名からrolesテーブルのIDを取得
	var roleID int
	err = tx.QueryRow("SELECT id FROM roles WHERE name = ?", roleName).Scan(&roleID)
	if err != nil {
		if err == sql.ErrNoRows {
			// ロールが存在しない場合は何もせず終了
			return nil
		}
		return err
	}

	// ユーザーにそのロールが既に割り当てられているか確認
	var exists int
	err = tx.QueryRow("SELECT COUNT(*) FROM user_roles WHERE user_id = ? AND role_id = ?", userID, roleID).Scan(&exists)
	if err != nil {
		return err
	}

	// ロールがまだ割り当てられていない場合のみ挿入
	if exists == 0 {
		_, err = tx.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, roleID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *userRepository) IsEmailWhitelisted(email string) (bool, error) {
	var role string
	err := r.db.QueryRow("SELECT role FROM whitelisted_emails WHERE email = ?", email).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Not whitelisted, but not an error
		}
		return false, err // Other DB error
	}
	return true, nil // Whitelisted
}

func (r *userRepository) FindUsers(query string, searchType string) ([]*models.User, error) {
	baseQuery := "SELECT id, email, display_name, class_id, is_profile_complete, created_at, updated_at FROM users"
	var args []interface{}

	if query != "" {
		switch searchType {
		case "email":
			baseQuery += " WHERE email LIKE ?"
			args = append(args, "%"+query+"%")
		case "display_name":
			baseQuery += " WHERE display_name LIKE ?"
			args = append(args, "%"+query+"%")
		}
	}

	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	var userIDs []interface{}
	for rows.Next() {
		user := &models.User{}
		var tempClassID sql.NullInt32
		var tempDisplayName sql.NullString

		err := rows.Scan(&user.ID, &user.Email, &tempDisplayName, &tempClassID, &user.IsProfileComplete, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if tempDisplayName.Valid {
			user.DisplayName = &tempDisplayName.String
		} else {
			user.DisplayName = nil
		}
		if tempClassID.Valid {
			val := int(tempClassID.Int32)
			user.ClassID = &val
		} else {
			user.ClassID = nil
		}
		users = append(users, user)
		userIDs = append(userIDs, user.ID)
	}

	if len(users) == 0 {
		return users, nil
	}

	// Fetch roles for all users in a single query
	rolesQuery := `
		SELECT ur.user_id, r.id, r.name
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id IN (?` + strings.Repeat(",?", len(userIDs)-1) + `)`

	roleRows, err := r.db.Query(rolesQuery, userIDs...)
	if err != nil {
		return nil, err
	}
	defer roleRows.Close()

	rolesMap := make(map[string][]models.Role)
	for roleRows.Next() {
		var userID string
		var role models.Role
		if err := roleRows.Scan(&userID, &role.ID, &role.Name); err != nil {
			return nil, err
		}
		rolesMap[userID] = append(rolesMap[userID], role)
	}

	// Assign roles to users
	for _, user := range users {
		if roles, ok := rolesMap[user.ID]; ok {
			user.Roles = roles
		}
	}

	return users, nil
}

func (r *userRepository) GetUserByEmail(email string) (*models.User, error) {
	row := r.db.QueryRow("SELECT id, email, display_name, class_id, is_profile_complete, created_at, updated_at FROM users WHERE email = ?", email)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Email, &user.DisplayName, &user.ClassID, &user.IsProfileComplete, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, err
	}
	return user, nil
}

func (r *userRepository) CreateUser(user *models.User) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// ユーザーをusersテーブルに挿入
	_, err = tx.Exec("INSERT INTO users (id, email, display_name, class_id, is_profile_complete) VALUES (?, ?, ?, ?, ?)",
		user.ID, user.Email, user.DisplayName, user.ClassID, user.IsProfileComplete)
	if err != nil {
		return err
	}

	// emailに基づいてwhitelisted_emailsからroleを取得
	var roleName string
	err = r.db.QueryRow("SELECT role FROM whitelisted_emails WHERE email = ? AND event_id IS NULL", user.Email).Scan(&roleName)
	if err != nil {
		// ホワイトリストにない場合はエラーを返す
		if err == sql.ErrNoRows {
			return errors.New("email not in whitelist")
		} else {
			return err // その他のDBエラー
		}
	}

	// role名からrolesテーブルのIDを取得
	var roleID int
	err = r.db.QueryRow("SELECT id FROM roles WHERE name = ?", roleName).Scan(&roleID)
	if err != nil {
		// rolesテーブルに該当roleがない場合はエラー
		return err
	}

	// user_rolesテーブルにマッピングを挿入
	_, err = tx.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", user.ID, roleID)
	if err != nil {
		return err
	}

	// トランザクションをコミット
	return tx.Commit()
}

func (r *userRepository) UpdateUser(user *models.User) error {
	_, err := r.db.Exec("UPDATE users SET display_name = ?, class_id = ?, is_profile_complete = ? WHERE id = ?",
		user.DisplayName, user.ClassID, user.IsProfileComplete, user.ID)
	return err
}

func (r *userRepository) UpdateUserDisplayName(userID string, displayName string) error {
	_, err := r.db.Exec("UPDATE users SET display_name = ? WHERE id = ?", displayName, userID)
	return err
}

func (r *userRepository) GetUserWithRoles(userID string) (*models.User, error) {
	// ユーザー情報を取得
	row := r.db.QueryRow("SELECT id, email, display_name, class_id, is_profile_complete, created_at, updated_at FROM users WHERE id = ?", userID)

	user := &models.User{}
	var tempClassID sql.NullInt32
	var tempDisplayName sql.NullString

	err := row.Scan(&user.ID, &user.Email, &tempDisplayName, &tempClassID, &user.IsProfileComplete, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, err
	}

	// sql.NullXXXXからUserモデルに値をコピー
	if tempDisplayName.Valid {
		user.DisplayName = &tempDisplayName.String
	} else {
		user.DisplayName = nil
	}
	if tempClassID.Valid {
		val := int(tempClassID.Int32)
		user.ClassID = &val
	} else {
		user.ClassID = nil
	}

	// ユーザーのロール情報を取得
	rows, err := r.db.Query(`
		SELECT r.id, r.name 
		FROM roles r 
		INNER JOIN user_roles ur ON r.id = ur.role_id 
		WHERE ur.user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		err := rows.Scan(&role.ID, &role.Name)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	user.Roles = roles
	return user, nil
}

func (r *userRepository) UpdateUserRole(userID string, roleName string, eventID *int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 新しいロールのIDを取得
	var roleID int64
	err = tx.QueryRow("SELECT id FROM roles WHERE name = ?", roleName).Scan(&roleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Role does not exist, so create it
			result, err := tx.Exec("INSERT INTO roles (name) VALUES (?)", roleName)
			if err != nil {
				return err
			}
			roleID, err = result.LastInsertId()
			if err != nil {
				return err
			}
		} else {
			// Another database error occurred
			return err
		}
	}

	// REPLACE INTOを使用して、ロールの割り当てをアトミックに行う
	// これにより、(user_id, role_id)の組み合わせが既存の場合、event_idが更新される
	if eventID != nil {
		_, err = tx.Exec("REPLACE INTO user_roles (user_id, role_id, event_id) VALUES (?, ?, ?)", userID, roleID, *eventID)
	} else {
		_, err = tx.Exec("REPLACE INTO user_roles (user_id, role_id, event_id) VALUES (?, ?, NULL)", userID, roleID)
	}

	if err != nil {
		return err
	}

	return tx.Commit()
}
