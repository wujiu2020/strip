package static

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/teapots/teapot"
)

func Test_ServeStatic(t *testing.T) {
	assert := &teapot.Assert{T: t}

	tea := teapot.New()
	tea.Filter(ServeFilter("", "testdata"))

	assert.True(routeFound(tea, "GET", "/test.txt", "test\n"))
	assert.True(routeFound(tea, "HEAD", "/test.txt", ""))

	assert.True(routeNotFound(tea, "POST", "/test.txt"))
}

func Test_Prefix(t *testing.T) {
	assert := &teapot.Assert{T: t}

	tea := teapot.New()
	tea.Filter(ServeFilter("/", "testdata"))
	assert.True(routeFound(tea, "GET", "/test.txt", "test\n"))

	tea = teapot.New()
	tea.Filter(ServeFilter("/prefix", "testdata"))
	assert.True(routeFound(tea, "GET", "/prefix/test.txt", "test\n"))
	assert.True(routeFound(tea, "GET", "/prefix////test.txt", "test\n"))

	tea = teapot.New()
	tea.Filter(ServeFilter("/prefix/", "testdata"))
	assert.True(routeFound(tea, "GET", "/prefix/test.txt", "test\n"))
	assert.True(routeFound(tea, "GET", "/prefix////test.txt", "test\n"))

	tea = teapot.New()
	tea.Filter(ServeFilter("prefix", "testdata"))
	assert.True(routeFound(tea, "GET", "/prefix/test.txt", "test\n"))
	assert.True(routeFound(tea, "GET", "/prefix////test.txt", "test\n"))

	assert.True(routeNotFound(tea, "GET", "/test.txt"))
}

func Test_WrongPath(t *testing.T) {
	assert := &teapot.Assert{T: t}

	tea := teapot.New()
	tea.Filter(ServeFilter("", "testdata"))

	assert.True(routeNotFound(tea, "GET", "/../testdata/test.txt"))
}

func routeFound(tea *teapot.Teapot, method, urlStr, body string) bool {
	req, _ := http.NewRequest(method, urlStr, nil)
	rec := httptest.NewRecorder()
	tea.ServeHTTP(rec, req)
	return rec.Code == http.StatusOK && rec.Body.String() == body
}

func routeNotFound(tea *teapot.Teapot, method, urlStr string) bool {
	req, _ := http.NewRequest(method, urlStr, nil)
	rec := httptest.NewRecorder()
	tea.ServeHTTP(rec, req)
	return rec.Code == http.StatusNotFound
}
