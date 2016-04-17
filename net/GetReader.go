package net

import (
	//"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func ParseURL(urlS string) (*url.URL, error) {
	// First check if it's a file
	if _, err := os.Stat(urlS); !os.IsNotExist(err) {
		// get the absolute path for the file, for safety.
		urlContext, err := filepath.Abs(urlS)
		if err != nil {
			return nil, err
		}

		// parse the path as a file:// URL
		curUrl, urlErr := url.Parse("file://" + urlContext)
		if urlErr != nil {
			return nil, err
		}
		return curUrl, nil
	}

	// it wasn't a filename, so just parse it as a plain old
	return url.Parse(urlS)
}

func GetURLReader(u *url.URL) (io.ReadCloser, error) {
	switch u.Scheme {
	case "file":
		if _, err := os.Stat(u.Path); err != nil {
			return nil, err
		}
		return os.Open(u.Path)
	default:
		resp, err := http.Get(u.String())
		if err != nil {
			return nil, err
		}
		return resp.Body, nil
	}
}
