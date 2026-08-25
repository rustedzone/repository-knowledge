package port

import (
	"context"

	"example.invalid/access-service/internal/domain"
)

type ApprovalRepository interface {
	Get(context.Context, string) (*domain.Approval, error)
	Save(context.Context, *domain.Approval) error
}

type ApprovalCache interface {
	Invalidate(context.Context, string) error
}

type KetoTupleClient interface {
	ReplaceApprovalTuple(context.Context, *domain.Approval) error
	DeleteGroupTuples(context.Context, string) error
}

type GroupRepository interface {
	Delete(context.Context, string) error
}
