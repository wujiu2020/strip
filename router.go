package strip

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Handler interface{}

type Handlers []Handler

type nameParam string

func Name(name string) Handler {
	return nameParam(name)
}

type routeParams Handlers

func Router(path string, handlers ...Handler) Handler {
	path = strings.TrimSpace(path)
	if path == "" || path == "/" {
		panic("path can not be `" + path + "`")
	}
	return append(routeParams{path}, handlers...)
}

type RouteInfo struct {
	url.Values
	Keys []string
	Path string
}

type paramList []map[string]string

func (d paramList) values() ([]string, url.Values) {
	keys := make([]string, 0, len(d))
	values := make(url.Values, len(d))
	for _, m := range d {
		for key, value := range m {
			keys = append(keys, key)
			values.Set(key, value)
		}
	}
	return keys, values
}

type pathParam struct {
	path string

	isParam    bool
	isWild     bool
	paramName  string
	customVerb string
}

func (d *pathParam) set(p string) {
	d.path = p

	var paramName string
	if p[0] == ':' {
		d.isParam = true
		paramName = p[1:]
	}

	if p[0] == '*' && len(p) > 1 && p[1] == ':' {
		d.isWild = true
		paramName = p[2:]
	}

	suffix := paramName
	if suffix == "" {
		suffix = p
	}

	if i := strings.Index(suffix, ":"); i != -1 {
		d.customVerb = suffix[i:]
		paramName = suffix[:i]
	}

	if d.isParam || d.isWild {
		d.paramName = paramName
	}
}

func (d *pathParam) matchParamRoute(p string) (value string, ok bool) {
	if !d.isParam {
		return
	}
	if d.customVerb != "" {
		value = strings.TrimSuffix(p, d.customVerb)
		ok = len(value) != len(p)
		return
	}
	if d.isParam {
		value = p
		ok = true
	}
	return
}

type routeRoot struct {
	*route

	notFoundFilters filters

	namedRoutes map[string]*route
}

func newRouteRoot() *routeRoot {
	routeRoot := new(routeRoot)
	routeRoot.route = newRoute(routeRoot, nil)
	routeRoot.isEnd = true
	return routeRoot
}

func (r *routeRoot) notFound(handlers ...interface{}) {
	r.notFoundFilters = makeFilters(handlers)
}

func (r *routeRoot) handle(ctx Context, rw http.ResponseWriter, req *http.Request) {
	// https://tools.ietf.org/html/rfc7231#section-4.1
	// By convention, standardized methods are defined in all-uppercase US-ASCII letters.
	req.Method = strings.ToUpper(req.Method)

	var (
		method = method(req.Method)
		params = make(paramList, 0)
		paths  = splitRoutePath(req.URL.Path)

		// root route
		route = r.route
	)

	if paths[0] != "" {
		// non root route go deep match
		route = r.match(paths, &params)
		if route != nil && !route.isEnd {
			route = nil
		}
	}

	// deal with not found
	if route == nil {
		ctx := newNestContext(ctx, rw.(ResponseWriter), r.notFoundFilters, nil)
		ctx.run()
		return
	}

	routeAction := route.action[method]
	actionFunc := method.action()

	if routeAction != nil {
		actionFunc = routeAction.action
	}

	if routeAction == nil && route.allRoute != nil {
		allRoute := route.allRoute
		if allRoute.controller.isFunc() ||
			allRoute.controller.actionFuncExists(actionFunc) {
			routeAction = allRoute

		} else if allRoute.action != "" {
			routeAction = allRoute
			actionFunc = allRoute.action
		}
	}

	if routeAction == nil && route.anyRoute != nil {
		anyRoute := route.anyRoute
		if anyRoute.controller.isFunc() ||
			anyRoute.controller.actionFuncExists(anyRoute.action) {
			routeAction = anyRoute
			actionFunc = anyRoute.action
		}
	}

	// deal with not found
	if routeAction == nil {
		ctx := newNestContext(ctx, rw.(ResponseWriter), r.notFoundFilters, nil)
		ctx.run()
		return
	}

	info := &RouteInfo{
		Path: route.calcPath(),
	}
	info.Keys, info.Values = params.values()
	ctx.Provide(info)

	// handle target route action
	nestCtx := newNestContext(ctx, rw.(ResponseWriter),
		routeAction.filters,
		routeAction.wrapHandle(route, params, actionFunc))
	nestCtx.run()
}

type route struct {
	pathParam pathParam
	isEnd     bool

	action map[method]*routerAction

	pathRoutes map[string]*route

	paramRoutes []*route
	wildRoute   *route

	root   *routeRoot
	parent *route

	allRoute *routerAction
	anyRoute *routerAction
}

func newRoute(routeRoot *routeRoot, parent *route) *route {
	return &route{
		root:        routeRoot,
		parent:      parent,
		action:      make(map[method]*routerAction),
		pathRoutes:  make(map[string]*route),
		paramRoutes: make([]*route, 0),
	}
}

