package domain

import "testing"

func TestApprovalTransitions(t *testing.T) {
	approval := &Approval{ID: "a-1", Status: ApprovalPending}
	if err := approval.Approve("engineer-1"); err != nil {
		t.Fatal(err)
	}
	if err := approval.MarkProvisioned(); err != nil {
		t.Fatal(err)
	}
	if approval.Status != ApprovalProvisioned {
		t.Fatalf("status = %s", approval.Status)
	}
}

func TestApprovalRejectsRepeatedApproval(t *testing.T) {
	approval := &Approval{ID: "a-1", Status: ApprovalApproved}
	if err := approval.Approve("engineer-1"); err != ErrInvalidApprovalTransition {
		t.Fatalf("error = %v", err)
	}
}
