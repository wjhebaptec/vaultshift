package backoff_test

import (
	"testing"
	"time"

	"github.com/vaultshift/internal/backoff"
)

func TestFixed_ConstantDelay(t *testing.T) {
	b := backoff.New(backoff.Config{
		Strategy:  backoff.StrategyFixed,
		BaseDelay: 200 * time.Millisecond,
	})
	for attempt := 1; attempt <= 5; attempt++ {
		d := b.Delay(attempt)
		if d != 200*time.Millisecond {
			t.Errorf("attempt %d: expected 200ms, got %v", attempt, d)
		}
	}
}

func TestLinear_GrowsWithAttempt(t *testing.T) {
	b := backoff.New(backoff.Config{
		Strategy:  backoff.StrategyLinear,
		BaseDelay: 100 * time.Millisecond,
	})
	for attempt := 1; attempt <= 4; attempt++ {
		d := b.Delay(attempt)
		want := time.Duration(attempt) * 100 * time.Millisecond
		if d != want {
			t.Errorf("attempt %d: expected %v, got %v", attempt, want, d)
		}
	}
}

func TestExponential_DoublesEachAttempt(t *testing.T) {
	b := backoff.New(backoff.Config{
		Strategy:   backoff.StrategyExponential,
		BaseDelay:  100 * time.Millisecond,
		Multiplier: 2.0,
	})
	expected := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
		800 * time.Millisecond,
	}
	for i, want := range expected {
		got := b.Delay(i + 1)
		if got != want {
			t.Errorf("attempt %d: expected %v, got %v", i+1, want, got)
		}
	}
}

func TestMaxDelay_Capped(t *testing.T) {
	b := backoff.New(backoff.Config{
		Strategy:   backoff.StrategyExponential,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   300 * time.Millisecond,
		Multiplier: 2.0,
	})
	for attempt := 3; attempt <= 6; attempt++ {
		d := b.Delay(attempt)
		if d > 300*time.Millisecond {
			t.Errorf("attempt %d: delay %v exceeds max", attempt, d)
		}
	}
}

func TestJitter_AddsVariance(t *testing.T) {
	b := backoff.New(backoff.Config{
		Strategy:  backoff.StrategyFixed,
		BaseDelay: 500 * time.Millisecond,
		Jitter:    true,
	})
	var prev time.Duration
	varied := false
	for i := 0; i < 20; i++ {
		d := b.Delay(1)
		if d < 500*time.Millisecond {
			t.Errorf("jitter should not reduce below base: got %v", d)
		}
		if prev != 0 && d != prev {
			varied = true
		}
		prev = d
	}
	if !varied {
		t.Error("expected jitter to produce varying delays")
	}
}

func TestDefaultConfig_IsExponential(t *testing.T) {
	cfg := backoff.DefaultConfig()
	if cfg.Strategy != backoff.StrategyExponential {
		t.Errorf("expected exponential strategy, got %v", cfg.Strategy)
	}
}

func TestDelay_ZeroAttempt_TreatedAsOne(t *testing.T) {
	b := backoff.New(backoff.Config{
		Strategy:  backoff.StrategyFixed,
		BaseDelay: 50 * time.Millisecond,
	})
	if d := b.Delay(0); d != 50*time.Millisecond {
		t.Errorf("expected 50ms for attempt 0, got %v", d)
	}
}
