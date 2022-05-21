package httpsrv

import (
	"context"
	"net/http"

	"github.com/stn81/kate"
)

type PingHandler struct {
	BaseHandler
}

func (h *PingHandler) ServeHTTP(ctx context.Context, w kate.ResponseWriter, r *kate.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	WritePlain(w, []byte("OK"))
}
