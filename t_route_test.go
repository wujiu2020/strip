package strip

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var nopFunc = func() {}

func Test_pathParam(t *testing.T) {
	assert := &Assert{T: t}

	p := pathParam{}
	p.set("*:splat")
	assert.True(p.isWild)
	assert.True(!p.isParam)
	assert.True(p.customVerb == "")
	assert.True(p.paramName == "splat")

	p = pathParam{}
	p.set(":splat")
	assert.True(!p.isWild)
	assert.True(p.isParam)
	assert.True(p.customVerb == "")
	assert.True(p.paramName == "splat")

	p = pathParam{}
	p.set(":uid:undelete")
	assert.True(!p.isWild)
	assert.True(p.isParam)
	assert.True(p.customVerb == ":undelete")
	assert.True(p.paramName == "uid")

	value, matched := p.matchParamRoute("10:undelete")
	assert.True(matched)
	assert.True(value == "10")
}

func Test_NoRoute(t *testing.T) {
	assert := &Assert{T: t}

	sp := New().Routers()
	assert.True(routeNotFound(sp, "GET", "/"))
}

func Test_RootRoute(t *testing.T) {
	assert := &Assert{T: t}

	sp := New().Routers(
		Get(nopFunc),
	)
	assert.True(routeFound(sp, "GET", "/"))
	assert.True(routeFound(sp, "HEAD", "/"))
	assert.True(routeNotFound(sp, "POST", "/"))

	sp = New().Routers(
		Post(nopFunc),
	)
	assert.True(routeNotFound(sp, "GET", "/"))
	assert.True(routeNotFound(sp, "HEAD", "/"))
	assert.True(routeFound(sp, "POST", "/"))

}

func Test_UrlRoute(t *testing.T) {
	assert := &Assert{T: t}

	path := ""
	pathFunc := func(req *http.Request) {
		path = req.URL.Path
	}

	sp := New().Routers(
		// root
		Get(nopFunc),

		Router("/user",
			Get(nopFunc),
			Put(nopFunc),

			Router("/dashboard", Get(pathFunc)),

			Router("/:uid",
				Get(pathFunc),

				Router("/dashboard", Get(pathFunc)),
			),

			Router("/:uid:undelete",
				Post(pathFunc),
				Router("/pass", Patch(pathFunc)),
			),

			Router("/:uid:move",
				Name("xxx"),
				Post(pathFunc),
			),

			Router("/:uid/name/:name", Get(pathFunc)),
			Router("/:uid/test1/:nopanic/", Get(pathFunc)),

			Router("/:uid/test2/:nopanic/more", Get(pathFunc)),
		),

		Router("/user/dashboard", Put(pathFunc)),
		Router("/user/:uid/dashboard", Put(pathFunc)),
	)

	assert.True(routeFound(sp, "GET", "/"))

	assert.True(routeFound(sp, "GET", "/user"))
	assert.True(routeFound(sp, "PUT", "/user"))

	assert.True(routeFound(sp, "GET", "/user/dashboard"))
	assert.True(path == "/user/dashboard")

	assert.True(routeFound(sp, "GET", "/user/10"))
	assert.True(path == "/user/10")

	assert.True(routeFound(sp, "POST", "/user/10:undelete"))
	assert.True(path == "/user/10:undelete")

	assert.True(routeFound(sp, "PATCH", "/user/10:undelete/pass"))
	assert.True(path == "/user/10:undelete/pass")

	assert.True(routeFound(sp, "POST", "/user/10:move"))
	assert.True(path == "/user/10:move")

	assert.True(routeFound(sp, "GET", "/user/10/dashboard"))
	assert.True(path == "/user/10/dashboard")

	assert.True(routeFound(sp, "GET", "/user/10/name/username"))
	assert.True(path == "/user/10/name/username")

	assert.True(routeNotFound(sp, "GET", "/user/10/name/username/pass"))

	assert.True(routeFound(sp, "GET", "/user/10/test2/nopanic/more"))
	assert.True(path == "/user/10/test2/nopanic/more")

	assert.True(routeNotFound(sp, "GET", "/user/10/test2/nopanic"))

	assert.True(routeFound(sp, "PUT", "/user/dashboard"))
	assert.True(path == "/user/dashboard")

	assert.True(routeFound(sp, "PUT", "/user/10/dashboard"))
	assert.True(path == "/user/10/dashboard")

	assert.True(routeNotFound(sp, "POST", "/user"))
	assert.True(routeNotFound(sp, "POST", "/user/dashboard"))
}

