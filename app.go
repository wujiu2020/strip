package strip

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/wujiu2020/strip/inject"

	"github.com/wujiu2020/strip/utils"
)

const HeaderPoweredBy = "X-Powered-By"

type Strip struct {
	route *routeRoot

	// global filters
	filters filters

	// teapot app logger
	logger LoggerAdv

	// master inject of current app
	inject inject.Injector

	// app server
	Server *http.Server

	// app config
	Config *Config
}

var _ inject.TypeProvider = new(Strip)

func New() *Strip {
	strip := &Strip{
		route:  newRouteRoot(),
		inject: inject.New(),
		Config: newConfig(),
	}

	strip.Server = &http.Server{
		Handler: strip,
	}

	strip.Provide(strip.Config)

	log := NewLogger(log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds))
	log.SetColorMode(true)
	strip.SetLogger(log)

	strip.NotFound(defaultNotFound)
	return strip
}

func (s *Strip) ImportConfig(c Configer) {
	s.Config.setParent(c)
	s.Config.Bind(&s.Config.RunPath, "run_path")
	s.Config.Bind(&s.Config.RunMode, "run_mode")
	s.Config.Bind(&s.Config.HttpAddr, "http_addr")
	s.Config.Bind(&s.Config.HttpPort, "http_port")
}

func (s *Strip) NotFound(handlers ...interface{}) {
	s.route.notFound(handlers...)
}

func (s *Strip) Logger() Logger {
	return s.logger
}

func (s *Strip) SetLogger(logger LoggerAdv) {
	s.logger = logger
	s.ProvideAs(logger, (*Logger)(nil))
}

func (s *Strip) Injector() inject.Injector {
	return s.inject
}

func (s *Strip) Provide(provs ...interface{}) inject.TypeProvider {
	return s.inject.Provide(provs...)
}

func (s *Strip) ProvideAs(prov interface{}, typ interface{}, names ...string) inject.TypeProvider {
	return s.inject.ProvideAs(prov, typ, names...)
}

func (s *Strip) Filter(handlers ...interface{}) {
	s.filters = s.filters.append(makeFilters(handlers)...)
}

func (s *Strip) Routers(handlers ...Handler) *Strip {
	args := calcRouterArgs(handlers)

	s.route.configRoutes(args)
	return s
}

func (s *Strip) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// server info
	rw.Header().Set(HeaderPoweredBy, "Teapot")

	// wrap http.ResponseWriter
	trw := newResponseWriter(rw)

	var ctx Context
	ctx = newContext(nil, trw, s.filters, s.route.handle)

	goctx := utils.CtxWithValue(req.Context(), &ctx)
	req = req.WithContext(goctx)
	ctx.ReplaceContext(goctx)

	ctx.SetParent(s.inject)
	ctx.Provide(req)
	ctx.ProvideAs(trw, (*http.ResponseWriter)(nil))

	ctx.(*context).run()

	// flush header if have not written
	trw.Write(nil)
}

func (s *Strip) Run() error {
	mode := string(s.Config.RunMode)
	addr := fmt.Sprintf("%s:%s", s.Config.HttpAddr, s.Config.HttpPort)

	s.Server.Addr = addr

	if !s.Config.RunMode.IsDev() {
		s.logger.SetColorMode(false)
	} else {
		addr = newBrush("32")(s.Server.Addr)
		mode = newBrush("32")(mode)
	}

	s.logger.Infof("Teapot listening on %s in [%s] mode", addr, mode)
	err := s.Server.ListenAndServe()
	s.logger.Emergency(err)
	return err
}
