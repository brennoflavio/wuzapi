package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ServerMode represents the server operating mode
type ServerMode int

const (
	HTTP ServerMode = iota
	Stdio
)

type server struct {
	db     *sqlx.DB
	router *mux.Router
	exPath string
	mode   ServerMode
}

// Replace the global variables
var (
	address             = flag.String("address", "0.0.0.0", "Bind IP Address")
	port                = flag.String("port", "8080", "Listen Port")
	waDebug             = flag.String("wadebug", "", "Enable whatsmeow debug (INFO or DEBUG)")
	skipMedia           = flag.Bool("skipmedia", false, "Do not attempt to download media in messages")
	osName              = flag.String("osname", "Mac OS 10", "Connection OSName in Whatsapp")
	globalWebhook = flag.String("globalwebhook", "", "Global webhook URL to receive all events from all users")
	versionFlag         = flag.Bool("version", false, "Display version information and exit")
	mode                = flag.String("mode", "http", "Server mode: http or stdio")
	dataDir             = flag.String("datadir", "", "Data directory for database and session files (defaults to executable directory)")

	webhookRetryEnabled      = flag.Bool("webhookretry", true, "Enable webhook retry mechanism")
	webhookRetryCount        = flag.Int("retrycount", 5, "Number of times to retry failed webhooks")
	webhookRetryDelaySeconds = flag.Int("retrydelay", 30, "Delay in seconds between webhook retries")

	container        *sqlstore.Container
	clientManager    = NewClientManager()
	killchannel      = make(map[string](chan bool))
	userinfocache    = cache.New(5*time.Minute, 10*time.Minute)
	lastMessageCache = cache.New(24*time.Hour, 24*time.Hour)
	globalHTTPClient = newSafeHTTPClient()
)

var privateIPBlocks []*net.IPNet

const version = "1.0.5"

func newSafeHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, fmt.Errorf("unexpected address format from http transport: %q: %w", addr, err)
				}

				ips, err := net.LookupIP(host)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve host '%s': %w", host, err)
				}
				if len(ips) == 0 {
					return nil, fmt.Errorf("no IP addresses found for host: %s", host)
				}

				var (
					lastDialErr   error
					ssrfDetected  bool
					ssrfLastError error
				)

				for _, ip := range ips {
					if isPrivateOrLoopback(ip) {
						log.Warn().Str("ip", ip.String()).Str("host", host).Msg("SSRF attempt detected: refused to connect to private or local address")
						ssrfDetected = true
						if ssrfLastError == nil {
							ssrfLastError = fmt.Errorf("ssrf attempt detected: host '%s' resolves to one or more private IP addresses", host)
						}
						continue
					}

					dialer := &net.Dialer{
						Timeout:   4 * time.Second,
						KeepAlive: 30 * time.Second,
					}

					connAddr := net.JoinHostPort(ip.String(), port)
					conn, err := dialer.DialContext(ctx, network, connAddr)
					if err == nil {
						return conn, nil
					}
					lastDialErr = err
				}

				if lastDialErr != nil {
					return nil, lastDialErr
				}
				if ssrfDetected {
					return nil, ssrfLastError
				}
				if lastDialErr != nil {
					return nil, lastDialErr
				}
				return nil, fmt.Errorf("no dialable IP addresses found for host %s", host)
			},
		},
	}
}

