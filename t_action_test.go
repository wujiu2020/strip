package strip

import (
	"fmt"
	"net/http"
	"testing"
)

type TestFunc struct {
	name string
}

func (t TestFunc) Func(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte(t.name))
}

func (t *TestFunc) FuncPtr(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte(t.name))
}

type TestStruct struct {
	Name string
	Req  *http.Request       `inject`
	Rw   http.ResponseWriter `inject`
}

func (t *TestStruct) ChangeName() {
	name := t.Name
	t.Name = "changed"
	t.Rw.Write([]byte(name + t.Name))
}

func (t TestStruct) Struct() {
	t.Rw.Write([]byte(t.Req.URL.Path))
}

func (t *TestStruct) StructPtr() {
	t.Rw.Write([]byte(t.Req.URL.Path))
}

func (t *TestStruct) Param(id int, name string) {
	t.Rw.Write([]byte(fmt.Sprintf("%d%s", id, name)))
}

func (t *TestStruct) ParamId(id int) {
	t.Rw.Write([]byte(fmt.Sprintf("%d", id)))
}

func (t *TestStruct) ParamTwo(a, b string) {
	t.Rw.Write([]byte(a + b))
}

func (t *TestStruct) ParamThree(a, b, c string) {
	t.Rw.Write([]byte(a + b + c))
}

func (t *TestStruct) Get() {
	t.Rw.Write([]byte("Get" + t.Req.Method))
}

func (t *TestStruct) Head() {
	t.Rw.Write([]byte("Head" + t.Req.Method))
}

func (t *TestStruct) Custom() {
	t.Rw.Write([]byte("Custom" + t.Req.Method))
}

type TestAllAnyStruct struct {
	TestStruct
}

func (t *TestAllAnyStruct) All() {
	t.Rw.Write([]byte("All" + t.Req.Method))
}

func (t *TestAllAnyStruct) Any() {
	t.Rw.Write([]byte("Any" + t.Req.Method))
}

func (t *TestAllAnyStruct) Theany() {
	t.Rw.Write([]byte("Theany" + t.Req.Method))
}

func Test_FuncAction(t *testing.T) {
	assert := &Assert{T: t}

	path := ""
	pathFunc := func(req *http.Request) {
		path = req.URL.Path
	}
	sp := New().Routers(
		Get(nopFunc),
		Router("/home", Get(pathFunc)),
		Router("/func1", Get((&TestFunc{name: "Yeap"}).Func)),
		Router("/func2", Get((&TestFunc{name: "Yeap"}).FuncPtr)),
	)

	assert.True(routeFound(sp, "GET", "/"))
	assert.True(routeFound(sp, "GET", "/home"))
	assert.True(path == "/home")

	for _, p := range []string{"/func1", "/func2"} {
		assert.True(responseEqual(sp, "GET", p, "Yeap"))
	}
}

func Test_StructAction(t *testing.T) {
	assert := &Assert{T: t}

	sp := New().Routers(
		Get(nopFunc),
		Router("/route1", Get(&TestStruct{}).Action("Struct")),
		Router("/route2", Get(&TestStruct{}).Action("StructPtr")),
		Router("/route3", Get(TestStruct{}).Action("Struct")),
		Router("/route4", Get(TestStruct{}).Action("StructPtr")),
		Router("/param/:id/:name", Get(TestStruct{}).Action("Param")),
		Router("/param/:id", Put(TestStruct{}).Action("ParamId")),
		Router("/param/:id/name/:type", Any(TestStruct{}).Action("ParamTwo")),
		Router("/param/:id/:name/:target/end", Any(TestStruct{}).Action("ParamThree")),
		Router("/param-all/:id", All(TestStruct{}).Action("Param")),
		Router("/param-any/:id", Any(TestStruct{}).Action("Param")),
	)

	for _, p := range []string{"/route1", "/route2", "route3", "route4"} {
		assert.True(responseEqual(sp, "GET", p, p))
	}

	assert.True(responseEqual(sp, "GET", "/param/xx/sppot", "0sppot"))
	assert.True(responseEqual(sp, "GET", "/param/2147483648/sppot", "2147483648sppot"))
	assert.True(responseEqual(sp, "GET", "/param/100/sppot", "100sppot"))
	assert.True(responseEqual(sp, "PUT", "/param/100", "100"))
	assert.True(responseEqual(sp, "PUT", "/param/100/name/type", "100type"))
	assert.True(responseEqual(sp, "PUT", "/param/100/name/type/end", "100nametype"))
	assert.True(responseEqual(sp, "PUT", "/param-all/100", "100"))
	assert.True(responseEqual(sp, "PUT", "/param-any/100", "100"))
}

