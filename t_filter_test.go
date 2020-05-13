package strip

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_Filter(t *testing.T) {
	assert := &Assert{T: t}

	sp := New()

	globalFilter := func(req *http.Request, rw http.ResponseWriter) {
		rw.Header().Set(HeaderPoweredBy, "i'm a sppot")
	}

	routerFunc := func(req *http.Request, rw http.ResponseWriter) {
		rw.WriteHeader(http.StatusTeapot)
	}

	// app filter
	sp.Filter(globalFilter)

	sp.Routers(
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

	assert.True(justATeapot(sp, "GET", "/"))
	assert.False(justATeapot(sp, "GET", "/home"))
	assert.True(justATeapot(sp, "GET", "/dash"))
	assert.False(justATeapot(sp, "GET", "/user"))
	assert.True(justATeapot(sp, "GET", "/user/1"))
	assert.True(justATeapot(sp, "GET", "/user/name"))
	assert.False(justATeapot(sp, "GET", "/user/friend"))
	assert.False(justATeapot(sp, "GET", "/user/email"))

	req, _ := http.NewRequest("GET", "/home", nil)
	rec := httptest.NewRecorder()
	sp.ServeHTTP(rec, req)
	assert.True(rec.Header().Get(HeaderPoweredBy) == "i'm a sppot")
}

func Test_Filter_Context(t *testing.T) {
	assert := &Assert{T: t}
	written := false

	actionFilterProvide := func(rw http.ResponseWriter, ctx Context) {
		con := "i'm a sppot"
		ctx.Provide(&con)

		ctx.Next()

		written = rw.(ResponseWriter).Written()
	}

	sp := New()
	sp.Routers(
		Router("/home",
			// for nextContext in action filter
			Filter(actionFilterProvide),
			Get(func(rw http.ResponseWriter, content *string) {
				rw.Write([]byte(*content))
			}),
		),
	)

	assert.True(responseEqual(sp, "GET", "/home", "i'm a sppot"))
	assert.True(written)
}
