package wsevent

import "time"

const (
	responseStatusKey    = "status"
	responseErrorCodeKey = "code"
	responseErrorMsgKey  = "msg"
)

const (
	responseStatusOK    = "ok"
	responseStatusError = "err"
)

type responseMessage struct {
	Type string      `json:"t"`
	Data interface{} `json:"d"`
	Time int64       `json:"time"`
}

type responseData map[string]interface{}

func newOKResponseData(key string, data interface{}) responseData {
	return responseData{responseStatusKey: responseStatusOK, key: data}
}

func newErrorResponseData(err error) responseData {
	if err == nil {
		return responseData{responseStatusKey: responseStatusOK}
	}

	if theError, ok := err.(*Error); ok {
		return responseData{responseStatusKey: responseStatusError, responseErrorCodeKey: theError.Code(), responseErrorMsgKey: theError.Error()}
	} else {
		return responseData{responseStatusKey: responseStatusError, responseErrorCodeKey: ErrorCodeServerError, responseErrorMsgKey: err.Error()}
	}
}

func newResponseMessage(t string, d interface{}) *responseMessage {
	return &responseMessage{Type: t, Data: d, Time: time.Now().Unix()}
}