func Test_RouteInfo(t *testing.T) {
	assert := &Assert{T: t}

	var info *RouteInfo
	infoFunc := func(i *RouteInfo) {
		info = i
	}

	sp := New().Routers(
		// root
		Get(infoFunc),

		Router("/user",
			Get(infoFunc),

			Router("/:uid",
				Get(infoFunc),
			),

			Router("/:uid/name/:name",
				Get(infoFunc),
			),

			Router("/:uid/name/:name:customVerb",
				Get(infoFunc),
			),
		),
	)

	assert.True(routeFound(sp, "GET", "/user/101"))
	assert.True(info.Get("uid") == "101")

	assert.True(routeFound(sp, "GET", "/user/101/name/slene"))
	assert.True(info.Get("uid") == "101")
	assert.True(info.Get("name") == "slene")

	assert.True(routeFound(sp, "GET", "/"))
	assert.True(info.Path == "/")

	assert.True(routeFound(sp, "GET", "/user"))
	assert.True(info.Path == "/user")

	assert.True(routeFound(sp, "GET", "/user/101"))
	assert.True(info.Path == "/user/:uid")

	assert.True(routeFound(sp, "GET", "/user/101/name/slene"))
	assert.True(info.Path == "/user/:uid/name/:name")

	assert.True(routeFound(sp, "GET", "/user/101/name/slene:customVerb"))
	assert.True(info.Path == "/user/:uid/name/:name:customVerb")
	assert.True(info.Get("uid") == "101")
	assert.True(info.Get("name") == "slene")
}

func Test_RouteWild(t *testing.T) {
	assert := &Assert{T: t}

	pathFunc := func(rw http.ResponseWriter, req *http.Request, i *RouteInfo) {
		path := i.Get("splat")
		rw.Write([]byte(req.Method + ":" + i.Path + ":" + path))
	}

	sp := New().Routers(
		Get(nopFunc),
		Router("/route/wild",
			Get(pathFunc),
			Router("/*:splat",
				Get(pathFunc),
			),
		),
		Router("/route/wild2",
			Router("/*:splat",
				Get(pathFunc),
			),
		),
		Router("/route/wild2",
			Router("/*:splat",
				Get(pathFunc),
			),
		),
		Router("/route/wild3",
			Router("/user",
				Get(pathFunc),
			),
			Router("/:uid",
				Get(pathFunc),
			),
			Router("/:uid/order",
				Get(pathFunc),
			),
			Get(nopFunc),
			Router("/*:splat",
				Post(pathFunc),
				Get(pathFunc),
			),
		),
		Router("/route/wild4",
			Router("/*:splat", Get(pathFunc), Post(pathFunc)),
			Router("/*:splat", Put(pathFunc)),
		),
	)

	assert.True(responseEqual(sp, "GET", "/route/wild", "GET:/route/wild:"))
	assert.True(responseEqual(sp, "GET", "/route/wild/", "GET:/route/wild:"))
	assert.True(responseEqual(sp, "GET", "/route/wild/1/2",
		"GET:/route/wild/*:splat:1/2",
	))
	assert.True(responseEqual(sp, "GET", "/route/wild2/1/2",
		"GET:/route/wild2/*:splat:1/2",
	))
	assert.True(responseEqual(sp, "POST", "/route/wild3/1/2",
		"POST:/route/wild3/*:splat:1/2",
	))
	assert.True(responseEqual(sp, "GET", "/route/wild3/user",
		"GET:/route/wild3/user:",
	))
	assert.True(responseEqual(sp, "GET", "/route/wild3/1",
		"GET:/route/wild3/:uid:",
	))
	assert.True(responseEqual(sp, "GET", "/route/wild3/1/order",
		"GET:/route/wild3/:uid/order:",
	))
	assert.True(responseEqual(sp, "GET", "/route/wild4/1/get/order",
		"GET:/route/wild4/*:splat:1/get/order",
	))
	assert.True(responseEqual(sp, "POST", "/route/wild4/1/post/order",
		"POST:/route/wild4/*:splat:1/post/order",
	))
	assert.True(responseEqual(sp, "PUT", "/route/wild4/1/put/order",
		"PUT:/route/wild4/*:splat:1/put/order",
	))

	// wild route can not defined in middle of route path
	func() {
		defer func() {
			err := recover()
			assert.NotNil(err)
		}()

		New().Routers(
			Router("/route/wild/*:splat/name",
				Get(nopFunc),
			),
		)
	}()
}

