package autoklept

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"strings"
)

const (
	DIV_TAG = "div"
)

type ElementNodeFinder struct {
	Tag     string
	AttrKey string
	AttrVal string
}

func parseHtmlByTag(htmlBody string, lookup ElementNodeFinder) (io.Reader, error) {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return nil, fmt.Errorf("error parsing html: %w", err)
	}
	content := findElementNode(doc, lookup)
	var buf bytes.Buffer
	if err = html.Render(&buf, content); err != nil {
		return nil, fmt.Errorf("error rendering html: %w", err)
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
