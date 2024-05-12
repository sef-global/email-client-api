package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	// "github.com/joho/godotenv"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mayura-andrew/email-client/internal/data"
	"github.com/mayura-andrew/email-client/internal/jsonlog"
	"github.com/mayura-andrew/email-client/internal/mailer"
	"github.com/mayura-andrew/email-client/internal/vcs"
)

var (
	version = vcs.Version()
)

type config struct {

	// API SERVER - configurations
	port int
	env  string

	// SMTP - configurations
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}

	cors struct {
		trustedOrigns []string
	}

	db struct {
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
}

type application struct {
	config config
	mailer mailer.Mailer
	logger *jsonlog.Logger
	models data.Models
}

func main() {
	var cfg config

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading environment variables file")
	}

	flag.IntVar(&cfg.port, "port", 4000, "Email API Server Port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|statging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("EMAILAPI"),
		"PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-times", "15m", "PostgreSQL max connection idle time")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	envVarValue := os.Getenv("SMTPPORT")

	if envVarValue == "" {
		fmt.Println("Environment variable is not set")
		return
	}

	intValue, err := strconv.Atoi(envVarValue)
	if err != nil {
		fmt.Println("Error conversting environment variable to integer:", err)
		return
	}

	smtpSender, err := url.QueryUnescape(os.Getenv("SMTPSENDER"))
	if err != nil {
		log.Fatalf("Failed to decore the SMTPSENDER : %v", err)
	}

	flag.StringVar(&cfg.smtp.host, "SMTPHOST", os.Getenv("SMTPHOST"), "SMTP host")
	flag.IntVar(&cfg.smtp.port, "SMTPPORT", intValue, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "SMTPUSERNAME", os.Getenv("SMTPUSERNAME"), "SMTP username")
	flag.StringVar(&cfg.smtp.password, "SMTPPASS",
	 os.Getenv("SMTPPASS"), "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "SMTPSENDER", smtpSender, "SMTP sender")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigns = strings.Fields(val)
		return nil
	})

	displayVersion := flag.Bool("version", false, "Display version and exit")
	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version: \t%s\n", version)
		os.Exit(0)
	}

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	expvar.NewString("version").Set(version)

	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	db, err := openDB(cfg)

	if err != nil {
		logger.PrintFatal(err, map[string]string{})
	}
	defer db.Close()

	logger.PrintInfo("database connection pool established", map[string]string{})

	app := &application{
		config: cfg,
		logger: logger,
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
		models: data.NewModel(db),
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}

}

func openDB(cfg config) (*sql.DB, error) {

	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
