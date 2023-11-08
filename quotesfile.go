package main

import (
	"bufio"
	"html/template"
	"math/rand"
	"os"
	"strings"

	"github.com/mitchellh/go-wordwrap"
	"github.com/rs/zerolog/log"
)

type Quote struct {
	quote        []string
	spaceEscaped []string
	source       string
}

func (q Quote) Text() string {
	var builder strings.Builder
	for i, line := range q.quote {
		builder.WriteString(line)
		if i < len(q.quote)-1 {
			builder.WriteString("\n")
		}
	}

	if q.source != "" {
		builder.WriteString("\n")
		sourceLines := strings.Split(wordwrap.WrapString(q.source, 65), "\n")
		for i, s := range sourceLines {
			if i == 0 {
				builder.WriteString("    -- ")
			} else {
				builder.WriteString("       ")
			}
			builder.WriteString(s)
			if i < len(sourceLines)-1 {
				builder.WriteString("\n")
			}
		}
	}

	return builder.String()
}

func (q Quote) HTML() template.HTML {
	var builder strings.Builder
	builder.WriteString(`<div id="plain-quote" class="d-none">`)
	builder.WriteString(q.Text())
	builder.WriteString(`</div>`)
	builder.WriteString(`<figure id="quote" class="quote p-4 m-0">`)
	builder.WriteString(`<blockquote class="m-0">`)
	builder.WriteString(strings.Join(q.spaceEscaped, "<br>"))
	builder.WriteString("</blockquote>")
	if q.source != "" {
		builder.WriteString(`<figcaption class="mt-3">`)
		sourceContextStr := wordwrap.WrapString(q.source, 65)
		for i, s := range strings.Split(sourceContextStr, "\n") {
			if i == 0 {
				builder.WriteString(`&nbsp;&nbsp;&nbsp;&nbsp;&mdash;&nbsp;`)
			} else {
				builder.WriteString(`&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;`)
			}
			builder.WriteString(s)
			builder.WriteString("<br>")
		}
		builder.WriteString("</figcaption>")
	}
	builder.WriteString("</figure>")
	return template.HTML(builder.String())
}

type Quotesfile struct {
	quotes []Quote
}

func NewQuotesfile(quotesfile string) *Quotesfile {
	file, err := os.Open(quotesfile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open quotes file")
	}
	defer file.Close()

	var quotes []Quote

	scanner := bufio.NewScanner(file)
	var quote Quote
	var inSource bool
	for scanner.Scan() {
		text := scanner.Text()
		if text == "%" {
			quotes = append(quotes, quote)
			quote = Quote{}
			inSource = false
		} else if inSource {
			quote.source = quote.source + " " + strings.TrimSpace(text)
		} else if strings.HasPrefix(text, "    -- ") {
			inSource = true
			quote.source = text[7:]
		} else {
			quote.quote = append(quote.quote, text)
			quote.spaceEscaped = append(quote.spaceEscaped, strings.ReplaceAll(text, " ", "&nbsp;"))
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal().Err(err).Msg("Failed to read quotes file")
	}

	return &Quotesfile{quotes: quotes}
}

func (q *Quotesfile) GetRandomQuote() Quote {
	return q.quotes[rand.Intn(len(q.quotes))]
}
