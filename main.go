package main

import (
	"encoding/hex"
	"net/http"
	"os"
	"time"

	"github.com/a-h/templ"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/sumnerevans/webfortune/quotesfile"
	"github.com/sumnerevans/webfortune/templates"
)

type Application struct {
	quotesfile        *quotesfile.Quotesfile
	sourceURL         templ.SafeURL
	goatcounterDomain string
	hostRoot          string
}

func NewApplication(quotesfilePath, hostRoot string) *Application {
	return &Application{
		quotesfile:        quotesfile.NewQuotesfile(quotesfilePath),
		sourceURL:         templ.URL(os.Getenv("QUOTESFILE_SOURCE_URL")),
		goatcounterDomain: os.Getenv("GOATCOUNTER_DOMAIN"),
		hostRoot:          hostRoot,
	}
}

func (a *Application) Home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		quoteHash := r.URL.Query().Get("id")
		var quote quotesfile.Quote
		var ok bool
		if len(quoteHash) == 32 {
			hash, err := hex.DecodeString(quoteHash)
			if err == nil {
				quote, ok = a.quotesfile.GetQuoteByHash([16]byte(hash))
			}
		}
		if !ok {
			quote = a.quotesfile.GetRandomQuote()
			http.Redirect(w, r, quote.Permalink(a.hostRoot), http.StatusFound)
			return
		}

		err := templates.Home(templates.PageParameters{
			HostRoot:          a.hostRoot,
			GoatcounterDomain: a.goatcounterDomain,
			Quote:             quote,
			SourceURL:         a.sourceURL,
		}).Render(r.Context(), w)
		if err != nil {
			log.Err(err).Msg("Failed to execute the template")
		}
	}
}

func (a *Application) AllQuotes(w http.ResponseWriter, r *http.Request) {
	for _, quote := range a.quotesfile.AllQuotes() {
		w.Write([]byte(quote.Text()))
		w.Write([]byte("\n%\n"))
	}
}

func (a *Application) RawQuote(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(a.quotesfile.GetRandomQuote().Text()))
}

func (a *Application) HTMLQuote(w http.ResponseWriter, r *http.Request) {
	quote := a.quotesfile.GetRandomQuote()
	w.Header().Set("HX-Push-Url", quote.Permalink(a.hostRoot))
	if err := templates.Quote(a.hostRoot, quote).Render(r.Context(), w); err != nil {
		log.Err(err).Msg("Failed to execute the template")
	}
}

func (a *Application) Start(listen string) {
	log.Info().Msg("Starting router")

	http.Handle("/", a.Home())
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
