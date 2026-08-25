package postgres

import (
	"context"
	"database/sql"
)

type PostgresGroupRepository struct {
	database *sql.DB
}

func NewPostgresGroupRepository(database *sql.DB) *PostgresGroupRepository {
	return &PostgresGroupRepository{database: database}
}

func (repository *PostgresGroupRepository) Delete(ctx context.Context, groupID string) error {
	_, err := repository.database.ExecContext(ctx, "DELETE FROM user_group WHERE id = $1", groupID)
	return err
}
