package domain

import "errors"

type ApprovalStatus string

const (
	ApprovalPending     ApprovalStatus = "PENDING"
	ApprovalApproved    ApprovalStatus = "APPROVED"
	ApprovalProvisioned ApprovalStatus = "PROVISIONED"
	ApprovalFailed      ApprovalStatus = "FAILED"
)

var ErrInvalidApprovalTransition = errors.New("invalid approval transition")

type Approval struct {
	ID         string
	Status     ApprovalStatus
	ApprovedBy string
}

func (approval *Approval) Approve(actor string) error {
	if approval.Status != ApprovalPending {
		return ErrInvalidApprovalTransition
	}
	approval.Status = ApprovalApproved
	approval.ApprovedBy = actor
	return nil
}

func (approval *Approval) MarkProvisioned() error {
	if approval.Status != ApprovalApproved {
		return ErrInvalidApprovalTransition
	}
	approval.Status = ApprovalProvisioned
	return nil
}

func (approval *Approval) MarkFailed() error {
	if approval.Status != ApprovalApproved {
		return ErrInvalidApprovalTransition
	}
	approval.Status = ApprovalFailed
	return nil
}
