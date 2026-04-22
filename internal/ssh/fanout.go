package ssh

import (
	"context"
	"sync"
)

// ServerResult holds the return value (or error) from running a function
// against a single server.
type ServerResult[T any] struct {
	Host   string
	Result T
	Err    error
}

// ForEachServer runs fn concurrently on every host, honouring the
// concurrency limit.  It always returns a result for every host (even on
// error) and never returns early — callers inspect each ServerResult.Err.
func ForEachServer[T any](
	ctx context.Context,
	hosts []string,
	concurrency int,
	fn func(ctx context.Context, host string) (T, error),
) []ServerResult[T] {
	if concurrency <= 0 {
		concurrency = 1
	}

	results := make([]ServerResult[T], len(hosts))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, host := range hosts {
		wg.Add(1)
		go func(idx int, h string) {
			defer wg.Done()

			// Acquire semaphore slot.
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				var zero T
				results[idx] = ServerResult[T]{Host: h, Result: zero, Err: ctx.Err()}
				return
			}

			res, err := fn(ctx, h)
			results[idx] = ServerResult[T]{Host: h, Result: res, Err: err}
		}(i, host)
	}

	wg.Wait()
	return results
}
