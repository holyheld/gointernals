package parallel

import (
	"sync"
)

// SyncDispatcher is a convenience helper that distributes input tasks into
// workerCount runners then collects the result.
//
// Output values are not expected to have the same order as they came in.
//
// Panics when provided executor panics.
func SyncDispatcher[I any, O any](
	inputData []I,
	f func(input I) O,
	workerCount int,
) []O {
	var wg sync.WaitGroup

	input := make(chan I, len(inputData))
	for _, in := range inputData {
		input <- in
	}
	close(input)

	results := make(chan O, min(workerCount, len(inputData)))
	// Start workers
	for range workerCount {
		wg.Go(func() {
			for in := range input {
				result := f(in)
				results <- result
			}
		})
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	resultData := make([]O, 0, len(inputData))
	for result := range results {
		resultData = append(resultData, result)
	}

	return resultData
}

// SyncDispatcher2 is a convenience helper that distributes input tasks into
// workerCount runners then collects the result.
//
// Output values are not expected to have the same order as they came in.
//
// Fails when either of workers fails with nil res and error, otherwise error is nil
// and the result is OK.
func SyncDispatcher2[I any, O any](
	inputData []I,
	f func(input I) (O, error),
	workerCount int,
) ([]O, error) {
	var wg sync.WaitGroup

	input := make(chan I, len(inputData))
	for _, in := range inputData {
		input <- in
	}
	close(input)

	done := make(chan struct{})
	defer close(done)

	results := make(chan O, min(workerCount, len(inputData)))
	errors := make(chan error, workerCount)
	// Start workers
	for range workerCount {
		wg.Go(func() {
			for in := range input {
				select {
				case <-done:
					return
				default:
					result, err := f(in)
					results <- result
					errors <- err
				}
			}
		})
	}

	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	for err := range errors {
		if err != nil {
			return nil, err
		}
	}

	resultData := make([]O, 0, len(inputData))
	for result := range results {
		resultData = append(resultData, result)
	}

	return resultData, nil
}
