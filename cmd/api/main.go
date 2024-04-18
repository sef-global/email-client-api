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

	"github.com/joho/godotenv"
	"github.com/mayura-andrew/email-client/internal/jsonlog"
	"github.com/mayura-andrew/email-client/internal/mailer"
	"github.com/mayura-andrew/email-client/internal/vcs"
	_ "github.com/lib/pq"
)

var (version1 = vcs.Version())

const version = "1.0.0"

type config struct {

	// API SERVER - configurations
	port int
	env string

	// SMTP - configurations
	smtp struct {
		host string
		port int
		username string
		password string
		sender string
	}

	cors struct {
		trustedOrigns []string
	}

	db struct {
		dsn string
	}
}

type application struct {
	config config
	mailer mailer.Mailer
	logger *jsonlog.Logger
}
func main() {
	fmt.Println("Hello world")

	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "Email API Server Port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|statging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://andrew:OslpJueINuYbKRmc7UvNRjyqZ7bV1Byq@dpg-cogkfq21hbls738s5lm0-a.singapore-postgres.render.com/emailbulk", "PostgreSQL DSN")

	err := godotenv.Load(".env")
    if err != nil {
        log.Fatalf("Error loading environment variables file")
    }

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

	fmt.Printf("%d", intValue)

	smtpSender, err := url.QueryUnescape(os.Getenv("SMTPSENDER"))
	if err != nil {
		log.Fatalf("Failed to decore the SMTPSENDER : %v", err)
	}

	fmt.Printf("%s", smtpSender)
	flag.StringVar(&cfg.smtp.host, "smtp-host", os.Getenv("SMTPHOST"), "SMTP host")
	flag.IntVar(&cfg.smtp.port, "SMTPPORT", intValue, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", os.Getenv("SMTPUSERNAME"), "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("SMTPPASS"), "SMTP password")
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	err = db.PingContext(ctx)

	if err != nil {
		return nil, err
	}

	return db, nil

}