func isPrivateOrLoopback(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func main() {
	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"100.64.0.0/10",  // RFC6598 Carrier-Grade NAT
		"169.254.0.0/16", // RFC3927 link-local
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to parse CIDR string: %s", cidr)
		}
		privateIPBlocks = append(privateIPBlocks, block)
	}

	err := godotenv.Load()
	if err != nil {
		log.Warn().Err(err).Msg("It was not possible to load the .env file (it may not exist).")
	}

	flag.Parse()

	// Check for address in environment variable if flag is default or empty
	if *address == "0.0.0.0" || *address == "" {
		if v := os.Getenv("WUZAPI_ADDRESS"); v != "" {
			*address = v
			log.Info().Str("address", v).Msg("Address configured from environment variable")
		}
	}

	// Check for port in environment variable if flag is default or empty
	if *port == "8080" || *port == "" {
		if v := os.Getenv("WUZAPI_PORT"); v != "" {
			*port = v
			log.Info().Str("port", v).Msg("Port configured from environment variable")
		}
	}

	if v := os.Getenv("WEBHOOK_RETRY_ENABLED"); v != "" {
		*webhookRetryEnabled = strings.ToLower(v) == "true" || v == "1"
	}
	if v := os.Getenv("WEBHOOK_RETRY_COUNT"); v != "" {
		if count, err := strconv.Atoi(v); err == nil {
			*webhookRetryCount = count
		}
	}
	if v := os.Getenv("WEBHOOK_RETRY_DELAY_SECONDS"); v != "" {
		if delay, err := strconv.Atoi(v); err == nil {
			*webhookRetryDelaySeconds = delay
		}
	}
	log.Info().
		Bool("enabled", *webhookRetryEnabled).
		Int("count", *webhookRetryCount).
		Int("delay", *webhookRetryDelaySeconds).
		Msg("Webhook Retry Configured")

	// Novo bloco para sobrescrever o osName pelo ENV, se existir
	if v := os.Getenv("SESSION_DEVICE_NAME"); v != "" {
		*osName = v
	}

	if *versionFlag {
		fmt.Printf("WuzAPI version %s\n", version)
		os.Exit(0)
	}

	// In stdio mode, always log to stderr to avoid interfering with JSON responses on stdout
	logOutput := os.Stdout
	if *mode == "stdio" {
		logOutput = os.Stderr
	}

	output := zerolog.ConsoleWriter{
		Out:        logOutput,
		TimeFormat: "2006-01-02 15:04:05 -07:00",
		NoColor:    true,
	}

	log.Logger = zerolog.New(output).
		With().
		Timestamp().
		Str("role", filepath.Base(os.Args[0])).
		Logger()

	// Setup timezone (after logger is configured)
	tz := os.Getenv("TZ")
	if tz != "" {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			log.Warn().Err(err).Msgf("It was not possible to define TZ=%q, using UTC", tz)
		} else {
			time.Local = loc
			log.Info().Str("TZ", tz).Msg("Timezone defined")
		}
	}

	// Check for global webhook in environment variable
	if *globalWebhook == "" {
		if v := os.Getenv("WUZAPI_GLOBAL_WEBHOOK"); v != "" {
			*globalWebhook = v
			log.Info().Str("global_webhook", v).Msg("Global webhook configured from environment variable")
		}
	} else {
		log.Info().Str("global_webhook", *globalWebhook).Msg("Global webhook configured from command line")
	}

	ex, err := os.Executable()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get executable path")
		panic(err)
	}
	exPath := filepath.Dir(ex)

	db, err := InitializeDatabase(exPath, *dataDir)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
		os.Exit(1)
	}
	// Defer cleanup of the database connection
	defer func() {
		if err := db.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close database connection")
		}
	}()

	// Initialize the schema
	if err = initializeSchema(db); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize schema")
		// Perform cleanup before exiting
		if err := db.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close database connection during cleanup")
		}
		os.Exit(1)
	}

	var dbLog waLog.Logger
	if *waDebug != "" {
		dbLog = waLog.Stdout("Database", *waDebug, false)
	}

	// Get database configuration
	config := getDatabaseConfig(exPath, *dataDir)
	storeConnStr := "file:" + filepath.Join(config.Path, "main.db") + "?_pragma=foreign_keys(1)&_busy_timeout=3000"
	container, err = sqlstore.New(context.Background(), "sqlite", storeConnStr, dbLog)

	if err != nil {
		log.Fatal().Err(err).Msg("Error creating sqlstore")
		os.Exit(1)
	}

	serverMode := HTTP
	if *mode == "stdio" {
		serverMode = Stdio
	}

	s := &server{
		router: mux.NewRouter(),
		db:     db,
		exPath: exPath,
		mode:   serverMode,
	}
	s.routes()

	s.connectOnStartup()

	if serverMode == Stdio {
		startStdioMode(s)
	} else {
		startHTTPMode(s)
	}
}

func startHTTPMode(s *server) {
	srv := &http.Server{
		Addr:              *address + ":" + *port,
		Handler:           s.router,
		ReadHeaderTimeout: 20 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       180 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	var once sync.Once

	// Wait for signals in a separate goroutine
	go func() {
		for {
			<-done
			once.Do(func() {
				log.Warn().Msg("Stopping server...")

				// Graceful shutdown logic
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if err := srv.Shutdown(ctx); err != nil {
					log.Error().Err(err).Msg("Failed to stop server")
					os.Exit(1)
				}

				log.Info().Msg("Server Exited Properly")
				os.Exit(0)
			})
		}
	}()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed to start")
		}
	}()
	log.Info().Str("address", *address).Str("port", *port).Msg("Server started. Waiting for connections...")
	select {}
}

func startStdioMode(s *server) {
	stdioServer := NewStdioServer(s)
	if err := stdioServer.Start(); err != nil {
		log.Error().Err(err).Msg("Stdio server error")
		os.Exit(1)
	}
	log.Info().Msg("Stdio server exited properly")
}
