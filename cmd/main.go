package main

import (
	"context"
	"csjk-bk/internal/app/docs"
	"csjk-bk/internal/app/router"
	"csjk-bk/internal/module/alert"
	"csjk-bk/internal/module/slurm"
	"csjk-bk/internal/pkg/client/alertmanager"
	"csjk-bk/internal/pkg/client/postgres"
	"csjk-bk/internal/pkg/client/slurmrest"
	"csjk-bk/internal/pkg/log"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/common/version"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           csjk-bk
// @version         0.0.1-alpha
// @description     csjk backend
// @schema			http
// @BasePath        /api/v1
// @contact.email	hecheng@nscc-tj.cn
func main() {
	var (
		logOutput          string
		logFormat          string
		logFile            string
		logLevel           string
		slurmrestTimeout   time.Duration
		srvlisenAddr       string
		srvshutdownTimeout time.Duration
	)
	app := kingpin.New(filepath.Base(os.Args[0]), "csjk backend server.")
	app.HelpFlag.Short('h')
	// Logging related flags
	app.Flag("log.level", "Log level, one of [debug, info, warn, error].").Default("info").EnumVar(&logLevel, "debug", "info", "warn", "error")
	app.Flag("log.output", "Log output, one of [stdout, stderr, file].").Default("stderr").EnumVar(&logOutput, "stdout", "stderr", "file")
	app.Flag("log.format", "Log format, one of [json, text].").Default("text").EnumVar(&logFormat, "json", "text")
	app.Flag("log.file", "Log file path when --output=file.").PlaceHolder("PATH").StringVar(&logFile)
	app.Flag("slrumrest.timeout", "Timeout for slurmrestd(official or customize) HTTP requests (Go duration, e.g. 5s, 1m).").Default("5s").DurationVar(&slurmrestTimeout)
	app.Flag("server.listen-addr", "Server listen address (e.g. :8080 or 127.0.0.1:8080)").Default(":8081").StringVar(&srvlisenAddr)
	app.Flag("server.shutdown-timeout", "Graceful shutdown timeout (e.g. 10s)").Default("10s").DurationVar(&srvshutdownTimeout)
	// Cross-flag validation
	app.PreAction(func(*kingpin.ParseContext) error {
		if strings.EqualFold(logOutput, "file") {
			if !isValidFilePath(logFile) {
				return fmt.Errorf("invalid --file path: %q", logFile)
			}
		}
		return nil
	})
	app.Version(version.Print("csbk-jk"))

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("failed to parse commandline arguments: %w", err))
		app.Usage(os.Args[1:])
		os.Exit(2)
	}
	// 创建 Logger
	logger, logClose, err := log.NewLogger(logOutput, logFormat, logFile, logLevel)
	if err != nil {
		fmt.Fprintf(os.Stdout, "unable to create logger: %w", err)
		return
	}
	defer logClose()

	// 创建各模块路由
	dbctx, dbcancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dbcancel()
	db, err := postgres.New(dbctx, "postgres://netbox:koC2FR3SfBNNpCUX@192.168.2.35:5432/monitor?sslmode=disable")
	defer db.Close()
	slurmrestClient := slurmrest.New(http.DefaultClient, slurmrestTimeout, logger)
	slurmRouter := slurm.NewRouter(db, slurmrestClient, logger)
	amClient := &alertmanager.Client{}
	alertRouter := alert.NewRouter(db, amClient, logger)
	// Build router
	r := router.New()

	docs.SwaggerInfo.BasePath = "/api/v1"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 注册所有模块（也可做“按需编译”或通过 build tag 控制）
	router.Register(
		slurmRouter,
		alertRouter,
	)
	router.Mount(r)
	srv := &http.Server{
		Addr:              srvlisenAddr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Start server in background
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server listening", slog.String("addr", srvlisenAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-serverErr:
		if err != nil {
			logger.Error("server failed", slog.Any("err", err))
			os.Exit(1)
		}
	case <-quit:
		// proceed to shutdown
	}
	logger.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), srvshutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", slog.Any("err", err))
	}
	logger.Info("server exiting")
}

// isValidFilePath performs a light-weight validation for file paths.
// It accepts both absolute and relative paths and rejects empty paths
// or paths that end with a path separator (which usually indicate a directory).
func isValidFilePath(p string) bool {
	if strings.TrimSpace(p) == "" {
		return false
	}
	// Reject paths that end with a separator, which imply directories
	if strings.HasSuffix(p, string(os.PathSeparator)) {
		return false
	}
	base := filepath.Base(p)
	if base == "." || base == string(os.PathSeparator) {
		return false
	}
	return true
}
