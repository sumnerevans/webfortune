package quotesfile

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"strings"

	"github.com/a-h/templ"
	"github.com/mitchellh/go-wordwrap"
)

type Quote struct {
	quote        []string
	spaceEscaped []string
	source       string
}

func (q Quote) Hash() [16]byte {
	return md5.Sum([]byte(q.Text()))
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

func (q Quote) SourceHTML() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		if q.source == "" {
			return
		}
		_, err = io.WriteString(w, `<figcaption class="mt-3">`)
		if err != nil {
			return
		}
		sourceContextStr := wordwrap.WrapString(q.source, 65)
		for i, s := range strings.Split(sourceContextStr, "\n") {
			if i == 0 {
				_, err = io.WriteString(w, "&nbsp;&nbsp;&nbsp;&nbsp;&mdash;&nbsp;")
			} else {
				_, err = io.WriteString(w, "&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;")
			}
			if err != nil {
				return
			}
			_, err = io.WriteString(w, s)
			if err != nil {
				return
			}
			_, err = io.WriteString(w, "<br>")
			if err != nil {
				return
			}
		}
		_, err = io.WriteString(w, "</figcaption>")
		return
	})
}

func (q Quote) QuoteHTML() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, err = io.WriteString(w, strings.Join(q.spaceEscaped, "<br>"))
		return
	})
}

func (q Quote) Permalink(hostRoot string) string {
	hash := q.Hash()
	return hostRoot + "/?id=" + hex.EncodeToString(hash[:])
}
