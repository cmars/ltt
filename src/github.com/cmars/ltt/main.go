package main

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/SlyMarbo/rss"
)

type Download struct {
	rss.Item

	URL url.URL
}

func main() {
	feed, err := rss.Fetch("https://www.reddit.com/r/listentothis/.rss")
	if err != nil {
		panic(err)
	}

	var available []*Download
	for _, item := range feed.Items {
		download, err := ParseDownload(item)
		if err != nil {
			log.Println("don't know how to download %q: %v", item.ID, err)
		}
		available = append(available, download)
	}

	for _, dl := range available {
		fmt.Println(dl.URL)
	}
}

func ParseDownload(item *rss.Item) (*Download, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(item.Content))
	if err != nil {
		return nil, err
	}
	a := doc.Find("a:contains('[link]')")
	if a.Length() == 0 {
		return nil, fmt.Errorf("download link not found")
	}
	href, ok := a.Attr("href")
	if !ok {
		return nil, fmt.Errorf("missing expected 'href' attribute in element")
	}
	u, err := url.Parse(href)
	if err != nil {
		return nil, err
	}
	if !isYouTubeURL(u) {
		return nil, fmt.Errorf("unsupported download URL: %q", u)
	}

	return &Download{
		Item: *item,
		URL:  *u,
	}, nil
}
