package httpsrv

import (
	"context"
	"net/http"

	"github.com/stn81/kate"
	"github.com/stn81/kate/log/ctxzap"
	"go.uber.org/zap"
)

// Recovery implements the recovery wrapper middleware
func Recovery(h kate.ContextHandler) kate.ContextHandler {
	f := func(ctx context.Context, w kate.ResponseWriter, r *kate.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				// nolint:errcheck
				w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
				ctxzap.Extract(ctx).Error("got panic", zap.Any("error", err), zap.Stack("stack"))
			}
		}()

		h.ServeHTTP(ctx, w, r)
	}
	return kate.ContextHandlerFunc(f)
}
