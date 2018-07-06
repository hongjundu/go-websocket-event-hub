package wsevent

const (
	ResponseStatusOK    = "ok"
	ResponseStatusError = "error"
)

type response struct {
	Status string      `json:"status"`
	Code   string      `json:"code,omitempty"`
	Msg    string      `json:"msg,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

func newOKResponse(data interface{}) *response {
	return &response{Status: ResponseStatusOK, Data: data}
}

func newErrorResponse(err error) *response {
	if err == nil {
		return newOKResponse(nil)
	} else {
		if theError, ok := err.(*Error); ok {
			return &response{Status: ResponseStatusError, Code: theError.Code(), Msg: theError.Error(), Data: nil}
		} else {
			return &response{Status: ResponseStatusError, Code: ErrorCodeServerError, Msg: err.Error(), Data: nil}
		}
	}
}
