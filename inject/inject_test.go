package inject

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

type Logger interface {
	Info(string)
}

type Log struct {
	req *http.Request
}

func (log *Log) Info(msg string) {}

type Base struct {
	Log    *Log    `inject`
	Logger Logger  `inject`
	Single *Single `inject`
}

type Child struct {
	Base
	Req     *http.Request `inject`
	ReqPost *http.Request `inject:"post"`
}

type Single struct {
	Count int
}

func CreateProvide() Injector {
	NewLogger := func() Provider {
		return func(req *http.Request) *Log {
			l := new(Log)
			l.req = req
			return l
		}
	}

	Singleton := func() Provider {
		single := new(Single)
		return func() *Single {
			single.Count += 1
			return single
		}
	}

	NilProv := func() interface{} {
		return func() *io.LimitedReader {
			return nil
		}
	}

	inj := New()

	// function
	inj.Provide(NewLogger())
	inj.Provide(Singleton())

	// Value
	req, _ := http.NewRequest("GET", "http://localhost/", nil)
	inj.Provide(req)
	inj.Provide(NilProv())

	// Object
	inj.Provide(Object{Value: NewLogger()})
	inj.Provide(&Object{Value: NewLogger()})

	inj.ProvideAs(NewLogger(), (*Logger)(nil))

	req, _ = http.NewRequest("POST", "http://localhost/", nil)
	inj.Provide(Object{Value: req, Name: "post"})

	return inj
}

func Test_Provide(t *testing.T) {
	assert := &Assert{T: t}

	inj := CreateProvide()
	inject := inj.(*injector)

	assert.True(len(inject.values) == 6)
}

func Test_Find(t *testing.T) {
	assert := &Assert{T: t}

	inj := CreateProvide()

	var req *http.Request
	assert.NoError(inj.Find(&req, ""))
	assert.NotNil(req)

	req = nil
	assert.NoError(inj.Find(&req, "post"))
	assert.NotNil(req)

	var log Logger
	assert.NoError(inj.Find(&log, ""))
	assert.NotNil(log)

	var lrp *io.LimitedReader
	assert.NoError(inj.Find(&lrp, ""))
	assert.True(lrp == nil)

	var lr io.LimitedReader
	assert.NoError(inj.Find(&lr, ""))
}

func Test_FindInterface(t *testing.T) {
	assert := &Assert{T: t}

	var log Logger
	log = new(Log)

	inj := New()
	inj.Provide(log)

	var findLog Logger
	assert.Error(inj.Find(&findLog, ""))
	assert.True(findLog == nil)

	inj = New()
	inj.Provide(&log)

	assert.NoError(inj.Find(&findLog, ""))
	assert.NotNil(findLog)
}

func Test_Invoke(t *testing.T) {
	assert := &Assert{T: t}

	inj := CreateProvide()

	// func
	exec := false
	funcA := func(req *http.Request, log Logger) {
		assert.NotNil(req)
		assert.NotNil(log)
		exec = true
	}
	_, err := inj.Invoke(funcA)
	assert.NoError(err)
	assert.True(exec)

	// func with return values
	funcB := func(req *http.Request, log Logger) string {
		assert.NotNil(req)
		assert.NotNil(log)
		return "name"
	}
	out, err := inj.Invoke(funcB)
	assert.NoError(err)
	assert.True(len(out) == 1)
	assert.True(out[0].Kind() == reflect.String)
	assert.True(out[0].String() == "name")

	// func with deps and return values
	funcC := Provide{
		Dep{0: "post"},
		func(req *http.Request, log Logger) string {
			assert.NotNil(req)
			assert.NotNil(log)
			return "name"
		},
	}
	out, err = inj.Invoke(funcC)
	assert.NoError(err)
	assert.True(len(out) == 1)
	assert.True(out[0].Kind() == reflect.String)
	assert.True(out[0].String() == "name")
}

func Test_Apply(t *testing.T) {
	assert := &Assert{T: t}

	inj := CreateProvide()
	child := &Child{}
	err := inj.Apply(child)
	assert.NoError(err)
	assert.NotNil(child.Log)
	assert.NotNil(child.Logger)
	assert.NotNil(child.Req)
	assert.NotNil(child.ReqPost)
	assert.NotNil(child.Single)
	assert.True(child.ReqPost.Method == "POST")
}

type Service struct{}

func Test_Parent(t *testing.T) {
	assert := &Assert{T: t}

	parent := New()
	parent.Provide(func() *Service {
		return new(Service)
	})

	inj := CreateProvide()
	inj.SetParent(parent)

	var service *Service
	funcA := func(s *Service) {
		service = s
	}
	_, err := inj.Invoke(funcA)
	assert.NoError(err)
	assert.NotNil(service)
}

func Test_InvokeCache(t *testing.T) {
	assert := &Assert{T: t}

	inj := CreateProvide()
	New().SetParent(inj).Apply(new(Child))
	New().SetParent(inj).Apply(new(Child))
	New().SetParent(inj).Apply(new(Child))
	New().SetParent(inj).Apply(new(Child))
	New().SetParent(inj).Apply(new(Child))

	var single *Single
	err := inj.Find(&single, "")
	assert.NoError(err)
	assert.NotNil(single)
	assert.True(single.Count == 6)
}

func Test_CycleDependencies(t *testing.T) {
	assert := &Assert{T: t}

	inj := New()
	inj.Provide(func(log Logger) *Service {
		return new(Service)
	})
	inj.Provide(func(s *Service) Logger {
		return new(Log)
	})

	_, err := inj.Invoke(func(log Logger) {})
	assert.True(err != nil)
	assert.True(strings.Index(err.Error(), "cycle dependencies") != -1)

	obj := &struct {
		Log Logger `inject`
	}{}
	inj.Apply(obj)
	assert.True(err != nil)
	assert.True(strings.Index(err.Error(), "cycle dependencies") != -1)
}

type Assert struct {
	T *testing.T
}

func (t *Assert) Error(err error) {
	if err == nil {
		t.T.Errorf("expected error but get nil", CallerInfo())
		t.T.FailNow()
	}
}

func (t *Assert) NoError(err error) {
	if err != nil {
		t.T.Errorf("expected no error but get `%v`\n%s", err, CallerInfo())
		t.T.FailNow()
	}
}

func (t *Assert) NotNil(ni interface{}) {
	val := reflect.ValueOf(ni)
	kind := val.Kind()
	if ni == nil || kind >= reflect.Chan && kind <= reflect.Slice && val.IsNil() {
		t.T.Errorf("expected not nil\n%s", CallerInfo())
		t.T.FailNow()
	}
}

func (t *Assert) True(b bool) {
	if !b {
		t.T.Errorf("expected true but get false\n%s", CallerInfo())
		t.T.FailNow()
	}
}

func CallerInfo() string {
	file := ""
	line := 0
	ok := false

	for i := 0; i < 3; i++ {
		_, file, line, ok = runtime.Caller(i)
		if !ok {
			return ""
		}
		parts := strings.Split(file, "/")
		file = parts[len(parts)-1]
	}

	return fmt.Sprintf("%s:%d", file, line)
}
