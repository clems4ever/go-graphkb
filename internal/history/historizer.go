package history

import (
	"context"
	"time"
)

type Status int

const (
	Success Status = iota
	Failure Status = iota
)

type Historizer interface {
	SaveSuccessfulQuery(ctx context.Context, cypher, sql string, duration time.Duration) error
	SaveFailedQuery(ctx context.Context, cypher, sql string, err error) error
}

type NoopHistorizer struct{}

func (h *NoopHistorizer) SaveSuccessfulQuery(ctx context.Context, cypher, sql string, duration time.Duration) error {
	return nil
}

func (h *NoopHistorizer) SaveFailedQuery(ctx context.Context, cypher, sql string, err error) error {
	return nil
}
