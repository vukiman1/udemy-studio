package httpapi

import (
	"bytes"
	_ "embed"
)

var (
	//go:embed page_themes/palette.css
	paletteCSS string
	//go:embed page_themes/article.css
	articleThemeCSS string
	//go:embed page_themes/quiz.css
	quizThemeCSS string
)

var headEnd = []byte("</head>")

// withPageTheme dresses a downloaded article or quiz page in the app's colours. The style goes last in the head so it
// overrides the page's own, and stays inline because the pages' sandbox policy allows no stylesheet from the app.
func withPageTheme(page []byte, themeCSS string) []byte {
	style := []byte("<style>" + paletteCSS + themeCSS + "</style>")
	position := bytes.Index(bytes.ToLower(page), headEnd)
	if position < 0 {
		return append(style, page...)
	}
	themed := make([]byte, 0, len(page)+len(style))
	themed = append(themed, page[:position]...)
	themed = append(themed, style...)
	return append(themed, page[position:]...)
}
