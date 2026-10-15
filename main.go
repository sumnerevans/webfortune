package main

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/sumnerevans/webfortune/ogimage"
	"github.com/sumnerevans/webfortune/quotesfile"
	"github.com/sumnerevans/webfortune/templates"
)

//go:embed screenshot.png
var ogImage []byte

type Application struct {
	quotesfile     *quotesfile.Quotesfile
	pageParameters *templates.PageParameters
	imageMu        sync.Mutex
	imageCache     map[[16]byte][]byte
	imageOrder     [][16]byte
}

func NewApplication(quotesfilePath, hostRoot string) *Application {
	return &Application{
		quotesfile: quotesfile.NewQuotesfile(quotesfilePath),
		imageCache: make(map[[16]byte][]byte),
		pageParameters: &templates.PageParameters{
			Copyright:         os.Getenv("COPYRIGHT"),
			SourceURL:         templ.URL(os.Getenv("QUOTESFILE_SOURCE_URL")),
			GoatcounterDomain: os.Getenv("GOATCOUNTER_DOMAIN"),
			HostRoot:          hostRoot,
		},
	}
}

func (a *Application) quoteByHash(hashString string) (quotesfile.Quote, bool) {
	if len(hashString) != 32 {
		return quotesfile.Quote{}, false
	}
	hash, err := hex.DecodeString(hashString)
	if err != nil {
		return quotesfile.Quote{}, false
	}
	return a.quotesfile.GetQuoteByHash([16]byte(hash))
}

func (a *Application) Home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		quote, ok := a.quoteByHash(r.URL.Query().Get("id"))
		if !ok {
			quote = a.quotesfile.GetRandomQuote()
			http.Redirect(w, r, quote.Permalink(a.pageParameters.HostRoot), http.StatusFound)
			return
		}

		err := templates.Home(templates.HomePageParameters{
			PageParameters: a.pageParameters,
			Quote:          quote,
		}).Render(r.Context(), w)
		if err != nil {
			log.Err(err).Msg("Failed to execute the template")
		}
	}
}

func (a *Application) OGImage(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("hash")
	if !strings.HasSuffix(name, ".png") {
		http.NotFound(w, r)
		return
	}
	quote, ok := a.quoteByHash(strings.TrimSuffix(name, ".png"))
	if !ok {
		http.NotFound(w, r)
		return
	}

	hash := quote.Hash()
	etag := `"og-v6-` + quote.HashString() + `"`
	if r.Header.Get("If-None-Match") == etag {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Header().Set("ETag", etag)
		w.WriteHeader(http.StatusNotModified)
		return
	}

	a.imageMu.Lock()
	data, cached := a.imageCache[hash]
	a.imageMu.Unlock()
	if !cached {
		var err error
		data, err = ogimage.Render(quote)
		if err != nil {
			log.Err(err).Msg("Failed to render Open Graph image")
			http.Error(w, "Failed to render image", http.StatusInternalServerError)
			return
		}
		a.imageMu.Lock()
		if _, exists := a.imageCache[hash]; !exists {
			const maxCachedImages = 64
			if len(a.imageOrder) == maxCachedImages {
				delete(a.imageCache, a.imageOrder[0])
				a.imageOrder = a.imageOrder[1:]
			}
			a.imageCache[hash] = data
			a.imageOrder = append(a.imageOrder, hash)
		}
		a.imageMu.Unlock()
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", "image/png")
	w.Write(data)
}

func (a *Application) AllQuotes(w http.ResponseWriter, r *http.Request) {
	quotes := a.quotesfile.AllQuotes()
	err := templates.AllQuotes(templates.AllPageParameters{
		PageParameters: a.pageParameters,
		Quotes:         quotes,
	}).Render(r.Context(), w)
	if err != nil {
		log.Err(err).Msg("Failed to execute the template")
	}
}

func (a *Application) RawQuote(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(a.quotesfile.GetRandomQuote().Text()))
}

func (a *Application) HTMLQuote(w http.ResponseWriter, r *http.Request) {
	quote := a.quotesfile.GetRandomQuote()
	w.Header().Set("HX-Push-Url", quote.Permalink(a.pageParameters.HostRoot))
	if err := templates.QuoteButtons(a.pageParameters.HostRoot, quote).Render(r.Context(), w); err != nil {
		log.Err(err).Msg("Failed to execute the template")
	}
	if err := templates.Quote(quote).Render(r.Context(), w); err != nil {
		log.Err(err).Msg("Failed to execute the template")
	}
}

func (a *Application) Start(listen string) {
	log.Info().Msg("Starting router")

	http.Handle("/", a.Home())
	http.HandleFunc("/all", a.AllQuotes)
	http.HandleFunc("/raw", a.RawQuote)
	http.HandleFunc("/html", a.HTMLQuote)
	http.HandleFunc("/og-image.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "og-image.png", time.Time{}, bytes.NewReader(ogImage))
	})
	http.HandleFunc("GET /og/v6/{hash}", a.OGImage)

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
