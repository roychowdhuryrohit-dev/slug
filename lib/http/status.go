package http

type StatusCode int

const (
	StatusOK                      StatusCode = 200
	StatusCreated                 StatusCode = 201
	StatusAccepted                StatusCode = 202
	StatusMovedPermanently        StatusCode = 301
	StatusFound                   StatusCode = 302
	StatusBadRequest              StatusCode = 400
	StatusUnauthorized            StatusCode = 401
	StatusForbidden               StatusCode = 403
	StatusNotFound                StatusCode = 404
	StatusMethodNotAllowed        StatusCode = 405
	StatusInternalServerError     StatusCode = 500
	StatusNotImplemented          StatusCode = 501
	StatusBadGateway              StatusCode = 502
	StatusHTTPVersionNotSupported StatusCode = 505
)

func (code StatusCode) GetStatus() string {
	switch code {
	case StatusOK:
		return "OK"
	case StatusCreated:
		return "Created"
	case StatusAccepted:
		return "Accepted"
	case StatusFound:
		return "Found"
	case StatusBadRequest:
		return "Bad Request"
	case StatusUnauthorized:
		return "Unauthorized"
	case StatusForbidden:
		return "Forbidden"
	case StatusNotFound:
		return "Not Found"
	case StatusMethodNotAllowed:
		return "Method Not Allowed"
	case StatusInternalServerError:
		return "Internal Server Error"
	case StatusNotImplemented:
		return "Not Implemented"
	case StatusBadGateway:
		return "Bad Gateway"
	case StatusHTTPVersionNotSupported:
		return "HTTP Version Not Supported"
	case StatusMovedPermanently:
		return "Moved Permanently"
	default:
		return ""
	}
}
