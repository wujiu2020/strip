package apires

import (
	"io"
	"net/http"

	"github.com/teapots/teapot"
)

type RawBody struct {
	Status int         `json:"-"`
	Body   interface{} `json:"-"`
}

func (r *RawBody) Write(ctx teapot.Context, rw http.ResponseWriter, req *http.Request) {
	if r.Status == 0 {
		r.Status = http.StatusOK
	}

	if r.Status == http.StatusNoContent {
		rw.WriteHeader(r.Status)
		return
	}

	if r.Body == nil {
		return
	}

	var err error
	switch v := r.Body.(type) {
	case string:
		_, err = rw.Write([]byte(v))
	case []byte:
		_, err = rw.Write(v)
	case io.Reader:
		buf := make([]byte, 4*1024)
		_, err = io.CopyBuffer(rw, v, buf)
		if c, ok := r.Body.(io.Closer); ok {
			c.Close()
		}
	default:
		err = writeJsonBody(r.Status, r.Body, ctx, rw, req)
	}
	if err != nil {
		var logger teapot.Logger
		ctx.Find(&logger, "")
		logger.Warn("response write failed:", err)
	}
}

func RawWith(body interface{}, status ...int) *RawBody {
	code := http.StatusOK
	if len(status) > 0 {
		code = status[0]
	}
	return &RawBody{Status: code, Body: body}
}

func RawRet(status int) *RawBody {
	return &RawBody{Status: status}
}
