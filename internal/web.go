package internal

import "net/http"

type ResponseWriter = http.ResponseWriter
type Request = http.Request
type HttpFunc = func(ResponseWriter, *Request)
