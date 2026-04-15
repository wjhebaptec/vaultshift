package quorum_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultshift/internal/provider/mock"
	"github.com/vaultshift/internal/quorum"
)

func TestNew_InvalidMinAcks(t *testing.T) {
	p := mock.New("p1")
	_, err := quorum.New(0, p)
	if err == nil {
		t.Fatal("expected error for minAcks=0")
	}
}

func TestNew_MinAcksExceedsProviders(t *testing.T) {
	p := mock.New("p1")
	_, err := quorum.New(2, p)
	if err == nil {
		t.Fatal("expected error when minAcks > provider count")
	}
}

func TestNew_NoProviders(t *testing.T) {
	_, err := quorum.New(1)
	if err == nil {
		t.Fatal("expected error for empty provider list")
	}
}

func TestPut_AllSucceed(t *testing.T) {
	p1 := mock.New("p1")
	p2 := mock.New("p2")
	q, err := quorum.New(2, p1, p2)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	failures, err := q.Put(context.Background(), "", "mykey", "myval")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(failures) != 0 {
		t.Fatalf("expected no failures, got %d", len(failures))
	}
}

func TestPut_QuorumMet_WithOneFailure(t *testing.T) {
	p1 := mock.New("p1")
	p2 := mock.New("p2")
	p2.ForceError(errors.New("write failed"))

	q, err := quorum.New(1, p1, p2)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	failures, err := q.Put(context.Background(), "", "k", "v")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(failures) != 1 {
		t.Fatalf("expected 1 failure, got %d", len(failures))
	}
}

func TestPut_QuorumNotMet(t *testing.T) {
	p1 := mock.New("p1")
	p2 := mock.New("p2")
	p1.ForceError(errors.New("err1"))
	p2.ForceError(errors.New("err2"))

	q, err := quorum.New(1, p1, p2)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = q.Put(context.Background(), "", "k", "v")
	if !errors.Is(err, quorum.ErrQuorumNotMet) {
		t.Fatalf("expected ErrQuorumNotMet, got %v", err)
	}
}

func TestPut_FiltersByProviderName(t *testing.T) {
	p1 := mock.New("p1")
	p2 := mock.New("p2")
	q, _ := quorum.New(1, p1, p2)

	// Only write to p1; p2 is skipped, acks=1 which meets minAcks=1
	_, err := q.Put(context.Background(), "p1", "k", "v")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMinAcks_AndProviderCount(t *testing.T) {
	p1 := mock.New("p1")
	p2 := mock.New("p2")
	p3 := mock.New("p3")
	q, _ := quorum.New(2, p1, p2, p3)
	if q.MinAcks() != 2 {
		t.Fatalf("expected MinAcks=2, got %d", q.MinAcks())
	}
	if q.ProviderCount() != 3 {
		t.Fatalf("expected ProviderCount=3, got %d", q.ProviderCount())
	}
}
