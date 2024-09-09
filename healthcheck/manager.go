package healthcheck

import (
	"context"
	"net/http"
	"sync"

	"github.com/Falokut/go-kit/json"
	"github.com/Falokut/go-kit/log"
)

type HealthcheckFunc func(context.Context) error

type Manager struct {
	logger     log.Logger
	toCheck    map[string]HealthcheckFunc
	listenPort string
	mu         *sync.RWMutex
}

func NewHealthManager(logger log.Logger, listenPort string) Manager {
	return Manager{
		logger:     logger,
		listenPort: listenPort,
		toCheck:    make(map[string]HealthcheckFunc),
		mu:         &sync.RWMutex{},
	}
}

func (m *Manager) RunHealthcheckEndpoint() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthcheck", m.healthcheckHandler)
	return http.ListenAndServe(":"+m.listenPort, mux)
}

func (m *Manager) Register(name string, toCheck HealthcheckFunc) {
	m.mu.Lock()
	m.toCheck[name] = toCheck
	m.mu.Unlock()
}

func (m *Manager) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.Unlock()

	ctx := log.ToContext(r.Context(), log.Any("action", "healthcheck"))
	result := Result{}
	hasError := false
	for resourceName, checkFn := range m.toCheck {
		err := checkFn(ctx)
		if err == nil {
			continue
		}
		result.FailDetails[resourceName] = err
		hasError = true
	}

	w.Header().Set("Content-Type", "application/health+json")
	if hasError {
		m.logger.Error(ctx, "healthcheck error", log.Any("failDetails", result.FailDetails))
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_ = json.NewEncoder(w).Encode(Result{Status: StatusPass})
}
