package main

import (
	"embed"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:embed templates/*
var TemplateFS embed.FS

type Application struct {
	quotesfile *Quotesfile
	sourceURL  string
}

func NewApplication(quotesfile, sourceURL string) *Application {
	return &Application{
		quotesfile: NewQuotesfile(quotesfile),
		sourceURL:  sourceURL,
	}
}

type HomeTemplateData struct {
	Quote     template.HTML
	SourceURL string
}

func (a *Application) Home() http.HandlerFunc {
	template, err := template.ParseFS(TemplateFS, "templates/home.html")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse template")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		templateData := HomeTemplateData{
			Quote:     a.quotesfile.GetRandomQuote().HTML(),
			SourceURL: a.sourceURL,
		}
		if err := template.ExecuteTemplate(w, "home.html", templateData); err != nil {
			log.Err(err).Msg("Failed to execute the template")
		}
	}
}

func (a *Application) AllQuotes(w http.ResponseWriter, r *http.Request) {
	for _, quote := range a.quotesfile.quotes {
		w.Write([]byte(quote.Text()))
		w.Write([]byte("%\n"))
	}
}

func (a *Application) RawQuote(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(a.quotesfile.GetRandomQuote().Text()))
}

func (a *Application) HTMLQuote(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(a.quotesfile.GetRandomQuote().HTML()))
}

func (a *Application) Start(listen string) {
	log.Info().Msg("Starting router")

	http.HandleFunc("/", a.Home())
	http.HandleFunc("/all", a.AllQuotes)
	http.HandleFunc("/raw", a.RawQuote)
	http.HandleFunc("/html", a.HTMLQuote)

	log.Info().Str("listen", listen).Msg("Starting server")
	err := http.ListenAndServe(listen, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}

func main() {
	logger := log.Output(os.Stdout)
	if os.Getenv("LOG_CONSOLE") != "" {
		logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}
	logger.Info().Msg("backend starting...")

	log.Logger = logger

	quotesfile := os.Getenv("QUOTESFILE")
	if quotesfile == "" {
		log.Fatal().Msg("QUOTESFILE not set")
	}

	app := NewApplication(quotesfile, os.Getenv("QUOTESFILE_SOURCE_URL"))

	listen := os.Getenv("LISTEN_ADDR")
	app.Start(listen)
}
