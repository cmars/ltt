package main

import (
	"fmt"

	"github.com/SlyMarbo/rss"
)

func main() {
	feed, err := rss.Fetch("https://www.reddit.com/r/listentothis/.rss")
	if err != nil {
		panic(err)
	}

	for _, item := range feed.Items {
		fmt.Println(item.Link)
	}
}
