package repository

import (
	"backapp/internal/models"
	"database/sql"
)

type RoleRepository interface {
	GetAllRoles() ([]models.Role, error)
}

type roleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) GetAllRoles() ([]models.Role, error) {
	rows, err := r.db.Query(`
		SELECT r.id, r.name FROM roles r
		WHERE r.name IN ('root', 'admin', 'student') OR EXISTS (
			SELECT 1 FROM user_roles ur
			WHERE ur.role_id = r.id AND ` + currentRoleAssignmentCondition + `
		) ORDER BY r.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}
