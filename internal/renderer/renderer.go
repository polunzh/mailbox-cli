package renderer

import "github.com/k3a/html2text"

// HTMLToText converts HTML content to plain text.
func HTMLToText(html string) string {
	return html2text.HTML2Text(html)
}
