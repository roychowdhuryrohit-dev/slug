package http

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
)

type Request struct {
	conn          net.Conn
	Method        string
	Proto         string
	URL           *url.URL
	Header        Header
	ContentLength int64
	Body          []byte
}

type Response struct {
	Request    *Request
	Proto      string
	StatusCode StatusCode
	Header     Header
	Body       []byte
	writer     *bufio.Writer
}

func (r *Request) ReadRequest() error {
	if r.conn == nil {
		return errors.New("unable to parse request header line (no connection found)")
	}
	reader := bufio.NewReader(r.conn)
	statusLineProcessed := false
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			// return errors.New("unable to parse request header line (no newline found)")
			return err
		}
		headerKey, headerValues, found := strings.Cut(message, " ")
		if !found {
			return fmt.Errorf("unable to parse request header line (no whitespace found) - %s", message)
		}
		headerKey, _ = strings.CutSuffix(headerKey, ":")
		if statusLineProcessed {
			headerKey = textproto.CanonicalMIMEHeaderKey(headerKey)
		}
		statusLineProcessed = true
		headerValues = strings.Replace(headerValues, "\r\n", "", -1)
		switch headerKey {
		case "GET", "POST", "PUT", "DELETE":
			if r.Method == "" {
				r.Method = headerKey
				urlMessage, protoMessage, found := strings.Cut(headerValues, " ")
				if !found {
					return fmt.Errorf("unable to parse request header line (no whitespace found) - %s", headerValues)
				}
				r.URL, err = url.ParseRequestURI(urlMessage)
				if err != nil {
					return errors.New("unable to parse request header line (invalid request URL)")
				}
				if protoMessage == "" {
					return errors.New("unable to parse request header line (request protocol not found)")
				}
				r.Proto = protoMessage
			}
		case "Content-Length":
			contentLenMessage, _ := strconv.ParseInt(headerValues, 10, 64)
			r.ContentLength = contentLenMessage
			fallthrough
		default:
			for _, headerVal := range strings.Split(headerValues, ";") {
				r.Header.Add(headerKey, headerVal)
			}
		}
		if strings.HasSuffix(message, "\r\n\r\n") {
			break
		}

	}
	if r.ContentLength > 0 {
		r.Body = make([]byte, r.ContentLength)
		for i := 0; i < int(r.ContentLength); i++ {
			byteMessage, err := reader.ReadByte()
			if err != nil {
				return errors.New("unable to parse request body")
			}
			r.Body[i] = byteMessage
		}
	}
	return nil
}

func (w *Response) WriteStatusLine() error {
	if w.Request.conn == nil {
		return errors.New("unable to write response status line (no request connection found)")
	}
	w.writer = bufio.NewWriter(w.Request.conn)

	if w.Proto == "" {
		w.Proto = "HTTP/1.1"
	}

	if w.StatusCode == 0 {
		return errors.New("empty status in response")
	}
	statusText := w.StatusCode.GetStatus()
	_, err := w.writer.WriteString(fmt.Sprintf("%s %d %s\r\n", w.Proto, w.StatusCode, statusText))
	if err != nil {
		return errors.New("unable to write response status line")
	}
	return nil
}

func (w *Response) WriteHeader() error {
	if w.writer == nil {
		return errors.New("unable to write response header line (write buffer not initialised)")
	}

	for k, v := range w.Header {
		valueMessage := strings.Join(v, ";")
		_, err := w.writer.WriteString(fmt.Sprintf("%s: %s\r\n", k, valueMessage))
		if err != nil {
			return errors.New("unable to write response header lines")
		}
	}

	_, err := w.writer.WriteString("\r\n")
	if err != nil {
		return errors.New("unable to write response header lines")
	}
	return nil
}

func (w *Response) WriteBody() error {
	if w.writer == nil {
		return errors.New("unable to write response header line (write buffer not initialised)")
	}

	if w.Body == nil {
		return errors.New("unable to write body payload (empty Body)")
	}
	contentType, ok := w.Header.Get("Content-Type")
	if !ok {
		return errors.New("unable to write body payload (Content-Type header missing)")
	}

	if strings.HasPrefix(contentType, "text") {
		_, err := w.writer.WriteString(string(w.Body))
		if err != nil {
			return errors.New("unable to write response header lines")
		}

	} else {
		_, err := w.writer.Write(w.Body)
		if err != nil {
			return errors.New("unable to write response header lines")
		}
	}

	return nil
}

func (w *Response) Flush() {
	defer w.writer.Flush()

}
