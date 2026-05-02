package templates

import (
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
