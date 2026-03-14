package parallel

import (
	"slices"
	"sync"

	"github.com/holyheld/gointernals/typeutil"
)

// ExecutePool is a convenience helper that distributes input tasks into
// workerCount workers then collects the result.
//
// Output values are not expected to have the same order as they came in.
//
// Panics when provided executor panics.
func ExecutePool[I any, O any](
	tasks []I,
	f func(task I) O,
	workerCount int,
) []O {
	return executePool(typeutil.ChanSlice(tasks), f, workerCount)
}

func executePool[I any, O any](
	tasks <-chan I,
	f func(task I) O,
	workerCount int,
) []O {
	results := make(chan O, min(workerCount, len(tasks)))

	wg := sync.WaitGroup{}
	for range workerCount {
		wg.Go(func() {
			for task := range tasks {
				result := f(task)
				results <- result
			}
		})
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	resultData := make([]O, 0, len(tasks))
	for result := range results {
		resultData = append(resultData, result)
	}

	return resultData
}

// ExecutePool2 is a convenience helper that distributes input tasks into
// workerCount runners then collects the result.
//
// Output values are not expected to have the same order as they came in.
//
// Fails when either of workers fails with nil res and error, otherwise error is nil
// and the result is OK.
func ExecutePool2[I any, O any](
	tasks []I,
	f func(task I) (O, error),
	workerCount int,
) ([]O, error) {
	return executePool2(typeutil.ChanSlice(tasks), f, workerCount, len(tasks))
}

func executePool2[I any, O any](
	tasks <-chan I,
	f func(task I) (O, error),
	workerCount int,
	length int,
) ([]O, error) {
	done := make(chan struct{})
	defer close(done)

	type resultHolder struct {
		Data  O
		Error error
	}

	results := make(chan resultHolder, length)

	var wg sync.WaitGroup
	for range workerCount {
		wg.Go(func() {
			for in := range tasks {
				select {
				case <-done:
					return
				default:
					result, err := f(in)
					results <- resultHolder{
						Data:  result,
						Error: err,
					}
				}
			}
		})
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	resultData := make([]O, 0, length)

	for result := range results {
		if result.Error != nil {
			return nil, result.Error
		}

		resultData = append(resultData, result.Data)
	}

	return resultData, nil
}

// ExecuteChunkSync executes provided function over tasks in chunks synchronously.
func ExecuteChunkSync[I any, O any](
	tasks []I,
	chunkSize int,
	f func(chunk []I) ([]O, error),
) ([]O, error) {
	res := make([]O, 0, typeutil.DivUp(len(tasks), chunkSize))

	for chunk := range slices.Chunk(tasks, chunkSize) {
		batchRes, err := f(chunk)
		if err != nil {
			return nil, err
		}

		res = append(res, batchRes...)
	}

	return res, nil
}

// ExecuteChunkAsync executes provided data in chunks with pool workers.
func ExecuteChunkAsync[I any, O any](
	tasks []I,
	chunkSize int,
	f func(chunk []I) ([]O, error),
	workerCount int,
) ([]O, error) {
	chunks := typeutil.ChanSeqSized(
		slices.Chunk(tasks, chunkSize),
		typeutil.DivUp(len(tasks), chunkSize),
	)

	resChunk, err := executePool2(chunks, f, workerCount, len(tasks))
	if err != nil {
		return nil, err
	}

	res := make([]O, 0, len(tasks))
	for _, chunk := range resChunk {
		res = append(res, chunk...)
	}

	return res, nil
}
