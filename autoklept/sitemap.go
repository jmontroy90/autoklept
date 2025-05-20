package autoklept

import (
	"encoding/xml"
	"fmt"
	"net/url"
)

// TODO: We could parse these and recursively get all sub-sitemaps and their URLs
type sitemapIndex struct {
	Sitemaps []sitemap `xml:"sitemap"`
}

type sitemap struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod"`
}

type urlSet struct {
	URLs []singleURL `xml:"url"`
}

type singleURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod"`
}

func extractUrlSet(raw []byte) ([]url.URL, error) {
	var us urlSet
	if err := xml.Unmarshal(raw, &us); err != nil {
		return nil, fmt.Errorf("error unmarshaling sitemap: %w", err)
	}
	var urls []url.URL
	for _, u := range us.URLs {
		up, err := url.Parse(u.Loc)
		if err != nil {
			return nil, fmt.Errorf("error parsing sitemap url %s: %w", u.Loc, err)
		}
		urls = append(urls, *up)
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
