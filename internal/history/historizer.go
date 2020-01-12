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
