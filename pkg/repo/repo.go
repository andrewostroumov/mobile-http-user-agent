package repo

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/logrusadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

type Config struct {
	MinConn uint
	MaxConn uint
	URL     string
}

type Repo struct {
	ctx           context.Context
	contextLogger *log.Entry
	pool          *pgxpool.Pool
}

func NewRepo(ctx context.Context, logger *log.Logger, config Config) *Repo {
	r := &Repo{ctx: ctx}
	r.contextLogger = logger.WithFields(log.Fields{
		"package": "repo",
		"db":      "pg",
	})

	r.connect(config)
	r.poolMetrics()

	go func() {
		<-ctx.Done()
		if r.pool != nil {
			r.pool.Close()
		}
	}()

	return r
}

func (r *Repo) Stat() *pgxpool.Stat {
	return r.pool.Stat()
}

func (r *Repo) connect(config Config) {
	c, err := pgxpool.ParseConfig(config.URL)
	if err != nil {
		r.contextLogger.Error("Parse DB URL error: ", err)
		os.Exit(1)
	}

	c.ConnConfig.Logger = logrusadapter.NewLogger(r.contextLogger)
	c.ConnConfig.LogLevel = pgx.LogLevelDebug
	c.ConnConfig.PreferSimpleProtocol = true
	c.MinConns = int32(config.MinConn)
	c.MaxConns = int32(config.MaxConn)
	c.MaxConnLifetime = 3 * time.Minute
	c.MaxConnIdleTime = time.Minute

	pool, err := pgxpool.ConnectConfig(r.ctx, c)
	if err != nil {
		r.contextLogger.Error("Connect to pg error: ", err)
		os.Exit(1)
	}

	r.pool = pool
}

func (r *Repo) poolMetrics() {
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "coap",
		Subsystem: "database",
		Name:      "max_connections",
		Help:      "Database max connections",
	}, func() float64 {
		return float64(r.pool.Stat().MaxConns())
	})

	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "coap",
		Subsystem: "database",
		Name:      "total_connections",
		Help:      "Database total connections",
	}, func() float64 {
		return float64(r.pool.Stat().TotalConns())
	})

	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "coap",
		Subsystem: "database",
		Name:      "idle_connections",
		Help:      "Database idle connections",
	}, func() float64 {
		return float64(r.pool.Stat().IdleConns())
	})
}
