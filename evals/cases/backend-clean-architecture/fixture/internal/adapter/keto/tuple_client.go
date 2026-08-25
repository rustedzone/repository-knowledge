package keto

import (
	"context"

	"example.invalid/access-service/internal/domain"
)

type KetoTupleClient struct{}

func NewKetoTupleClient() *KetoTupleClient { return &KetoTupleClient{} }

func (client *KetoTupleClient) ReplaceApprovalTuple(ctx context.Context, approval *domain.Approval) error {
	return nil
}

func (client *KetoTupleClient) DeleteGroupTuples(ctx context.Context, groupID string) error {
	return nil
}
