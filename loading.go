package main

import (
	"github.com/driusan/Gob/net"
)

func loadPage(filename string) (*Page, error) {
	u, err := net.ParseURL(filename)
	if err != nil {
		return nil, err
	}
	r, err := net.GetURLReader(u)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	p := loadHTML(r, u)
	p.URL = u
	return p, nil
}