func (r *route) match(nextPaths []string, params *paramList) (rt *route) {
	p := nextPaths[0]
	paths := nextPaths[1:]
	more := len(paths) > 0

	// first match pathRoutes
	if rt = r.pathRoutes[p]; rt != nil {
		if more {
			newParams := make(paramList, 0, len(paths))
			rt = rt.match(paths, &newParams)
			if rt != nil && rt.isEnd {
				*params = append(*params, newParams...)
			}
		}
	}

	// second match paramRoutes
	if rt == nil && len(r.paramRoutes) > 0 {

		for _, pr := range r.paramRoutes {
			value, matched := pr.pathParam.matchParamRoute(p)
			if !matched {
				continue
			}

			if !more && !pr.isEnd {
				continue
			}

			if !more && pr.isEnd {
				rt = pr
				*params = append(*params, map[string]string{pr.pathParam.paramName: value})
				break
			}

			if more {
				newParams := make(paramList, 0, len(paths))
				nr := pr.match(paths, &newParams)
				if nr != nil && nr.isEnd {
					*params = append(*params, map[string]string{pr.pathParam.paramName: value})
					*params = append(*params, newParams...)
					rt = nr
					break
				}
			}
		}

	}

	if rt != nil && rt.isEnd || r.wildRoute == nil {
		return
	}

	if rt = r.wildRoute; rt != nil {
		*params = append(*params, map[string]string{
			rt.pathParam.paramName: strings.Join(nextPaths, "/"),
		})
	}
	return
}

// config current router
func (r *route) configRoutes(args routeArgs) {
	args.filters = args.filters.remove(args.exempts...)

	var (
		allMethod methodParams
		allArgs   routeArgs
		anyMethod methodParams
		anyArgs   routeArgs
	)

	for _, mt := range args.methods {
		mArgs := calcRouterArgs(Handlers(mt))
		method := mArgs.method

		// detect All method
		if method.isAll() {
			allMethod = mt
			allArgs = mArgs
			continue
		}

		// detect Any method
		if method.isAny() {
			anyMethod = mt
			anyArgs = mArgs
			continue
		}

		act := string(mArgs.action)
		if act == "" {
			act = method.action()
		}

		actFilters := args.filters.append(mArgs.filters...).remove(mArgs.exempts...)
		r.action[method] = newRouteAction(mArgs.ctroler, act, actFilters)
	}

	// default use GET route for HEAD method
	if r.action[HEAD] == nil && r.action[GET] != nil {
		r.action[HEAD] = r.action[GET]
	}

	if allMethod != nil {
		act := string(allArgs.action)
		if act == "" && !allArgs.ctroler.isFunc() {
			if allArgs.ctroler.actionFuncExists("All") {
				act = "All"
			}
		}
		if act != "" {
			allArgs.ctroler.ensureHasAction(act)
		}
		actFilters := args.filters.append(allArgs.filters...).remove(allArgs.exempts...)
		r.allRoute = newRouteAction(allArgs.ctroler, act, actFilters)
	}

	if anyMethod != nil {
		act := string(anyArgs.action)
		if act == "" && !anyArgs.ctroler.isFunc() {
			act = "Any"
		}
		if act != "" {
			anyArgs.ctroler.ensureHasAction(act)
		}
		actFilters := args.filters.append(anyArgs.filters...).remove(anyArgs.exempts...)
		r.anyRoute = newRouteAction(anyArgs.ctroler, act, actFilters)
	}

	// loop nested routers
	for _, rtParam := range args.routers {
		rtArgs := calcRouterArgs(Handlers(rtParam))
		pathStr := rtArgs.strings[0]
		paths := splitRoutePath(pathStr)

		if paths[0] == "" {
			continue
		}

		targetRoute := r
		for i, p := range paths {
			var rt *route

			for _, er := range targetRoute.paramRoutes {
				if er.pathParam.path == p {
					rt = er
					break
				}
			}

			if rt == nil && targetRoute.pathRoutes != nil &&
				targetRoute.pathRoutes[p] != nil {
				rt = targetRoute.pathRoutes[p]
			}

			if rt == nil && targetRoute.wildRoute != nil &&
				targetRoute.wildRoute.pathParam.path == p {
				rt = targetRoute.wildRoute
			}

			isNew := rt == nil
			if rt == nil {
				rt = newRoute(targetRoute.root, targetRoute)
				rt.pathParam.set(p)
			}

			if rt.pathParam.isParam {
				setting := rt.pathParam.paramName
				for _, er := range targetRoute.paramRoutes {
					exists := er.pathParam.paramName
					if exists != setting {
						panic(fmt.Sprintf("route param conflict, please change `:%s` to `:%s`", setting, exists))
					}
				}
				if isNew {
					if rt.pathParam.customVerb != "" {
						targetRoute.paramRoutes = append([]*route{rt}, targetRoute.paramRoutes...)
					} else {
						targetRoute.paramRoutes = append(targetRoute.paramRoutes, rt)
					}
				}

			} else if rt.pathParam.isWild {
				if targetRoute.wildRoute != nil {
					exists := targetRoute.wildRoute.pathParam.paramName
					setting := rt.pathParam.paramName
					if exists != setting {
						panic(fmt.Sprintf("route param conflict, please change `:%s` to `:%s`", setting, exists))
					}
				}

				targetRoute.wildRoute = rt
			} else {
				targetRoute.pathRoutes[rt.pathParam.path] = rt
			}

			isEnd := i == len(paths)-1
			if isEnd {
				rt.isEnd = true
			}

			if rt.pathParam.isWild && !isEnd {
				panic(fmt.Sprintf("wild route `%s` must end with route param `:%s`",
					pathStr, rt.pathParam.paramName))
			}

			targetRoute = rt
		}

		rtArgs.filters = args.filters.append(rtArgs.filters...)

		targetRoute.configRoutes(rtArgs)
	}
}

func (r *route) calcPath() string {
	var path string
	rt := r
	for rt != nil {
		if rt.pathParam.path != "" {
			path = "/" + rt.pathParam.path + path
		}
		rt = rt.parent
	}
	if path == "" {
		path = "/"
	}
	return path
}
