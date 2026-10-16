package main

import (
	"bytes"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sumnerevans/webfortune/quotesfile"
)

func TestQuoteOpenGraphMetadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "quotes")
	if err := os.WriteFile(path, []byte("First quote\n%\nSecond quote\n%\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	app := NewApplication(path, "https://fortune.example")
	routes := app.routes()

	for _, quote := range app.quotesfile.AllQuotes() {
		t.Run(quote.Text(), func(t *testing.T) {
			url := "https://fortune.example/q/" + quote.HashString()
			if got := quote.Permalink(app.pageParameters.HostRoot); got != url {
				t.Fatalf("permalink = %q, want %q", got, url)
			}
			response := httptest.NewRecorder()
			routes.ServeHTTP(response, httptest.NewRequest(http.MethodGet, url, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", response.Code)
			}
			body := response.Body.String()
			if !strings.Contains(body, "<title>Web Fortune</title>") {
				t.Error("response missing Web Fortune page title")
			}
			for _, tag := range []string{
				`<meta property="og:title" content="Web Fortune"`,
				`<meta property="og:description" content="` + quote.Text() + `"`,
				`<meta property="og:url" content="` + url + `"`,
				`<link rel="canonical" href="` + url + `"`,
				`<meta property="og:image" content="` + quote.OGImageURL(app.pageParameters.HostRoot) + `"`,
				`href="` + quote.OGImageURL(app.pageParameters.HostRoot) + `" download="` + quote.HashString() + `.png"`,
			} {
				if !strings.Contains(body, tag) {
					t.Errorf("response missing %q", tag)
				}
			}
		})
	}
}

func TestQuoteRoutes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "quotes")
	if err := os.WriteFile(path, []byte("First quote\n%\nSecond quote\n%\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	app := NewApplication(path, "https://fortune.example")
	routes := app.routes()
	quote := app.quotesfile.AllQuotes()[0]
	canonical := quote.Permalink(app.pageParameters.HostRoot)

	for _, test := range []struct {
		name, url, location string
		status              int
	}{
		{"legacy query", "https://fortune.example/?id=" + quote.HashString(), canonical, http.StatusMovedPermanently},
		{"uppercase path", "https://fortune.example/q/" + strings.ToUpper(quote.HashString()), canonical, http.StatusMovedPermanently},
		{"random root", "https://fortune.example/", "", http.StatusFound},
		{"invalid legacy hash", "https://fortune.example/?id=invalid", "", http.StatusNotFound},
		{"empty legacy hash", "https://fortune.example/?id=", "", http.StatusNotFound},
		{"invalid path hash", "https://fortune.example/q/invalid", "", http.StatusNotFound},
		{"unknown path hash", "https://fortune.example/q/00000000000000000000000000000000", "", http.StatusNotFound},
		{"unknown path", "https://fortune.example/unknown", "", http.StatusNotFound},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			routes.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.url, nil))
			if response.Code != test.status {
				t.Errorf("status = %d, want %d", response.Code, test.status)
			}
			if test.location != "" && response.Header().Get("Location") != test.location {
				t.Errorf("Location = %q, want %q", response.Header().Get("Location"), test.location)
			}
			if test.name == "random root" {
				location := response.Header().Get("Location")
				if _, ok := quoteForURL(app, location); !ok {
					t.Errorf("random redirect location = %q, want known quote permalink", location)
				}
			}
		})
	}
}

func quoteForURL(app *Application, url string) (quotesfile.Quote, bool) {
	for _, quote := range app.quotesfile.AllQuotes() {
		if quote.Permalink(app.pageParameters.HostRoot) == url {
			return quote, true
		}
	}
	return quotesfile.Quote{}, false
}

func TestAllQuotesImageDownloadLinks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "quotes")
	if err := os.WriteFile(path, []byte("First quote\n%\nSecond quote\n%\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	app := NewApplication(path, "https://fortune.example")
	response := httptest.NewRecorder()
	app.AllQuotes(response, httptest.NewRequest(http.MethodGet, "https://fortune.example/all", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	for _, quote := range app.quotesfile.AllQuotes() {
		link := `href="` + quote.OGImageURL(app.pageParameters.HostRoot) + `" download="` + quote.HashString() + `.png"`
		if !strings.Contains(response.Body.String(), link) {
			t.Errorf("all quotes page missing image download link %q", link)
		}
		if !strings.Contains(response.Body.String(), quote.Permalink(app.pageParameters.HostRoot)) {
			t.Errorf("all quotes page missing quote permalink %q", quote.Permalink(app.pageParameters.HostRoot))
		}
	}
}

func TestHTMLQuotePushesPathPermalink(t *testing.T) {
	path := filepath.Join(t.TempDir(), "quotes")
	if err := os.WriteFile(path, []byte("First quote\n%\nSecond quote\n%\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	app := NewApplication(path, "https://fortune.example")
	response := httptest.NewRecorder()
	app.routes().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "https://fortune.example/html", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if _, ok := quoteForURL(app, response.Header().Get("HX-Push-Url")); !ok {
		t.Errorf("HX-Push-Url = %q, want known quote permalink", response.Header().Get("HX-Push-Url"))
	}
}

func TestQuoteOpenGraphImage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "quotes")
	if err := os.WriteFile(path, []byte("A quote with a source\n    -- Author\n%\nA different quote\n%\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	app := NewApplication(path, "https://fortune.example")
	quote := app.quotesfile.AllQuotes()[0]
	request := httptest.NewRequest(http.MethodGet, quote.OGImageURL(app.pageParameters.HostRoot), nil)
	request.SetPathValue("hash", quote.HashString()+".png")
	response := httptest.NewRecorder()
	app.OGImage(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if got := response.Header().Get("Content-Type"); got != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", got)
	}
	img, err := png.Decode(bytes.NewReader(response.Body.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != 1200 || img.Bounds().Dy() != 630 {
		t.Errorf("image bounds = %v, want 1200x630", img.Bounds())
	}
	if len(app.imageCache) != 1 {
		t.Errorf("cached images = %d, want 1", len(app.imageCache))
	}
	other := app.quotesfile.AllQuotes()[1]
	otherRequest := httptest.NewRequest(http.MethodGet, other.OGImageURL(app.pageParameters.HostRoot), nil)
	otherRequest.SetPathValue("hash", other.HashString()+".png")
	otherResponse := httptest.NewRecorder()
	app.OGImage(otherResponse, otherRequest)
	if otherResponse.Code != http.StatusOK || bytes.Equal(response.Body.Bytes(), otherResponse.Body.Bytes()) {
		t.Error("different quotes should produce different PNGs")
	}

	conditional := httptest.NewRequest(http.MethodGet, quote.OGImageURL(app.pageParameters.HostRoot), nil)
	conditional.SetPathValue("hash", quote.HashString()+".png")
	conditional.Header.Set("If-None-Match", response.Header().Get("ETag"))
	notModified := httptest.NewRecorder()
	app.OGImage(notModified, conditional)
	if notModified.Code != http.StatusNotModified || notModified.Body.Len() != 0 {
		t.Errorf("conditional response = %d, %d bytes; want 304 with no body", notModified.Code, notModified.Body.Len())
	}

	invalid := httptest.NewRequest(http.MethodGet, "https://fortune.example/og/v6/invalid.png", nil)
	invalid.SetPathValue("hash", "invalid.png")
	notFound := httptest.NewRecorder()
	app.OGImage(notFound, invalid)
	if notFound.Code != http.StatusNotFound {
		t.Errorf("unknown image status = %d, want 404", notFound.Code)
	}
}
