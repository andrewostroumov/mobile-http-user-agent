package http

import (
	"context"
	"github.com/andrewostroumov/mobile-http-user-agent/pkg/repo"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"sync"
	"time"
)

type Config struct {
	Addr string
}

type Server struct {
	ctx           context.Context
	repo          *repo.Repo
	config        Config
	inner         *http.Server
	mux           *mux.Router
	logger        *log.Logger
	contextLogger *log.Entry
}

func NewServer(ctx context.Context, logger *log.Logger, repo *repo.Repo, config Config) *Server {
	s := &Server{
		ctx:           ctx,
		repo:          repo,
		config:        config,
		logger:        logger,
		contextLogger: logger.WithFields(log.Fields{"package": "http"}),
	}

	s.newMux()
	s.newInner()

	return s
}

func (s *Server) newMux() {
	m := mux.NewRouter()
	m.Path("/heartbeat").
		Methods("GET").
		HandlerFunc(s.heartbeat)

	m.Path("/metrics").
		Methods("GET").
		Handler(promhttp.Handler())

	m.Path("/agents").
		Methods("GET").
		HandlerFunc(s.userAgent)

	s.mux = m
}

func (s *Server) newInner() {
	inner := &http.Server{
		Addr:         s.config.Addr,
		Handler:      s.mux,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		IdleTimeout:  3 * time.Second,
	}

	s.inner = inner
}

func (s *Server) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	l, err := s.listen()

	if err != nil {
		s.contextLogger.Fatalf("listen http error: %v", err)
	}

	s.contextLogger.Infof("listening http on %s", s.inner.Addr)

	go func() {
		<-s.ctx.Done()
		s.inner.Shutdown(s.ctx)
	}()

	err = s.inner.Serve(l)

	if err != nil && err != http.ErrServerClosed {
		s.contextLogger.Errorf("http error: %v", err)
	}
}

func (s *Server) listen() (net.Listener, error) {
	return net.Listen("tcp", s.inner.Addr)
}
