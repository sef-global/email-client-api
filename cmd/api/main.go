package main

import (
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
	"github.com/mayura-andrew/email-client/internal/vcs"
)

var (version9 = vcs.Version())

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
}

type application struct {
	config config
	mailer Mailer
	logger *jsonlog.Logger
	
}
func main() {
	fmt.Println("Hello world")

	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "Email API Server Port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|statging|production)")

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
	flag.IntVar(&cfg.smtp.port, "smtp-port", intValue, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", os.Getenv("SMTPUSERNAME"), "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("SMTPPASS"), "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", smtpSender, "SMTP sender")


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


	recipients := []string{"recipient1@example.com", "recipient2@example.com", "mayuraahalakoon@gmail.com"}
	body := `"key":"value"`
	mailer, err := New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender, recipients, body)	
	if err != nil {
		log.Fatalf("Failed to create mailer: %v", err)
	}

	app := &application{
		config: cfg,
		logger: logger,
		mailer: *mailer,
	}

	app.serve()
	logger.PrintFatal(err, nil)
	// if err := app.serve(); err != nil {
    //     fmt.Println("Error starting server:", err)
    //     return
    // }
}