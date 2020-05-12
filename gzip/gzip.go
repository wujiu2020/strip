package gzip

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/teapots/teapot"
)

const (
	HeaderAcceptEncoding  = "Accept-Encoding"
	HeaderContentEncoding = "Content-Encoding"
	HeaderContentLength   = "Content-Length"
	HeaderContentType     = "Content-Type"
	HeaderVary            = "Vary"
)

// All returns a Handler that adds gzip compression to all requests
func All() interface{} {
	return func(rw http.ResponseWriter, req *http.Request, ctx teapot.Context) {
		if !strings.Contains(req.Header.Get(HeaderAcceptEncoding), "gzip") {
			return
		}

		headers := rw.Header()
		headers.Set(HeaderContentEncoding, "gzip")
		headers.Set(HeaderVary, HeaderAcceptEncoding)

		gzw := &gzipResponseWriter{ResponseWriter: rw.(teapot.ResponseWriter)}
		defer gzw.close()
		ctx.ProvideAs(gzw, (*http.ResponseWriter)(nil))

		ctx.Next()
		// for defer, gzw need close at the end of request
	}
}

type gzipResponseWriter struct {
	teapot.ResponseWriter
	gzw *gzip.Writer
	one sync.Once
}

func (g *gzipResponseWriter) Write(p []byte) (int, error) {
	g.one.Do(func() {
		g.gzw = gzip.NewWriter(g.ResponseWriter)

		// delete content length after we know we have been written to
		g.Header().Del(HeaderContentLength)

		if len(g.Header().Get(HeaderContentType)) == 0 {
			g.Header().Set(HeaderContentType, http.DetectContentType(p))
		}
	})

	return g.gzw.Write(p)
}

func (g *gzipResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := g.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (g *gzipResponseWriter) close() {
	if g.gzw != nil {
		g.gzw.Close()
	}
}
