package strip

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_Filter(t *testing.T) {
	assert := &Assert{T: t}

	tea := New()

	globalFilter := func(req *http.Request, rw http.ResponseWriter) {
		rw.Header().Set(HeaderPoweredBy, "i'm a teapot")
	}

	routerFunc := func(req *http.Request, rw http.ResponseWriter) {
		rw.WriteHeader(http.StatusTeapot)
	}

	// app filter
	tea.Filter(globalFilter)

	tea.Routers(
		// app filter cannot be exempt
		// because it execute before route match
		Exempt(globalFilter),

		// global router filter
		Filter(routerFunc),

		Get(nopFunc),

		Router("/home",
			Exempt(routerFunc),
			Get(nopFunc),
		),

		Router("/dash",
			Get(nopFunc),
		),

		Router("/user",
			Get(nopFunc).Exempt(routerFunc),

			Router("/name",
				Filter(routerFunc),
				Get(nopFunc),
			),

			Router("/age",
				Get(nopFunc).Filter(routerFunc),
			),

			Router("/friend",
				Exempt(routerFunc),
				Get(nopFunc),
			),

			Router("/email",
				Get(nopFunc).Exempt(routerFunc),
			),
		),

		Router("/user/:uid",
			Get(nopFunc),
		),
	)

	assert.True(justATeapot(tea, "GET", "/"))
	assert.False(justATeapot(tea, "GET", "/home"))
	assert.True(justATeapot(tea, "GET", "/dash"))
	assert.False(justATeapot(tea, "GET", "/user"))
	assert.True(justATeapot(tea, "GET", "/user/1"))
	assert.True(justATeapot(tea, "GET", "/user/name"))
	assert.False(justATeapot(tea, "GET", "/user/friend"))
	assert.False(justATeapot(tea, "GET", "/user/email"))

	req, _ := http.NewRequest("GET", "/home", nil)
	rec := httptest.NewRecorder()
	tea.ServeHTTP(rec, req)
	assert.True(rec.Header().Get(HeaderPoweredBy) == "i'm a teapot")
}

func Test_Filter_Context(t *testing.T) {
	assert := &Assert{T: t}
	written := false

	actionFilterProvide := func(rw http.ResponseWriter, ctx Context) {
		con := "i'm a teapot"
		ctx.Provide(&con)

		ctx.Next()

		written = rw.(ResponseWriter).Written()
	}

	tea := New()
	tea.Routers(
		Router("/home",
			// for nextContext in action filter
			Filter(actionFilterProvide),
			Get(func(rw http.ResponseWriter, content *string) {
				rw.Write([]byte(*content))
			}),
		),
	)

	assert.True(responseEqual(tea, "GET", "/home", "i'm a teapot"))
	assert.True(written)
}
