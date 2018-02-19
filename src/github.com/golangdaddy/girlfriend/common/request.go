package gf

import 	(
		"io"
		"net/http"
		)

type RequestInterface interface {
	Config() *Config
	FullPath() string
	IsTLS() bool
	Method() string
	Device() string
	Body(string) interface{}
	// accesses the request params of the payload
	Param(string) interface{}
	Params() map[string]interface{}
	StrParam(string) string
	SetParam(string, interface{})
	SetHeader(string, string)
	GetHeader(string) string
	RawBody() (*ResponseStatus, []byte)
	ReadBodyObject() *ResponseStatus
	ReadBodyArray() *ResponseStatus
	BodyArray() []interface{}
	BodyObject() map[string]interface{}
	Redirect(string, int) *ResponseStatus
	ServeFile(string)
	HttpError(string, int)
	Writer() io.Writer
	Write([]byte)
	Fail() *ResponseStatus
	Respond(args ...interface{}) *ResponseStatus
	// logging
	Debug(string)
	NewError(string) error
	Error(error)
	DebugJSON(interface{})
	Reflect(interface{})
	//
	Res() http.ResponseWriter
	R() *http.Request
}
