package knowledge

import (
	"context"
	"time"

	"github.com/clems4ever/go-graphkb/internal/history"
	"github.com/clems4ever/go-graphkb/internal/metrics"
	"github.com/clems4ever/go-graphkb/internal/query"
	"github.com/sirupsen/logrus"
)

type Querier struct {
	GraphDB    GraphDB
	historizer history.Historizer
}

type QuerierResult struct {
	Cursor      Cursor
	Projections []Projection
	Statistics  Statistics
}

// NewQuerier create an instance of a querier
func NewQuerier(db GraphDB, historizer history.Historizer) *Querier {
	return &Querier{GraphDB: db, historizer: historizer}
}

// Query run a query against the graph DB.
func (q *Querier) Query(ctx context.Context, queryString string) (*QuerierResult, error) {
	qr, _, err := q.queryInternal(ctx, queryString)
	if err != nil {
		return nil, err
	}
	return qr, nil
}

func (q *Querier) queryInternal(ctx context.Context, cypherQuery string) (*QuerierResult, string, error) {
	s := Statistics{}

	var err error
	var queryCypher *query.QueryCypher

	s.Parsing = MeasureDuration(func() {
		queryCypher, err = query.TransformCypher(cypherQuery)
	})

	if err != nil {
		return nil, "", err
	}

	translation, err := NewSQLQueryTranslator().Translate(queryCypher)
	if err != nil {
		return nil, "", err
	}

	var res *GraphQueryResult
	s.Execution = MeasureDuration(func() {
		res, err = q.GraphDB.Query(ctx, *translation)
	})

	if err != nil {
		return nil, translation.Query, err
	}

	executionTime := s.Execution.Milliseconds()

	logrus.Debugf("Found results in %s", s.Execution)

	metrics.GraphQueryTimeExecution.
		WithLabelValues().Observe(float64(executionTime))

	result := &QuerierResult{
		Cursor:      res.Cursor,
		Projections: res.Projections,
		Statistics:  s,
	}
	return result, translation.Query, nil
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
