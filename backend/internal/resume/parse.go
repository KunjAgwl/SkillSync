package resume

import (
	"bytes"
	"strings"

	"github.com/ledongthuc/pdf"
)

// Parse extracts text from raw file data (bytes).
// It attempts to detect the content type. For now, it assumes PDF or Text.
func Parse(data []byte, filename string) (string, error) {
	if strings.HasSuffix(strings.ToLower(filename), ".pdf") {
		return parsePDF(data)
	}
	// Fallback to treating as plain text
	return string(data), nil
}

func parsePDF(data []byte) (string, error) {
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}

	var textBuilder strings.Builder
	totalPage := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		text, err := p.GetPlainText(nil)
		if err != nil {
			// Continue to next page rather than failing entire parse?
			// Or return partial? Let's log/continue.
			continue
		}
		textBuilder.WriteString(text)
		textBuilder.WriteString("\n")
	}

	return strings.TrimSpace(textBuilder.String()), nil
}

// Legacy string-only parser (if needed, but we should use bytes now)
func ParseText(text string) string {
	return strings.TrimSpace(text)
}
