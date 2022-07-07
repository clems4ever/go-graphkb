package kbcontext

import (
	"context"
)

var (
	ContextKeyXForwardedUser = contextKey("xForwardedUser")
)

type contextKey string

func (c contextKey) String() string {
	return "server" + string(c)
}

// XForwardedUser gets the service user from context
func XForwardedUser(ctx context.Context) (string, bool) {
	xForwardedUser, ok := ctx.Value(ContextKeyXForwardedUser).(string)
	return xForwardedUser, ok
}