func Test_ConflictParamRoute1(t *testing.T) {
	assert := &Assert{T: t}
	defer func() {
		err := recover()
		if err != nil {
			errStr, _ := err.(string)
			assert.True(strings.Contains(errStr, "conflict"))
			assert.True(strings.Contains(errStr, "`:panic` to `:uid`"))
		}
		assert.NotNil(err)
	}()
	New().Routers(
		Router("/:uid/name/:name", Get(nopFunc)),
		Router("/:panic/name", Get(nopFunc)),
	)
}

func Test_ConflictParamRoute2(t *testing.T) {
	assert := &Assert{T: t}
	defer func() {
		err := recover()
		if err != nil {
			errStr, _ := err.(string)
			assert.True(strings.Contains(errStr, "conflict"))
			assert.True(strings.Contains(errStr, "`:panic` to `:name`"))
		}
		assert.NotNil(err)
	}()
	New().Routers(
		Router("/:uid/name/:name", Get(nopFunc)),
		Router("/:uid/name/:panic/test", Get(nopFunc)),
	)
}

func Test_WildRouteConflict(t *testing.T) {
	assert := &Assert{T: t}
	defer func() {
		err := recover()
		if err != nil {
			errStr, _ := err.(string)
			assert.True(strings.Contains(errStr, "`:splat2` to `:splat`"))
		}
		assert.NotNil(err)
	}()
	New().Routers(
		Router("/:uid/name/:splat", Get(nopFunc)),
		Router("/:uid/name/:splat2", Get(nopFunc)),
	)
}

func Test_WildRouteMustEnd(t *testing.T) {
	assert := &Assert{T: t}
	defer func() {
		err := recover()
		if err != nil {
			errStr, _ := err.(string)
			assert.True(strings.Contains(errStr, "must end with route param"))
		}
		assert.NotNil(err)
	}()
	New().Routers(
		Router("/:uid/name/*:splat/path/x/x/x", Get(nopFunc)),
	)
}

func routeFound(sp *Strip, method, urlStr string) bool {
	req, _ := http.NewRequest(method, urlStr, nil)
	rec := httptest.NewRecorder()
	sp.ServeHTTP(rec, req)
	return rec.Code == http.StatusOK
}

func routeNotFound(sp *Strip, method, urlStr string) bool {
	req, _ := http.NewRequest(method, urlStr, nil)
	rec := httptest.NewRecorder()
	sp.ServeHTTP(rec, req)
	return rec.Code == http.StatusNotFound
}

func responseEqual(sp *Strip, method, urlStr string, resp string) bool {
	req, _ := http.NewRequest(method, urlStr, nil)
	rec := httptest.NewRecorder()
	sp.ServeHTTP(rec, req)
	return rec.Body.String() == resp
}

func justATeapot(sp *Strip, method, urlStr string) bool {
	req, _ := http.NewRequest(method, urlStr, nil)
	rec := httptest.NewRecorder()
	sp.ServeHTTP(rec, req)
	return rec.Code == http.StatusTeapot
}
