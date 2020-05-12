package apires

import (
	"encoding/json"
	"fmt"
)

type ResError struct {
	ResBody
	Code    int         `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func NewResError(status, code int, message string, datas ...interface{}) *ResError {
	var data interface{}
	if len(datas) > 0 {
		data = datas[0]
	}
	er := &ResError{Code: code, Message: message, Data: data}
	er.Status = status
	er.Body = er
	return er
}

func (e *ResError) Error() string {
	if e == nil {
		return ""
	}
	b, _ := json.Marshal(e)
	return string(b)
}

func (e *ResError) HttpCode() int {
	return e.Status
}

func (e *ResError) WithData(data interface{}, message ...string) *ResError {
	ne := *e
	if len(message) > 0 {
		ne.Message = message[0]
	} else {
		switch v := data.(type) {
		case error:
			ne.Message = v.Error()
		case string:
			ne.Message = v
		}
	}
	ne.ResBody = ResBody{Status: ne.Status, Body: &ne}
	ne.Data = data
	return &ne
}

func (e *ResError) WithMsg(message string) *ResError {
	ne := *e
	ne.Message = message
	ne.ResBody = ResBody{Status: ne.Status, Body: &ne}
	return &ne
}

func (e *ResError) WithMsgf(format string, a ...interface{}) *ResError {
	return e.WithMsg(fmt.Sprintf(format, a...))
}

func (e *ResError) EqualAny(errs ...*ResError) bool {
	for _, er := range errs {
		if e.Status == er.Status && e.Code == er.Code {
			return true
		}
	}
	return false
}

func IsResError(err error) (ok bool) {
	_, ok = err.(*ResError)
	return
}

func GetHttpCode(err error) (ok bool, status int) {
	type rpcError interface {
		HttpCode() int
	}
	if v, e := err.(rpcError); e {
		ok = true
		status = v.HttpCode()
		return
	}
	return
}
