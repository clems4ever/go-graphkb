package client

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/schema"
	"github.com/sirupsen/logrus"
)

// Transaction represent a transaction generating updates by diffing the provided graph against
// the previous version.
type Transaction struct {
	client *GraphClient

	// The graph being updated
	graph  *knowledge.Graph
	binder *knowledge.GraphBinder

	// Lock used when binding or relating assets
	mutex sync.Mutex

	// Number of parallel queries to the Graph API
	parallelization int

	// The number of items to send to the streaming API in one request
	chunkSize int

	retryCount         int
	retryDelay         time.Duration
	retryBackoffFactor float64

	err error

	onSuccess func(*knowledge.Graph)
	onError   func(error)
}

// Relate create a relation between two assets
func (cgt *Transaction) Relate(from string, relationType schema.RelationType, to string) {
	cgt.mutex.Lock()
	err := cgt.binder.Relate(from, relationType, to)
	if err != nil && cgt.err == nil {
		cgt.err = fmt.Errorf("tx: %w", err)
	}
	cgt.mutex.Unlock()
}

// Bind bind one asset to an asset type from the schema
func (cgt *Transaction) Bind(asset string, assetType schema.AssetType) {
	cgt.mutex.Lock()
	err := cgt.binder.Bind(asset, assetType)
	if err != nil && cgt.err == nil {
		cgt.err = fmt.Errorf("tx: %w", err)
	}
	cgt.mutex.Unlock()
}

// withRetryOnTooManyRequests helper retrying the function when too many request error has been received
func withRetryOnTooManyRequests(fn func() error, backoffFactor float64, maxRetries int, delay time.Duration) error {
	trials := 0
	for {
		err := fn()
		if err != nil {
			logrus.Error(err)
			backoffTime := time.Duration(int(math.Pow(backoffFactor, float64(trials)))) * delay
			logrus.Infof("Sleeping for %s", backoffTime)
			time.Sleep(backoffTime)
		} else {
			return nil
		}
		trials++
		if trials == maxRetries {
			return fmt.Errorf("Too many retries... Aborting: %v", err)
		}
	}
}

// Commit commit the transaction and gives ownership to the source for caching.
func (cgt *Transaction) Commit() error {
	if cgt.err != nil {
		return fmt.Errorf("tx: commit: %w", cgt.err)
	}

	sg := cgt.graph.ExtractSchema()

	logrus.Debug("Start uploading the schema of the graph...")
	if err := cgt.client.UpdateSchema(sg); err != nil {
		err := fmt.Errorf("Unable to update the schema of the graph: %v", err)
		cgt.onError(err)
		return err
	}

	logrus.Debug("Finished uploading the schema of the graph...")

	logrus.Debug("Start uploading the graph...")
	now := time.Now()

	totalCount := 0

	count, err := chunkedTransfer(
		cgt.parallelization,
		cgt.chunkSize,
		cgt.graph.Assets(),
		knowledge.GraphEntryAdd,
		cgt.client.InsertAssets,
	)
	if err != nil {
		return err
	}
	logrus.Debugf("Inserted %d new assets", count)
	totalCount += count

	count, err = chunkedTransfer(
		cgt.parallelization,
		cgt.chunkSize,
		cgt.graph.Relations(),
		knowledge.GraphEntryAdd,
		cgt.client.InsertRelations,
	)
	if err != nil {
		return err
	}
	logrus.Debugf("Inserted %d new relations", count)
	totalCount += count

	count, err = chunkedTransfer(
		cgt.parallelization,
		cgt.chunkSize,
		cgt.graph.Relations(),
		knowledge.GraphEntryRemove,
		cgt.client.DeleteRelations,
	)
	if err != nil {
		return err
	}
	logrus.Debugf("Deleted %d old relations", count)
	totalCount += count

	count, err = chunkedTransfer(
		cgt.parallelization,
		cgt.chunkSize,
		cgt.graph.Assets(),
		knowledge.GraphEntryRemove,
		cgt.client.DeleteAssets,
	)
	if err != nil {
		return err
	}
	logrus.Debugf("Deleted %d old assets", count)
	totalCount += count

	if totalCount == 0 {
		// if there were no operations to perform, make an empty
		// call so the server knows that we ran
		err = cgt.client.InsertAssets(nil)
		if err != nil {
			return err
		}
	}

	elapsed := time.Since(now)
	logrus.Debugf("Finished uploading the graph (%d operations) in %s...", totalCount, elapsed)

	cgt.onSuccess(cgt.graph)
	cgt.graph = knowledge.NewGraph()
	return nil
}

func chunkedTransfer[T comparable](
	parallelization,
	chunkSize int,
	in map[T]knowledge.GraphEntryAction,
	actionMatch knowledge.GraphEntryAction,
	do func([]T) error,
) (int, error) {
	tasks := make(chan func() error)
	stop := make(chan struct{})

	var err error
	var wg sync.WaitGroup
	var stopOnce sync.Once

	wg.Add(parallelization)
	for i := 0; i < parallelization; i++ {
		go func() {
			defer wg.Done()

			for task := range tasks {
				if e := task(); e != nil {
					stopOnce.Do(func() {
						err = e
						close(stop)
					})
				}
			}
		}()
	}

	count := 0
	chunk := make([]T, 0, chunkSize)
	for el, action := range in {
		select {
		case <-stop:
			fmt.Println("stop")
			goto Done
		default:
		}

		if action == actionMatch {
			chunk = append(chunk, el)
			count++
		}
		if len(chunk) == chunkSize {
			func(chunk []T) {
				tasks <- func() error {
					return do(chunk)
				}
			}(chunk)
			chunk = make([]T, 0, chunkSize)
		}
	}
	if len(chunk) > 0 {
		tasks <- func() error {
			return do(chunk)
		}
	}

Done:
	close(tasks)

	wg.Wait()
	return count, err
}
