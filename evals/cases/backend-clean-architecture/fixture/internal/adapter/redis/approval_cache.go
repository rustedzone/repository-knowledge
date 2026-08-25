package redis

import "context"

type ApprovalCache struct{}

func NewApprovalCache() *ApprovalCache { return &ApprovalCache{} }

func (cache *ApprovalCache) Invalidate(ctx context.Context, approvalID string) error {
	return nil
}
