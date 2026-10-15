package ogimage

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"strings"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"github.com/sumnerevans/webfortune/quotesfile"
)

const (
	Width  = 1200
	Height = 630
)

//go:embed Iosevka-Regular.ttf.gz
var fontData []byte

var (
	fontOnce   sync.Once
	parsedFont *opentype.Font
	fontErr    error
)

func face(size int) (font.Face, error) {
	fontOnce.Do(func() {
		reader, err := gzip.NewReader(bytes.NewReader(fontData))
		if err != nil {
			fontErr = err
			return
		}
		defer reader.Close()
		data, err := io.ReadAll(reader)
		if err != nil {
			fontErr = err
			return
		}
		parsedFont, fontErr = opentype.Parse(data)
	})
	if fontErr != nil {
		return nil, fontErr
	}
	return opentype.NewFace(parsedFont, &opentype.FaceOptions{Size: float64(size), DPI: 72, Hinting: font.HintingFull})
}

func Render(quote quotesfile.Quote) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, Width, Height))
	fill(img, img.Bounds(), color.RGBA{42, 47, 52, 255})
	fill(img, image.Rect(0, 0, Width, 86), color.RGBA{33, 37, 41, 255})
	fill(img, image.Rect(0, 0, 8, 86), color.RGBA{25, 135, 84, 255})

	white := color.RGBA{248, 249, 250, 255}
	muted := color.RGBA{173, 181, 189, 255}
	green := color.RGBA{25, 135, 84, 255}

	brandFace, err := face(36)
	if err != nil {
		return nil, err
	}
	defer brandFace.Close()
	drawText(img, brandFace, white, 52, 59, "WEB FORTUNE")

	markFace, err := face(110)
	if err != nil {
		return nil, err
	}
	defer markFace.Close()

	source := strings.Join(strings.Fields(quote.Source()), " ")
	const textX = 125
	const textWidth = 1050
	const firstBaseline = 177
	lines := wrapColumns(quote.BodyText(), 72)
	var bodyFace font.Face
	var bodySize, lineHeight, maxLines int
	for size := 32; size >= 14; size-- {
		candidate, err := face(size)
		if err != nil {
			return nil, err
		}
		if !linesFit(candidate, lines, textWidth) {
			candidate.Close()
			continue
		}
		height := size * 3 / 2
		limit := (Height-40-firstBaseline)/height + 1
		if source != "" {
			limit = (Height - 40 - firstBaseline - size*3/4) / height
		}
		if len(lines) <= limit || size == 14 {
			bodyFace, bodySize, lineHeight, maxLines = candidate, size, height, limit
			break
		}
		candidate.Close()
	}
	defer bodyFace.Close()
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines[maxLines-1] = fitLine(bodyFace, lines[maxLines-1], textWidth, true)
	}
	firstInkTop := firstBaseline - bodySize
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		bounds, _ := font.BoundString(bodyFace, line)
		firstInkTop = firstBaseline + i*lineHeight + bounds.Min.Y.Floor()
		break
	}
	markBounds, _ := font.BoundString(markFace, "“")
	markBaseline := firstInkTop - markBounds.Min.Y.Floor()
	sourceBaseline := firstBaseline + len(lines)*lineHeight + bodySize*3/4
	contentBottom := max(markBaseline+markBounds.Max.Y.Ceil(), firstBaseline+(len(lines)-1)*lineHeight+bodySize/4)
	if source != "" {
		contentBottom = max(contentBottom, sourceBaseline+8)
	}
	// Keep short quotes centered in the space below the header while keeping
	// the attribution attached to the quote rather than pinned to the bottom.
	verticalOffset := max(0, (86+Height-firstInkTop-contentBottom)/2)
	verticalOffset = min(verticalOffset, Height-32-contentBottom)
	drawText(img, markFace, green, 36, markBaseline+verticalOffset, "“")
	for i, line := range lines {
		drawText(img, bodyFace, white, textX, firstBaseline+verticalOffset+i*lineHeight, line)
	}

	if source != "" {
		sourceSize := bodySize
		sourceFace, err := face(sourceSize)
		if err != nil {
			return nil, err
		}
		defer sourceFace.Close()
		// The page's figcaption starts four character cells to the right of
		// the blockquote, so keep that same indentation in the image.
		sourceIndent := width(sourceFace, "    ")
		sourceX := textX + sourceIndent
		drawText(img, sourceFace, muted, sourceX, sourceBaseline+verticalOffset, fitLine(sourceFace, "— "+source, textWidth-sourceIndent, false))
	}

	var output bytes.Buffer
	err = (&png.Encoder{CompressionLevel: png.BestSpeed}).Encode(&output, img)
	return output.Bytes(), err
}

func fill(img draw.Image, rect image.Rectangle, c color.Color) {
	draw.Draw(img, rect, image.NewUniform(c), image.Point{}, draw.Src)
}

func drawText(img draw.Image, face font.Face, c color.Color, x, baseline int, text string) {
	d := font.Drawer{Dst: img, Src: image.NewUniform(c), Face: face, Dot: fixed.P(x, baseline)}
	d.DrawString(text)
}

func width(face font.Face, text string) int {
	d := font.Drawer{Face: face}
	return d.MeasureString(text).Ceil()
}

func linesFit(face font.Face, lines []string, maxWidth int) bool {
	for _, line := range lines {
		if width(face, line) > maxWidth {
			return false
		}
	}
	return true
}

// wrapColumns keeps the original spacing and line breaks. Only lines wider than
// the card's 72-column text area are split.
func wrapColumns(text string, columns int) []string {
	var lines []string
	for _, original := range strings.Split(strings.ReplaceAll(text, "\r", ""), "\n") {
		var line strings.Builder
		column := 0
		appendRune := func(r rune) {
			if column == columns {
				lines = append(lines, line.String())
				line.Reset()
				column = 0
			}
			line.WriteRune(r)
			column++
		}
		for _, r := range original {
			if r == '\t' {
				spaces := 8 - column%8
				for range spaces {
					appendRune(' ')
				}
				continue
			}
			appendRune(r)
		}
		if column > 0 || original == "" {
			lines = append(lines, line.String())
		}
	}
	return lines
}

func fitLine(face font.Face, line string, maxWidth int, forceEllipsis bool) string {
	if !forceEllipsis && width(face, line) <= maxWidth {
		return line
	}
	runes := []rune(strings.TrimRight(line, " "))
	for len(runes) > 0 && width(face, string(runes)+"…") > maxWidth {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}
