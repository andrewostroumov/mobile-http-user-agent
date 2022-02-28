package main

import (
	"context"
	"github.com/andrewostroumov/mobile-http-user-agent/internal/app/http"
	"github.com/andrewostroumov/mobile-http-user-agent/pkg/devices"
	"github.com/andrewostroumov/mobile-http-user-agent/pkg/repo"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	httpAddress       = kingpin.Flag("http.address", "HTTP address").Default(":8080").Envar("HTTP_ADDRESS").String()
	coreProd          = kingpin.Flag("core.production", "Run in production").Default("false").Envar("CORE_PRODUCTION").Bool()
	configAndroidPath = kingpin.Flag("config.devices.android.path", "Config devices android file path").Default("docs/android.json").Envar("CONFIG_DEVICES_ANDROID_PATH").String()
	configChromePath  = kingpin.Flag("config.devices.chrome.path", "Config devices chrome file path").Default("docs/chrome.json").Envar("CONFIG_DEVICES_CHROME_PATH").String()
	revVerPath        = kingpin.Flag("rev.ver-path", "Version file path").Default(".version").Envar("REV_VERSION_PATH").String()
	revRevPath        = kingpin.Flag("rev.rev-path", "Revision file path").Default(".revision").Envar("REV_REVISION_PATH").String()
	dbMinConn         = kingpin.Flag("db.min-conn", "Database min connections pool").Default("1").Envar("DATABASE_POOL_MIN_CONNECTIONS").Uint()
	dbMaxConn         = kingpin.Flag("db.max-conn", "Database max connections pool").Default("64").Envar("DATABASE_POOL_MAX_CONNECTIONS").Uint()
	dbUrl             = kingpin.Flag("db.url", "Database URL address").Default("postgresql://localhost:5432/postgres").Envar("DATABASE_URL").String()
)

var logger = log.New()
var contextLogger = logger.WithFields(log.Fields{"package": "main"})

func init() {
	kingpin.HelpFlag.Short('h')
	kingpin.Version("0.0.1")
	kingpin.Parse()

	logger.SetNoLock()

	if *coreProd {
		logger.SetFormatter(&log.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
			FieldMap: log.FieldMap{
				log.FieldKeyLevel: "severity",
			},
		})
		logger.SetLevel(log.TraceLevel)
	} else {
		logger.SetFormatter(&log.TextFormatter{})
		logger.SetLevel(log.TraceLevel)
	}
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	i := make(chan os.Signal)
	signal.Notify(i, syscall.SIGINT, syscall.SIGTERM)

	defer func() {
		signal.Stop(i)
		cancel()
	}()

	contextLogger.Info("Running Mobile UserAgent API")

	go func() {
		select {
		case sig := <-i:
			contextLogger.WithFields(log.Fields{
				"signal": sig,
			}).Info("Got signal. Aborting...")
			cancel()
		case <-ctx.Done():
		}
	}()

	devices.Parse(devices.Config{
		AndroidFile: *configAndroidPath,
		ChromeFile:  *configChromePath,
	}, logger)

	r := repo.NewRepo(ctx, logger, repo.Config{
		MinConn: *dbMinConn,
		MaxConn: *dbMaxConn,
		URL:     *dbUrl,
	})

	h := http.NewServer(ctx, logger, r, http.Config{
		Addr: *httpAddress,
	})

	var wg sync.WaitGroup
	wg.Add(1)

	go h.Run(&wg)

	wg.Wait()
}
