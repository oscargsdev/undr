package postgres

import (
	"context"

	"github.com/lib/pq"
	"github.com/oscargsdev/undr/internal/identity/domain"
)

func (r *repository) GetAllRolesForUser(ctx context.Context, userID int64) (domain.Roles, error) {
	query := `
		SELECT roles.code
		FROM roles
		INNER JOIN users_roles ON users_roles.role_id = roles.id
		INNER JOIN users ON users_roles.user_id = users.id
		WHERE users.id = $1`

	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles domain.Roles

	for rows.Next() {
		var role string

		err := rows.Scan(&role)
		if err != nil {
			return nil, err
		}

		roles = append(roles, role)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *repository) AddRoleForUser(ctx context.Context, userID int64, codes ...string) error {
	query := `
        INSERT INTO users_roles
        SELECT $1, roles.id FROM roles WHERE roles.code = ANY($2)`

	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	_, err := r.db.ExecContext(ctx, query, userID, pq.Array(codes))
	return err
}
