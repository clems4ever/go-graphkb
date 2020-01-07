package knowledge

import (
	"context"
	"fmt"
	"time"

	"github.com/clems4ever/go-graphkb/internal/query"
)

type Querier struct {
	GraphDB GraphDB
}

type QuerierResult struct {
	Cursor      Cursor
	Projections []Projection
	Statistics  Statistics
}

func NewQuerier(db GraphDB) *Querier {
	return &Querier{GraphDB: db}
}

func (q *Querier) Query(ctx context.Context, cypherQuery string) (*QuerierResult, error) {
	s := Statistics{}

	var err error
	var queryIL *query.QueryIL

	s.Parsing = MeasureDuration(func() {
		queryIL, err = query.TransformCypher(cypherQuery)
	})

	if err != nil {
		return nil, err
	}

	var res *GraphQueryResult
	s.Execution = MeasureDuration(func() {
		res, err = q.GraphDB.Query(ctx, queryIL)
	})
	fmt.Printf("Found results in %dms\n", s.Execution/time.Millisecond)

	if err != nil {
		return nil, err
	}

	result := &QuerierResult{
		Cursor:      res.Cursor,
		Projections: res.Projections,
		Statistics:  s,
	}
	return result, nil
}

type Statistics struct {
	Parsing   time.Duration
	Execution time.Duration
}

func MeasureDuration(Func func()) time.Duration {
	now := time.Now()
	Func()
	return time.Since(now)
}
