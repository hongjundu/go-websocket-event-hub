package wsevent

const (
	ErrorCodeServerError  = "server_error"
	ErrorUnregistered     = "unregistered"
	ErrorCodeNotSupported = "not_supported"
)

type Error struct {
	code string
	msg  string
}

func NewError(code, msg string) error {
	return &Error{code, msg}
}

func (this Error) Code() string {
	return this.code
}

func (this Error) Error() string {
	return this.msg
}
