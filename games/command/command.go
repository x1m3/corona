package command

type Request interface {
	Request()
	Type() int
}

type Response interface {
	Response()
	Type() int
}

type request struct {}

func (r *request) Request() {}

type response struct {}

func (r *response) Response() {}