func Test_Action_Method(t *testing.T) {
	assert := &Assert{T: t}

	sp := New().Routers(
		Get(nopFunc),
		Router("/route1", Get(&TestStruct{})),
		Router("/route2", Get(&TestStruct{}).Action("Get")),
		Router("/route3", Head(&TestStruct{}).Action("Get")),
		Router("/route4", Head(&TestStruct{})),

		Router("/route/all",
			Put(&TestStruct{}).Action("Struct"),
			All(&TestStruct{}),
		),

		Router("/route/all/action",
			Put(&TestStruct{}).Action("Struct"),
			All(&TestStruct{}).Action("Custom"),
		),

		Router("/route/all/method",
			Put(&TestAllAnyStruct{}).Action("Struct"),
			All(&TestAllAnyStruct{}),
		),

		Router("/route/any",
			Put(&TestAllAnyStruct{}).Action("Struct"),
			Any(&TestAllAnyStruct{}),
		),

		Router("/route/any/action",
			Put(&TestAllAnyStruct{}).Action("Struct"),
			Any(&TestAllAnyStruct{}).Action("Theany"),
		),
	)

	// not match method should 404 not found
	assert.True(routeNotFound(sp, "SOME", "/route1"))
	assert.True(routeNotFound(sp, "SOME", "/route/all"))

	// lowwer-case method shoud support
	assert.True(responseEqual(sp, "get", "/route1", "GetGET"))

	assert.True(responseEqual(sp, "GET", "/route1", "GetGET"))
	assert.True(responseEqual(sp, "GET", "/route2", "GetGET"))
	assert.True(responseEqual(sp, "HEAD", "/route3", "GetHEAD"))
	assert.True(responseEqual(sp, "HEAD", "/route4", "HeadHEAD"))

	assert.True(responseEqual(sp, "PUT", "/route/all", "/route/all"))
	assert.True(responseEqual(sp, "GET", "/route/all", "GetGET"))
	assert.True(responseEqual(sp, "HEAD", "/route/all", "HeadHEAD"))
	assert.True(responseEqual(sp, "CUSTOM", "/route/all", "CustomCUSTOM"))
	assert.True(routeNotFound(sp, "POST", "/route/all"))

	assert.True(responseEqual(sp, "POST", "/route/all/method", "AllPOST"))

	assert.True(responseEqual(sp, "PUT", "/route/all/action", "/route/all/action"))
	assert.True(responseEqual(sp, "GET", "/route/all/action", "GetGET"))
	assert.True(responseEqual(sp, "HEAD", "/route/all/action", "HeadHEAD"))
	assert.True(responseEqual(sp, "CUSTOM", "/route/all/action", "CustomCUSTOM"))
	assert.True(responseEqual(sp, "NOTDEFINED", "/route/all/action", "CustomNOTDEFINED"))
	assert.True(responseEqual(sp, "POST", "/route/all/action", "CustomPOST"))

	assert.True(responseEqual(sp, "PUT", "/route/any", "/route/any"))
	assert.True(responseEqual(sp, "GET", "/route/any", "AnyGET"))
	assert.True(responseEqual(sp, "HEAD", "/route/any", "AnyHEAD"))
	assert.True(responseEqual(sp, "CUSTOM", "/route/any", "AnyCUSTOM"))
	assert.True(responseEqual(sp, "NOTDEFINED", "/route/any", "AnyNOTDEFINED"))
	assert.True(responseEqual(sp, "POST", "/route/any", "AnyPOST"))

	assert.True(responseEqual(sp, "PUT", "/route/any/action", "/route/any/action"))
	assert.True(responseEqual(sp, "GET", "/route/any/action", "TheanyGET"))
	assert.True(responseEqual(sp, "HEAD", "/route/any/action", "TheanyHEAD"))
	assert.True(responseEqual(sp, "CUSTOM", "/route/any/action", "TheanyCUSTOM"))
	assert.True(responseEqual(sp, "NOTDEFINED", "/route/any/action", "TheanyNOTDEFINED"))
	assert.True(responseEqual(sp, "POST", "/route/any/action", "TheanyPOST"))
}

func Test_StructValue(t *testing.T) {

	assert := &Assert{T: t}

	router := &TestStruct{Name: "name1"}

	sp := New().Routers(
		Put(router).Action("ChangeName"),
	)

	assert.True(responseEqual(sp, "PUT", "/", "name1changed"))
	assert.True(router.Name == "name1")

	router.Name = "name2"
	assert.True(responseEqual(sp, "PUT", "/", "name2changed"))
}
