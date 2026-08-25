package postgres

import (
	"context"
	"database/sql"

	"example.invalid/access-service/internal/domain"
)

type PostgresApprovalRepository struct {
	database *sql.DB
}

func NewPostgresApprovalRepository(database *sql.DB) *PostgresApprovalRepository {
	return &PostgresApprovalRepository{database: database}
}

func (repository *PostgresApprovalRepository) Get(ctx context.Context, id string) (*domain.Approval, error) {
	approval := &domain.Approval{ID: id}
	err := repository.database.QueryRowContext(ctx, "SELECT status, approved_by FROM approval_step WHERE id = $1", id).Scan(&approval.Status, &approval.ApprovedBy)
	return approval, err
}

func (repository *PostgresApprovalRepository) Save(ctx context.Context, approval *domain.Approval) error {
	_, err := repository.database.ExecContext(ctx, "UPDATE approval_step SET status = $1, approved_by = $2 WHERE id = $3", approval.Status, approval.ApprovedBy, approval.ID)
	return err
}
