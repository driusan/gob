package net

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

var client http.Client

func ParseURL(urlS string) (*url.URL, error) {
	// First check if it's a file
	if _, err := os.Stat(urlS); !os.IsNotExist(err) {
		// get the absolute path for the file, for safety.
		urlContext, err := filepath.Abs(urlS)
		if err != nil {
			return nil, err
		}

		// parse the path as a file:// URL
		curUrl, urlErr := url.Parse("file:" + urlContext)
		if urlErr != nil {
			return nil, err
		}
		return curUrl, nil
	}

	// it wasn't a filename, so just parse it as a plain old
	return url.Parse(urlS)
}

type URLReader interface {
	GetURL(u *url.URL) (body io.ReadCloser, statuscode int, err error)
	HasVisited(u *url.URL) bool
}

type DefaultReader struct{}

func (d DefaultReader) GetURL(u *url.URL) (body io.ReadCloser, statuscode int, err error) {
	switch u.Scheme {
	case "file":
		if _, err := os.Stat(u.Opaque); err != nil {
			return nil, 404, err
		}
		f, err := os.Open(u.Opaque)
		return f, 200, err
	case "data":
		// See RFC 2397
		pieces := strings.SplitN(u.Opaque, ",", 2)
		media := pieces[0]
		if !strings.HasSuffix(media, ";base64") {
			return nil, 500, fmt.Errorf("Only base64 currently supported")
		}
		datareader := strings.NewReader(pieces[1])
		return ioutil.NopCloser(
				base64.NewDecoder(base64.StdEncoding, datareader),
			),
			200,
			nil
	default:
		if cached := getCacheReader(u); cached != nil {
			return cached, 200, nil
		}

		req, _ := http.NewRequest("GET", u.String(), nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537..36 (KHTML, like Gecko) Chrome/39.0.2171.27 Safari/537.36")
		resp, err := client.Do(req)
		//resp, err := http.Get(u.String())
		if err != nil {
			return nil, 400, err
		}
		if resp.StatusCode == 200 {
			// Only cache 200 response codes, because the filesystem doesn't store
			// information about the response code in the cache..
			cw := GetCacheWriter(resp.Body, GetCacheDir(), u)
			return cw, 200, nil
		}
		return resp.Body, resp.StatusCode, nil
	}
}

func (d DefaultReader) HasVisited(u *url.URL) bool {
	l := GetCacheLocation(u)
	if l == "" {
		return false
	}
	_, err := os.Stat(l)
	return !os.IsNotExist(err)
}

func GetCacheDir() string {
	user, err := user.Current()
	if err != nil {
		return ""
	}
	return filepath.Join(user.HomeDir, ".gob", "cache")
}
func GetCacheLocation(resource *url.URL) string {
	cachedir := GetCacheDir()
	dir := filepath.Join(cachedir, resource.Scheme, resource.Host)
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, escapeString(resource.Path+"?"+resource.RawQuery))
}

func getCacheReader(u *url.URL) io.ReadCloser {
	cacheLocation := GetCacheLocation(u)
	if _, err := os.Stat(cacheLocation); !os.IsNotExist(err) {
		o, err := os.Open(cacheLocation)
		if err != nil {
			return nil
		}
		return o
	}
	return nil
}
