package accesslog

import (
	"context"
	"errors"
	"testing"
)

type stubProvider struct {
	getVal string
	getErr error
	putErr error
	delErr error
	listVal []string
	listErr error
}

func (s *stubProvider) Get(_ context.Context, _ string) (string, error) {
	return s.getVal, s.getErr
}
func (s *stubProvider) Put(_ context.Context, _, _ string) error { return s.putErr }
func (s *stubProvider) Delete(_ context.Context, _ string) error { return s.delErr }
func (s *stubProvider) List(_ context.Context) ([]string, error) { return s.listVal, s.listErr }

func TestWrapGet_RecordsSuccess(t *testing.T) {
	l := New()
	stub := &stubProvider{getVal: "secret"}
	mw := Wrap("aws", stub, l)
	v, err := mw.Get(context.Background(), "db/pass")
	if err != nil || v != "secret" {
		t.Fatalf("unexpected result: %v %v", v, err)
	}
	entries := l.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if !entries[0].Success || entries[0].Operation != OpGet {
		t.Errorf("expected successful get entry")
	}
}

func TestWrapGet_RecordsError(t *testing.T) {
	l := New()
	stub := &stubProvider{getErr: errors.New("not found")}
	mw := Wrap("gcp", stub, l)
	_, err := mw.Get(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	e := l.Entries()[0]
	if e.Success || e.Error == "" {
		t.Error("expected failure entry with error message")
	}
}

func TestWrapPut_RecordsOperation(t *testing.T) {
	l := New()
	mw := Wrap("vault", &stubProvider{}, l)
	_ = mw.Put(context.Background(), "k", "v")
	if l.Entries()[0].Operation != OpPut {
		t.Error("expected put operation")
	}
}

func TestWrapDelete_RecordsOperation(t *testing.T) {
	l := New()
	mw := Wrap("aws", &stubProvider{}, l)
	_ = mw.Delete(context.Background(), "k")
	if l.Entries()[0].Operation != OpDelete {
		t.Error("expected delete operation")
	}
}

func TestWrapList_RecordsOperation(t *testing.T) {
	l := New()
	mw := Wrap("aws", &stubProvider{listVal: []string{"a", "b"}}, l)
	keys, err := mw.List(context.Background())
	if err != nil || len(keys) != 2 {
		t.Fatalf("unexpected list result: %v %v", keys, err)
	}
	if l.Entries()[0].Operation != OpList {
		t.Error("expected list operation")
	}
}

func TestWrap_RecordsProviderName(t *testing.T) {
	l := New()
	mw := Wrap("my-provider", &stubProvider{}, l)
	_ = mw.Put(context.Background(), "k", "v")
	if l.Entries()[0].Provider != "my-provider" {
		t.Error("expected provider name to be recorded")
	}
}
