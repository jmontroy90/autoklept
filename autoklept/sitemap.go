package autoklept

import (
	"encoding/xml"
	"fmt"
	"net/url"
)

// TODO: We could parse these and recursively get all sub-sitemaps and their URLs
type SitemapIndex struct {
	Sitemaps []Sitemap `xml:"sitemap"`
}

type Sitemap struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod"`
}

type URLSet struct {
	URLs []URL `xml:"url"`
}

type URL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod"`
}

func ExtractUrlSet(raw []byte) ([]*url.URL, error) {
	var urlSet URLSet
	if err := xml.Unmarshal(raw, &urlSet); err != nil {
		return nil, fmt.Errorf("error unmarshaling sitemap: %w", err)
	}
	var urls []*url.URL
	for _, us := range urlSet.URLs {
		u, err := url.Parse(us.Loc)
		if err != nil {
			return nil, fmt.Errorf("error parsing sitemap url %s: %w", us.Loc, err)
		}
		urls = append(urls, u)
	}
	return urls, nil
}

type RootDetector struct {
	XMLName xml.Name
}

// TODO: partial unmarshal to determine what kind of sitemap we have
func detectRootElement(raw []byte) (string, error) {
	var detector RootDetector
	err := xml.Unmarshal(raw, &detector)
	if err != nil {
		return "", err
	}
	return detector.XMLName.Local, nil
}
