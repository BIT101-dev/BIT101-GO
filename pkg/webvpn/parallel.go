package webvpn

import (
	"context"
	"sync"
)

const schoolParallelism = 4

func parallelMap[I, O any](parent context.Context, items []I, fn func(context.Context, I) (O, error)) ([]O, error) {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	type result struct {
		index int
		value O
		err   error
	}
	results := make(chan result, len(items))
	slots := make(chan struct{}, schoolParallelism)
	var workers sync.WaitGroup
	for index, item := range items {
		workers.Add(1)
		go func(index int, item I) {
			defer workers.Done()
			select {
			case slots <- struct{}{}:
			case <-ctx.Done():
				results <- result{index: index, err: ctx.Err()}
				return
			}
			defer func() { <-slots }()
			value, err := fn(ctx, item)
			results <- result{index: index, value: value, err: err}
		}(index, item)
	}
	go func() {
		workers.Wait()
		close(results)
	}()

	output := make([]O, len(items))
	var firstErr error
	for result := range results {
		if result.err != nil {
			if firstErr == nil {
				firstErr = result.err
				cancel()
			}
			continue
		}
		output[result.index] = result.value
	}
	return output, firstErr
}
