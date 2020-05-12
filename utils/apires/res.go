package apires

import (
	"encoding/json"
	"net/http"

	"github.com/teapots/teapot"
)

var jsonContentType = "application/json; charset=UTF-8"

type ResBody struct {
	Status int         `json:"-"`
	Body   interface{} `json:"-"`
}

func (r *ResBody) Write(ctx teapot.Context, rw http.ResponseWriter, req *http.Request) {
	if r.Status == 0 {
		r.Status = http.StatusOK
	}
	if r.Body == nil {
		r.Body = struct{}{}
	}

	if r.Status == http.StatusNoContent {
		rw.WriteHeader(r.Status)
		return
	}

	if err := writeJsonBody(r.Status, r.Body, ctx, rw, req); err != nil {
		var logger teapot.Logger
		ctx.Find(&logger, "")
		logger.Warn("response write failed:", err)
	}
}

func With(body interface{}, status ...int) *ResBody {
	code := http.StatusOK
	if len(status) > 0 {
		code = status[0]
	}
	return &ResBody{Status: code, Body: body}
}

func Ret(status int) *ResBody {
	return &ResBody{Status: status}
}

func writeJsonBody(status int, body interface{}, ctx teapot.Context, rw http.ResponseWriter, req *http.Request) (err error) {
	config := new(teapot.Config)
	// use struct, so u can just skip error
	ctx.Find(&config, "")

	var resBody []byte
	if config.RunMode.IsDev() {
		resBody, err = json.MarshalIndent(body, "", "  ")
	} else {
		resBody, err = json.Marshal(body)
	}

	rw.Header().Set("Content-Type", jsonContentType)
	rw.WriteHeader(status)

	if err == nil {
		_, err = rw.Write(resBody)
	} else {
		res := map[string]interface{}{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		}
		if config.RunMode.IsDev() {
			resBody, _ = json.MarshalIndent(res, "", "  ")
		} else {
			resBody, _ = json.Marshal(res)
		}
		rw.Write(resBody)
	}
	return
}
