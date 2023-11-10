package main

import (
	"embed"
	"encoding/hex"
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
	quotesfile      *Quotesfile
	sourceURL       string
	plausibleDomain string
	hostRoot        string
}

func NewApplication(quotesfile, hostRoot string) *Application {
	return &Application{
		quotesfile:      NewQuotesfile(quotesfile),
		sourceURL:       os.Getenv("QUOTESFILE_SOURCE_URL"),
		plausibleDomain: os.Getenv("PLAUSIBLE_DOMAIN"),
		hostRoot:        hostRoot,
	}
}

type HomeTemplateData struct {
	Wrapped         template.HTML
	SourceURL       string
	PlausibleDomain string
}

func (a *Application) Home() http.HandlerFunc {
	template, err := template.ParseFS(TemplateFS, "templates/home.html")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse template")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		quoteHash := r.URL.Query().Get("id")
		var quote Quote
		var ok bool
		if len(quoteHash) == 32 {
			hash, err := hex.DecodeString(quoteHash)
			if err == nil {
				quote, ok = a.quotesfile.GetQuoteByHash([16]byte(hash))
			}
		}
		if !ok {
			quote = a.quotesfile.GetRandomQuote()
		}

		templateData := HomeTemplateData{
			Wrapped:         quote.HTML(a.hostRoot),
			SourceURL:       a.sourceURL,
			PlausibleDomain: a.plausibleDomain,
		}
		if err := template.ExecuteTemplate(w, "home.html", templateData); err != nil {
			log.Err(err).Msg("Failed to execute the template")
		}
	}
}

func (a *Application) AllQuotes(w http.ResponseWriter, r *http.Request) {
	for _, quote := range a.quotesfile.quotes {
		w.Write([]byte(quote.Text()))
		w.Write([]byte("\n%\n"))
	}
}

func (a *Application) RawQuote(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(a.quotesfile.GetRandomQuote().Text()))
}

func (a *Application) HTMLQuote(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(a.quotesfile.GetRandomQuote().HTML(a.hostRoot)))
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

	listen := os.Getenv("LISTEN_ADDR")
	hostRoot := os.Getenv("HOST_ROOT")
	if hostRoot == "" {
		hostRoot = listen
	}

	app := NewApplication(quotesfile, hostRoot)
	app.Start(listen)
}
