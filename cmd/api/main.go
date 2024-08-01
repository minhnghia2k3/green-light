package main

import (
	"context"
	"database/sql"
	"flag"
	_ "github.com/lib/pq"
	"github.com/minhnghia2k3/greenlight/internal/data"
	"github.com/minhnghia2k3/greenlight/internal/jsonlog"
	"github.com/minhnghia2k3/greenlight/internal/mailer"
	"os"
	"strings"
	"sync"
	"time"
)

// Application version number
const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

// Application struct hold the HTTP handlers, helpers, and middleware
type application struct {
	config config
	logger *jsonlog.Logger
	models *data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	// Read command-line flags
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://root:secret@localhost:5432/greenlight?sslmode=disable", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "82320722e3ba07", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "871db0730a25b9", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@greenlight.minhnghia2k3.net>", "SMTP sender")

	flag.Func("cors-trusted-origins", "Trusted CORS origin (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	flag.Parse()

	// creates a new Logger which writes to the std out stream
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Create a connection pool
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	// Declare an instance of application struct
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	logger.PrintInfo("database connection pool established", nil)
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

// openDB function returns a sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	pool, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	pool.SetMaxOpenConns(cfg.db.maxOpenConns)
	pool.SetMaxIdleConns(cfg.db.maxIdleConns)
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	pool.SetConnMaxIdleTime(duration)

	// Create a context with a 5-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = pool.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
