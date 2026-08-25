package usecase

import (
	"context"
	"fmt"

	"example.invalid/access-service/internal/port"
)

type ApprovalUsecase struct {
	repository port.ApprovalRepository
	cache      port.ApprovalCache
	keto       port.KetoTupleClient
}

func NewApprovalUsecase(repository port.ApprovalRepository, cache port.ApprovalCache, keto port.KetoTupleClient) *ApprovalUsecase {
	return &ApprovalUsecase{repository: repository, cache: cache, keto: keto}
}

func (usecase *ApprovalUsecase) Approve(ctx context.Context, approvalID, actor string) error {
	approval, err := usecase.repository.Get(ctx, approvalID)
	if err != nil {
		return fmt.Errorf("load approval: %w", err)
	}
	if err := approval.Approve(actor); err != nil {
		return err
	}
	if err := usecase.repository.Save(ctx, approval); err != nil {
		return fmt.Errorf("save approved state: %w", err)
	}
	_ = usecase.cache.Invalidate(ctx, approval.ID)

	if err := usecase.keto.ReplaceApprovalTuple(ctx, approval); err != nil {
		_ = approval.MarkFailed()
		_ = usecase.repository.Save(ctx, approval)
		return fmt.Errorf("synchronize keto tuple: %w", err)
	}
	if err := approval.MarkProvisioned(); err != nil {
		return err
	}
	if err := usecase.repository.Save(ctx, approval); err != nil {
		return fmt.Errorf("save provisioned state: %w", err)
	}
	return usecase.cache.Invalidate(ctx, approval.ID)
}
