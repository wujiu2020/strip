package reqlogger

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/teapots/inject"
	"github.com/teapots/teapot"
	"github.com/teapots/utils"
)

const (
	HeaderReqid         = "X-Reqid"
	HeaderResponseTime  = "X-Response-Time"
	HeaderXRealIp       = "X-Real-Ip"
	HeaderXForwardedFor = "X-Forwarded-For"
)

type LoggerContentFilter func(ctx teapot.Context, rw http.ResponseWriter, req *http.Request, content string) string

type LoggerOption struct {
	ColorMode     bool
	LineInfo      bool
	ShortLine     bool
	FlatLine      bool
	LogStackLevel teapot.Level
	ReqidFilter   LoggerContentFilter
	PrefixFilter  LoggerContentFilter
	ReqBegFilter  LoggerContentFilter
	ReqEndFilter  LoggerContentFilter
}

// ReqLogger filter
func ReqLoggerFilter(out teapot.LogPrinter, opts ...LoggerOption) inject.Provider {
	var opt LoggerOption
	if len(opts) > 0 {
		opt = opts[0]
	}
	return func(ctx teapot.Context, rw http.ResponseWriter, req *http.Request) {

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

		log := teapot.NewReqLogger(out, reqId)
		log.SetPrefix(prefix)

		// write request id to request and response
		req.Header.Set(HeaderReqid, reqId)
		rw.Header().Set(HeaderReqid, reqId)

		log.SetLineInfo(opt.LineInfo)
		log.SetShortLine(opt.ShortLine)
		log.SetFlatLine(opt.FlatLine)
		log.SetColorMode(opt.ColorMode)
		log.EnableLogStack(opt.LogStackLevel)

		ctx.ProvideAs(log, (*teapot.Logger)(nil))
		ctx.ProvideAs(log, (*teapot.ReqLogger)(nil))

		req = setLoggerInContext(ctx, req, log)

		remoteAddr := realIp(req)

		start := time.Now()
		reqBeg := fmt.Sprintf("[REQ_BEG] %s %s%s %s", req.Method, req.Host, req.URL, remoteAddr)
		if opt.ReqBegFilter != nil {
			reqBeg = opt.ReqBegFilter(ctx, rw, req, reqBeg)
		}
		if reqBeg != "" {
			log.Info(reqBeg, teapot.LineOpt{Hidden: true})
		}

		res := rw.(teapot.ResponseWriter)
		res.Before(func(rw teapot.ResponseWriter) {
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
			log.Info(reqEnd, teapot.LineOpt{Hidden: true})
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

func setLoggerInContext(ctx teapot.Context, req *http.Request, log teapot.Logger) *http.Request {
	var l teapot.Logger = log
	reqCtx := utils.CtxWithValue(req.Context(), &l)
	req = req.WithContext(reqCtx)

	ctx.Provide(req)
	ctx.ReplaceContext(req.Context())
	return req
}
