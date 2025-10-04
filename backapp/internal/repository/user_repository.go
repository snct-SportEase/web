package repository

import (
	"backapp/internal/models"
	"database/sql"
)

type UserRepository interface {
	GetUserByEmail(email string) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	GetUserWithRoles(userID string) (*models.User, error)
	IsEmailWhitelisted(email string) (bool, error)
	GetRoleByEmail(email string) (string, error)
	AddUserRoleIfNotExists(userID string, roleName string) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetRoleByEmail(email string) (string, error) {
	var roleName string
	err := r.db.QueryRow("SELECT role FROM whitelisted_emails WHERE email = ?", email).Scan(&roleName)
	if err != nil {
		if err == sql.ErrNoRows {
			return "student", nil // Not in whitelist, default to student
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
	err = r.db.QueryRow("SELECT role FROM whitelisted_emails WHERE email = ?", user.Email).Scan(&roleName)
	if err != nil {
		// ホワイトリストにない場合は'student'ロールを付与
		if err == sql.ErrNoRows {
			roleName = "student"
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
