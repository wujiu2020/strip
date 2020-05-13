package static

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wujiu2020/strip"
)

func Test_ServeStatic(t *testing.T) {
	assert := &strip.Assert{T: t}

	tea := strip.New()
	tea.Filter(ServeFilter("", "testdata"))

	assert.True(routeFound(tea, "GET", "/test.txt", "test\n"))
	assert.True(routeFound(tea, "HEAD", "/test.txt", ""))

	assert.True(routeNotFound(tea, "POST", "/test.txt"))
}

func Test_Prefix(t *testing.T) {
	assert := &strip.Assert{T: t}

	tea := strip.New()
	tea.Filter(ServeFilter("/", "testdata"))
	assert.True(routeFound(tea, "GET", "/test.txt", "test\n"))

	tea = strip.New()
	tea.Filter(ServeFilter("/prefix", "testdata"))
	assert.True(routeFound(tea, "GET", "/prefix/test.txt", "test\n"))
	assert.True(routeFound(tea, "GET", "/prefix////test.txt", "test\n"))

	tea = strip.New()
	tea.Filter(ServeFilter("/prefix/", "testdata"))
	assert.True(routeFound(tea, "GET", "/prefix/test.txt", "test\n"))
	assert.True(routeFound(tea, "GET", "/prefix////test.txt", "test\n"))

	tea = strip.New()
	tea.Filter(ServeFilter("prefix", "testdata"))
	assert.True(routeFound(tea, "GET", "/prefix/test.txt", "test\n"))
	assert.True(routeFound(tea, "GET", "/prefix////test.txt", "test\n"))

	assert.True(routeNotFound(tea, "GET", "/test.txt"))
}

func Test_WrongPath(t *testing.T) {
	assert := &strip.Assert{T: t}

	tea := strip.New()
	tea.Filter(ServeFilter("", "testdata"))

	assert.True(routeNotFound(tea, "GET", "/../testdata/test.txt"))
}

func routeFound(sp *strip.Strip, method, urlStr, body string) bool {
	req, _ := http.NewRequest(method, urlStr, nil)
	rec := httptest.NewRecorder()
	sp.ServeHTTP(rec, req)
	return rec.Code == http.StatusOK && rec.Body.String() == body
}

func routeNotFound(sp *strip.Strip, method, urlStr string) bool {
	req, _ := http.NewRequest(method, urlStr, nil)
	rec := httptest.NewRecorder()
	sp.ServeHTTP(rec, req)
	return rec.Code == http.StatusNotFound
}
