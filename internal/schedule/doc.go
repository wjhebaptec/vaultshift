// Package schedule provides a lightweight job scheduler for vaultshift.
//
// It allows callers to register named rotation tasks with a fixed interval.
// Jobs are started in background goroutines and can be stopped gracefully
// via Stop or by cancelling the provided context.
//
// Example usage:
//
//	s := schedule.New()
//	_ = s.Register(schedule.Job{
//		Name:     "rotate-api-key",
//		Interval: 24 * time.Hour,
//		Task: func(ctx context.Context) error {
//			return rotator.RotateAll(ctx)
//		},
//	})
//	s.Start(ctx)
//	defer s.Stop()
package schedule
