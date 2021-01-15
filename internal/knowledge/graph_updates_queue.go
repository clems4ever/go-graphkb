package knowledge

import (
	"fmt"

	"github.com/golang-collections/go-datastructures/queue"
)

// GraphUpdatesQueue represent a graph updates queue with
type GraphUpdatesQueue struct {
	q *queue.Queue

	length int
}

// NewGraphUpdatesQueue create a new instance of updates queue
func NewGraphUpdatesQueue(length int) *GraphUpdatesQueue {
	return &GraphUpdatesQueue{
		q:      queue.New(int64(length)),
		length: length,
	}
}

// IsFull check if the queue is full
func (guq *GraphUpdatesQueue) IsFull() bool {
	return int(guq.q.Len()) == guq.length
}

// Enqueue an item into the queue
func (guq *GraphUpdatesQueue) Enqueue(update SourceSubGraphUpdates) error {
	return guq.q.Put(update)
}

// Dequeue an item from the queue
func (guq *GraphUpdatesQueue) Dequeue() (*SourceSubGraphUpdates, error) {
	it, err := guq.q.Get(1)
	if err != nil {
		return nil, fmt.Errorf("Unable to dequeue item: %v", err)
	}

	if len(it) != 1 {
		return nil, fmt.Errorf("There should be one item coming out of the queue: %v", err)
	}

	u := it[0].(SourceSubGraphUpdates)

	return &u, nil
}

// Dispose will dispose of this queue. Any subsequent calls to Get or Put will return an error.
func (guq *GraphUpdatesQueue) Dispose() {
	guq.q.Dispose()
}
