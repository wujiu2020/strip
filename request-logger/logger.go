package reqlogger

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/wujiu2020/strip"
	"github.com/wujiu2020/strip/inject"
	"github.com/wujiu2020/strip/utils"
)

const (
	HeaderReqid         = "X-Reqid"
	HeaderResponseTime  = "X-Response-Time"
	HeaderXRealIp       = "X-Real-Ip"
	HeaderXForwardedFor = "X-Forwarded-For"
)

type LoggerContentFilter func(ctx strip.Context, rw http.ResponseWriter, req *http.Request, content string) string

type LoggerOption struct {
	ColorMode     bool
	LineInfo      bool
	ShortLine     bool
	FlatLine      bool
	LogStackLevel strip.Level
	ReqidFilter   LoggerContentFilter
	PrefixFilter  LoggerContentFilter
	ReqBegFilter  LoggerContentFilter
	ReqEndFilter  LoggerContentFilter
}

// ReqLogger filter
func ReqLoggerFilter(out strip.LogPrinter, opts ...LoggerOption) inject.Provider {
	var opt LoggerOption
	if len(opts) > 0 {
		opt = opts[0]
	}
	return func(ctx strip.Context, rw http.ResponseWriter, req *http.Request) {

		// use origin request id or create new request id
		reqId := req.Header.Get(HeaderReqid)
		reqIdLength := len(reqId)
		if reqIdLength < 10 || reqIdLength > 32 {
			reqId = NewReqId()
		}

		if opt.ReqidFilter != nil {
			reqId = opt.ReqidFilter(ctx, rw, req, reqId)
		}

		prefix := ""
		if opt.PrefixFilter != nil {
			prefix = opt.PrefixFilter(ctx, rw, req, prefix)
		}

		log := strip.NewReqLogger(out, reqId)
		log.SetPrefix(prefix)

		// write request id to request and response
		req.Header.Set(HeaderReqid, reqId)
		rw.Header().Set(HeaderReqid, reqId)

		log.SetLineInfo(opt.LineInfo)
		log.SetShortLine(opt.ShortLine)
		log.SetFlatLine(opt.FlatLine)
		log.SetColorMode(opt.ColorMode)
		log.EnableLogStack(opt.LogStackLevel)

		ctx.ProvideAs(log, (*strip.Logger)(nil))
		ctx.ProvideAs(log, (*strip.ReqLogger)(nil))

		req = setLoggerInContext(ctx, req, log)

		remoteAddr := realIp(req)

		start := time.Now()
		reqBeg := fmt.Sprintf("[REQ_BEG] %s %s%s %s", req.Method, req.Host, req.URL, remoteAddr)
		if opt.ReqBegFilter != nil {
			reqBeg = opt.ReqBegFilter(ctx, rw, req, reqBeg)
		}
		if reqBeg != "" {
			log.Info(reqBeg, strip.LineOpt{Hidden: true})
		}

		res := rw.(strip.ResponseWriter)
		res.Before(func(rw strip.ResponseWriter) {
			rw.Header().Del(HeaderReqid)
			rw.Header().Set(HeaderReqid, reqId)
			times := fmt.Sprintf("%0.3fms", float64(time.Since(start).Nanoseconds())/1e6)
			rw.Header().Set(HeaderResponseTime, times)
		})

		ctx.Next()

		status := res.Status()
		if status == 0 {
			status = http.StatusOK
		}
		times := fmt.Sprintf("%0.3fms", float64(time.Since(start).Nanoseconds())/1e6)
		reqEnd := fmt.Sprintf("[REQ_END] %d %0.3fk %s", status, float64(res.Size())/1024.0, times)
		if opt.ReqEndFilter != nil {
			reqEnd = opt.ReqEndFilter(ctx, rw, req, reqEnd)
		}
		if reqEnd != "" {
			log.Info(reqEnd, strip.LineOpt{Hidden: true})
		}
	}
}

func realIp(req *http.Request) string {
	if ip := req.Header.Get(HeaderXRealIp); ip != "" {
		return strings.TrimSpace(ip)
	}

	parts := strings.Split(req.Header.Get(HeaderXForwardedFor), ",")
	if len(parts) > 0 && parts[0] != "" {
		parts = strings.Split(parts[0], ":")
		return strings.TrimSpace(parts[0])
	}

	host := req.RemoteAddr
	if idx := strings.LastIndex(req.RemoteAddr, ":"); idx != -1 {
		host = host[:idx]
	}
	return host
}

func setLoggerInContext(ctx strip.Context, req *http.Request, log strip.Logger) *http.Request {
	var l = log
	reqCtx := utils.CtxWithValue(req.Context(), &l)
	req = req.WithContext(reqCtx)

	ctx.Provide(req)
	ctx.ReplaceContext(req.Context())
	return req
}
