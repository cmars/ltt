package main

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/SlyMarbo/rss"
)

type Download struct {
	rss.Item

	URL url.URL
}

type Library struct {
	*bolt.DB

	Path string
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
			log.Printf("don't know how to download %q: %v", item.ID, err)
		} else {
			available = append(available, download)
		}
	}

	lib, err := NewLibrary(defaultPath())
	if err != nil {
		panic(err)
	}
	defer lib.Close()

	for _, dl := range available {
		if !lib.Contains(dl) {
			err := lib.Archive(dl)
			if err != nil {
				log.Println("failed to archive %q: %v", dl.ID, err)
			}
		}
	}
}

func NewLibrary(path string) (*Library, error) {
	dbpath := filepath.Join(path, ".history")
	db, err := bolt.Open(dbpath, 0600, nil)
	if err != nil {
		return nil, err
	}
	return &Library{
		DB:   db,
		Path: path,
	}, nil
}

func (l *Library) Archive(dl *Download) error {
	return l.Update(func(tx *bolt.Tx) error {
		cmd := exec.Command("youtube-dl", "-x", "--audio-format", "mp3", dl.URL.String())
		cmd.Dir = l.Path
		err := cmd.Run()
		if err != nil {
			return err
		}

		b, err := tx.CreateBucketIfNotExists([]byte("downloaded"))
		if err != nil {
			return err
		}
		if b.Get([]byte(dl.ID)) != nil {
			return fmt.Errorf("already downloaded %q", dl.ID)
		}

		err = b.Put([]byte(dl.ID, []byte(dl.URL.String())))
		if err != nil {
			return err
		}

		return nil
	})
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
	if !isSupportedURL(u) {
		return nil, fmt.Errorf("unsupported download URL: %q", u)
	}

	return &Download{
		Item: *item,
		Info: songInfo,
		URL:  *u,
	}, nil
}

func isSupportedURL(u *url.URL) bool {
	s := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) {
			return r
		}
		return rune(-1)
	}, u.Host)
	if strings.Contains(s, "youtube") {
		return true
	}
	return false
}
