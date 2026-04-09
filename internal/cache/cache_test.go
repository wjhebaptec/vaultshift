package cache_test

import (
	"testing"
	"time"

	"github.com/vaultshift/internal/cache"
)

func TestSet_AndGet(t *testing.T) {
	c := cache.New(5 * time.Minute)
	c.Set("db/password", "s3cr3t")

	val, ok := c.Get("db/password")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if val != "s3cr3t" {
		t.Fatalf("expected s3cr3t, got %s", val)
	}
}

func TestGet_NonExistent(t *testing.T) {
	c := cache.New(time.Minute)
	_, ok := c.Get("missing/key")
	if ok {
		t.Fatal("expected key to be absent")
	}
}

func TestGet_ExpiredEntry(t *testing.T) {
	c := cache.New(50 * time.Millisecond)
	c.Set("api/key", "abc123")

	time.Sleep(100 * time.Millisecond)

	_, ok := c.Get("api/key")
	if ok {
		t.Fatal("expected expired entry to be absent")
	}
}

func TestGet_ZeroTTL_NeverExpires(t *testing.T) {
	c := cache.New(0)
	c.Set("perm/key", "forever")

	time.Sleep(10 * time.Millisecond)

	val, ok := c.Get("perm/key")
	if !ok {
		t.Fatal("expected zero-TTL entry to persist")
	}
	if val != "forever" {
		t.Fatalf("expected forever, got %s", val)
	}
}

func TestDelete_RemovesKey(t *testing.T) {
	c := cache.New(time.Minute)
	c.Set("to/delete", "value")
	c.Delete("to/delete")

	_, ok := c.Get("to/delete")
	if ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestFlush_ClearsAll(t *testing.T) {
	c := cache.New(time.Minute)
	c.Set("key1", "v1")
	c.Set("key2", "v2")
	c.Flush()

	if c.Size() != 0 {
		t.Fatalf("expected size 0 after flush, got %d", c.Size())
	}
}

func TestSize_ReturnsCount(t *testing.T) {
	c := cache.New(time.Minute)
	if c.Size() != 0 {
		t.Fatal("expected empty cache size to be 0")
	}
	c.Set("a", "1")
	c.Set("b", "2")
	if c.Size() != 2 {
		t.Fatalf("expected size 2, got %d", c.Size())
	}
}
