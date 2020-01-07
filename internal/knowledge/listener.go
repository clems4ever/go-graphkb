package knowledge

import "log"

// SourceSubGraphUpdates represents the updates to perform on a source subgraph
type SourceSubGraphUpdates struct {
	Updates GraphUpdatesBulk
	Source  string
}

// SourceListener represents the source listener waiting for source events
type SourceListener struct {
	db GraphDB
}

// NewSourceListener create a KB
func NewSourceListener(db GraphDB) *SourceListener {
	return &SourceListener{db}
}

// Listen events coming from the event bus
func (sl *SourceListener) Listen(updatesC chan SourceSubGraphUpdates) {
	for updates := range updatesC {
		if err := sl.db.UpdateGraph(updates.Source, &updates.Updates); err != nil {
			log.Fatal(err)
		}
	}
}
