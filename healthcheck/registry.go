package healthcheck

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/Falokut/go-kit/json"
	"github.com/Falokut/go-kit/log"
)

type HealthcheckFunc func(context.Context) error

func (f HealthcheckFunc) Healthcheck(ctx context.Context) error {
	return f(ctx)
}

type Checker interface {
	Healthcheck(ctx context.Context) error
}

type Registry struct {
	logger  log.Logger
	toCheck map[string]Checker
	mu      *sync.RWMutex
}

func NewRegistry(logger log.Logger) *Registry {
	return &Registry{
		logger:  logger,
		toCheck: make(map[string]Checker),
		mu:      &sync.RWMutex{},
	}
}

func (m *Registry) Register(name string, toCheck Checker) {
	m.mu.Lock()
	m.toCheck[name] = toCheck
	m.mu.Unlock()
}

func (r *Registry) Handler() http.Handler {
	return http.TimeoutHandler(http.HandlerFunc(r.handle), 1*time.Second, "timeout")
}

func (r *Registry) handle(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ctx := log.ToContext(req.Context(), log.Any("action", "healthcheck"))
	result := Result{
		Status:      StatusPass,
		FailDetails: make(map[string]any),
	}
	for resourceName, checker := range r.toCheck {
		err := checker.Healthcheck(ctx)
		if err == nil {
			continue
		}
		result.FailDetails[resourceName] = err
		result.Status = StatusFail
	}

	w.Header().Set("Content-Type", "application/health+json")
	if len(result.FailDetails) > 0 {
		r.logger.Error(ctx, "healthcheck error", log.Any("failDetails", result.FailDetails))
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_ = json.NewEncoder(w).Encode(result)
}
