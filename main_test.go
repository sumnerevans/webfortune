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
)

func TestQuoteOpenGraphMetadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "quotes")
	if err := os.WriteFile(path, []byte("First quote\n%\nSecond quote\n%\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	app := NewApplication(path, "https://fortune.example")

	for _, quote := range app.quotesfile.AllQuotes() {
		t.Run(quote.Text(), func(t *testing.T) {
			url := quote.Permalink(app.pageParameters.HostRoot)
			response := httptest.NewRecorder()
			app.Home().ServeHTTP(response, httptest.NewRequest(http.MethodGet, url, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", response.Code)
			}
			body := response.Body.String()
			if !strings.Contains(body, "<title>Web Fortune</title>") {
				t.Error("response missing Web Fortune page title")
			}
			for _, tag := range []string{
				`<meta property="og:title" content="` + quote.Text() + ` | Web Fortune"`,
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
