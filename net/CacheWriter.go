package net

import (
	//	"fmt"
	"io"
	"os"
	"strings"
	//"os/user"
	"net/url"
	"path/filepath"
	//"strconv"
)

type CacheWriter struct {
	Src   io.ReadCloser
	Cache io.WriteCloser
}

func (cw *CacheWriter) Read(p []byte) (int, error) {
	n, err := cw.Src.Read(p)
	if cw.Cache != nil {
		cw.Cache.Write(p)
	}
	return n, err
}
func (cw *CacheWriter) Close() error {
	if cw.Cache != nil {
		cw.Cache.Close()
	}

	return cw.Src.Close()
}

func escapeString(s string) string {
	sep := string(os.PathSeparator)
	// replace os.PathSeparator (forward slash) with ASCII path separator control character
	// (034) so that os.Create can create the filename.
	// This needs to be done instead of creating a directory, because on a web server a portion
	// of a URL might sometimes be a directory, and sometimes be a file depending on context,
	// so it can't directly map to the filesystem.
	return strings.Replace(s, sep, "\034", -1)
}
func GetCacheWriter(source io.ReadCloser, cachedir string, resource *url.URL) io.ReadCloser {
	var dir string
	if resource.Scheme != "" && resource.Host != "" && cachedir != "" {
		dir = filepath.Join(cachedir, resource.Scheme, resource.Host)
		err := os.MkdirAll(dir, 0700)
		if err != nil {
			return source
		}

	}
	var writer io.WriteCloser
	if dir != "" {
		filename := filepath.Join(dir, escapeString(resource.Path)+"?"+escapeString(resource.RawQuery))
		writer, _ = os.Create(filename)
	}
	return &CacheWriter{
		Src:   source,
		Cache: writer,
	}
}
