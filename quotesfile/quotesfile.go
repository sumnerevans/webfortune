package quotesfile

import (
	"bufio"
	"math/rand"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

type Quotesfile struct {
	quotes []Quote
	byHash map[[16]byte]Quote
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
	byHash := map[[16]byte]Quote{}
	var inSource bool
	for scanner.Scan() {
		text := scanner.Text()
		if text == "%" {
			quotes = append(quotes, quote)
			byHash[quote.Hash()] = quote
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

	return &Quotesfile{
		quotes: quotes,
		byHash: byHash,
	}
}

func (q *Quotesfile) AllQuotes() []Quote {
	return q.quotes
}

func (q *Quotesfile) GetRandomQuote() Quote {
	return q.quotes[rand.Intn(len(q.quotes))]
}

func (q *Quotesfile) GetQuoteByHash(hash [16]byte) (Quote, bool) {
	quote, ok := q.byHash[hash]
	return quote, ok
}
