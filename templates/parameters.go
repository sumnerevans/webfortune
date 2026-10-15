package templates

import (
	"strings"

	"github.com/a-h/templ"

	"github.com/sumnerevans/webfortune/quotesfile"
)

type PageParameters struct {
	Copyright         string
	HostRoot          string
	GoatcounterDomain string
	SourceURL         templ.SafeURL
}

type AllPageParameters struct {
	*PageParameters
	Quotes []quotesfile.Quote
}

type HomePageParameters struct {
	*PageParameters
	Quote quotesfile.Quote
}

func quoteDescription(quote quotesfile.Quote) string {
	description := []rune(strings.Join(strings.Fields(quote.Text()), " "))
	if len(description) > 200 {
		return string(description[:199]) + "…"
	}
	return string(description)
}

func quoteTitle(quote quotesfile.Quote) string {
	description := []rune(quoteDescription(quote))
	if len(description) > 80 {
		return string(description[:79]) + "… | Web Fortune"
	}
	if len(description) == 0 {
		return "Quote | Web Fortune"
	}
	return string(description) + " | Web Fortune"
}
