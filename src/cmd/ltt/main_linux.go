package main

import "os"

func init() {
	// Force C DNS resolver because pure Go one is buggy for me.
	// See:
	//	https://github.com/golang/go/issues/6464
	//  https://golang.org/pkg/net/
	if os.Getenv("LOCALDOMAIN") == "" {
		os.Setenv("LOCALDOMAIN", "")
	}
}
