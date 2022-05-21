package httpsrv

import (
	"context"

	"github.com/stn81/kate"
)

const (
	StatusOK   = "success"
	StatusFail = "fail"
)

type HealthStatus struct {
	Status    string            `json:"status"`
	SubStatus map[string]string `json:"substatus"`
}

type HealthCheckHandler struct {
	BaseHandler
}

func (h *HealthCheckHandler) ServeHTTP(ctx context.Context, w kate.ResponseWriter, r *kate.Request) {
	status := &HealthStatus{
		Status: StatusOK,
	}
	WriteJSON(w, status)
}
