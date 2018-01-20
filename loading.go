package main

import (
	"github.com/driusan/Gob/net"
	"github.com/driusan/Gob/parser"
)

func loadPage(filename string) (parser.Page, error) {
	u, err := net.ParseURL(filename)
	if err != nil {
		return parser.Page{}, err
	}

	loader := net.DefaultReader{}
	r, err := loader.GetURL(u)
	if err != nil {
		return parser.Page{}, err
	}
	defer r.Close()
	p := parser.LoadPage(r, loader, u)
	p.URL = u
	return p, nil
}
