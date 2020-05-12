package static

import (
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/wujiu2020/strip"
)

type StaticOption struct {
	AnyMethod    bool
	TryFilesThen string
	SkipPrefix   []string
}

// ServeFilter returns a middleware filter that serves static files in the given directory.
func ServeFilter(prefix, directory string, opts ...StaticOption) interface{} {
	var opt StaticOption
	if len(opts) > 0 {
		opt = opts[0]
	}

	directory, _ = filepath.Abs(directory)

	if directory == "" {
		panic("directory can not be empty")
	}

	dir := http.Dir(directory)

	// Normalize the prefix if provided
	if prefix != "" {
		// Ensure we have a leading '/'
		if prefix[0] != '/' {
			prefix = "/" + prefix
		}
		// Remove any trailing '/'
		prefix = strings.TrimRight(prefix, "/")
	}

	serveFile := func(file string) (f http.File, fi os.FileInfo, err error) {
		f, err = dir.Open(file)
		if err != nil {
			return
		}

		defer func() {
			if err != nil {
				f.Close()
			}
		}()

		fi, err = f.Stat()
		if err != nil {
			return
		}
		return
	}

	return func(rw http.ResponseWriter, req *http.Request, log strip.Logger) {
		if !opt.AnyMethod && req.Method != "GET" && req.Method != "HEAD" {
			return
		}

		var (
			err  error
			file = req.URL.Path
		)

		defer func() {
			if err != nil {
				log.Noticef("[STATIC] %s, %v", file, err)
				http.NotFound(rw, req)
				return
			}
		}()

		// if we have a prefix, filter requests by stripping the prefix
		if prefix != "" {
			if !strings.HasPrefix(file, prefix) {
				return
			}
			file = file[len(prefix):]
			if file != "" && file[0] != '/' {
				return
			}
		}

		for _, p := range opt.SkipPrefix {
			if strings.HasPrefix(file, p) {
				return
			}
		}

		f, fi, err := serveFile(file)
		if err != nil {
			goto try
		}
		defer f.Close()

		// try to serve index file
		if fi.IsDir() {
			// redirect if missing trailing slash
			if !strings.HasSuffix(req.URL.Path, "/") {
				dest := url.URL{
					Path:     req.URL.Path + "/",
					RawQuery: req.URL.RawQuery,
					Fragment: req.URL.Fragment,
				}
				http.Redirect(rw, req, dest.String(), http.StatusFound)
				return
			}

			file = path.Join(file, "index.html")
			f, err = dir.Open(file)
			if err != nil {
				return
			}
			defer f.Close()

			fi, err = f.Stat()
			if err != nil || fi.IsDir() {
				return
			}
		}

	try:
		if err != nil {
			if opt.TryFilesThen != "" {
				var er error
				f, fi, er = serveFile(opt.TryFilesThen)
				if er != nil {
					log.Warnf("[STATIC] %s, %v", file, err, er)
					err = er
					return
				}
				err = nil
			}
			if err != nil {
				return
			}
		}

		http.ServeContent(rw, req, file, fi.ModTime(), f)
	}
}
