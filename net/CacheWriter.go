package net

import (
	"fmt"
	"io"
	"os"
	//	"strings"
	"crypto/sha1"
	//"os/user"
	"net/url"
	"path/filepath"
	//"strconv"
)

type CacheWriter struct {
	Tee     io.Reader
	ToClose io.Closer
}

func (cw *CacheWriter) Read(p []byte) (int, error) {
	return cw.Tee.Read(p)
}
func (cw *CacheWriter) Close() error {
	return cw.ToClose.Close()
}

func escapeString(s string) string {
	//	sep := string(os.PathSeparator)
	// replace os.PathSeparator (forward slash) with ASCII path separator control character
	// (034) so that os.Create can create the filename.
	// This needs to be done instead of creating a directory, because on a web server a portion
	// of a URL might sometimes be a directory, and sometimes be a file depending on context,
	// so it can't directly map to the filesystem.
	return fmt.Sprintf("%x", sha1.Sum([]byte(s)))
	//return strings.Replace(s, sep, "\034", -1)
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
	if dir != "" {
		filename := filepath.Join(dir, escapeString(resource.Path+"?"+resource.RawQuery))
		writer, _ := os.Create(filename)
		return &CacheWriter{
			Tee:     io.TeeReader(source, writer),
			ToClose: writer,
		}
	}
	return source
}
