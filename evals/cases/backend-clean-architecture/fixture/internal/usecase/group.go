package usecase

import (
	"context"
	"fmt"

	"example.invalid/access-service/internal/port"
)

type GroupUsecase struct {
	repository port.GroupRepository
	keto       port.KetoTupleClient
}

func NewGroupUsecase(repository port.GroupRepository, keto port.KetoTupleClient) *GroupUsecase {
	return &GroupUsecase{repository: repository, keto: keto}
}

func (usecase *GroupUsecase) DeleteGroup(ctx context.Context, groupID string) error {
	if err := usecase.keto.DeleteGroupTuples(ctx, groupID); err != nil {
		return fmt.Errorf("delete keto tuples: %w", err)
	}
	if err := usecase.repository.Delete(ctx, groupID); err != nil {
		return fmt.Errorf("delete group after tuple cleanup: %w", err)
	}
	return nil
}
