package autoklept

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"strings"
)

// ElementNodeFinder lets the user specify a particular tag to start parsing from, instead of just parsing the whole input.
// Example: <div id="123abc">
type ElementNodeFinder struct {
	Tag     string
	AttrKey string
	AttrVal string
}

func parseHtmlByTag(htmlBody string, lookup *ElementNodeFinder) (*bytes.Buffer, error) {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return nil, fmt.Errorf("error parsing html: %w", err)
	}
	buf := bytes.Buffer{}
	if lookup != nil {
		content := findElementNode(doc, *lookup)
		if err = html.Render(&buf, content); err != nil {
			return nil, fmt.Errorf("error rendering html: %w", err)
		}
	} else {
		buf.WriteString(htmlBody)
	}
	return &buf, nil
}

func findElementNode(n *html.Node, lookup ElementNodeFinder) *html.Node {
	if n.Type == html.ElementNode && n.Data == lookup.Tag {
		for _, attr := range n.Attr {
			if attr.Key == lookup.AttrKey && attr.Val == lookup.AttrVal {
				return n
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findElementNode(c, lookup); found != nil {
			return found
		}
	}
	return nil
}
