package main

import (
	"embed"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

const (
	PrefixEnv = "BIFROST"
)

var (
	host      string
	port      string
	debugMode bool
	quietMode bool

	//go:embed web/dist/*
	web embed.FS
)

func init() {
	host = os.Getenv(fmt.Sprintf("%s_%s", PrefixEnv, "HOST"))
	port = os.Getenv(fmt.Sprintf("%s_%s", PrefixEnv, "PORT"))
	debugMode = os.Getenv(fmt.Sprintf("%s_%s", PrefixEnv, "DEBUG")) == "true"
	quietMode = os.Getenv(fmt.Sprintf("%s_%s", PrefixEnv, "QUIET")) == "true"

	flag.BoolVar(&debugMode, "debug", false, "sets log level to debug")
	flag.BoolVar(&quietMode, "quiet", false, "disable log")
	flag.StringVar(&host, "host", "0.0.0.0", "host address")
	flag.StringVar(&port, "port", "8080", "port listening")
	flag.Parse()

	log.Info().Msgf("Debug mode: %v", debugMode)
	log.Info().Msgf("Quiet mode: %v", quietMode)

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debugMode {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	if quietMode {
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}

	log.Logger = zerolog.New(os.Stdout).
		Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
		With().Timestamp().Caller().Stack().Logger()
}

func main() {
	app := fiber.New(fiber.Config{
		Views: html.NewFileSystem(http.FS(web), ".html"),
	})

	app.Use(
		logger.New(),
		etag.New(),
	)

	// app.Get("/static")
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{})
	})

	prometheus := fiberprometheus.New("bifrost")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	log.Info().Msg("Start Application")
	if err := app.Listen(fmt.Sprintf("%s:%s", host, port)); err != nil {
		log.Error().Err(err).Msg("")
	}
}