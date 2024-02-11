package http

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

type BasicRouter struct {
	routesMap map[string]Handler
}

func (router *BasicRouter) AddRoute(routePath string, routeHandler Handler) error {
	if router.routesMap == nil {
		return fmt.Errorf("router not initialised")
	}

	if _, ok := router.routesMap[routePath]; ok {
		return fmt.Errorf("route path already in use")
	}
	router.routesMap[routePath] = routeHandler
	return nil
}

func (router *BasicRouter) GetRoute(routePath string) (Handler, error) {
	r, ok := router.routesMap[routePath]
	if !ok {
		return nil, fmt.Errorf("invalid route path: %s", routePath)

	}

	return r, nil
}

func NewBasicRouter() *BasicRouter {
	var r BasicRouter
	r.routesMap = make(map[string]Handler)
	return &r
}

func ErrorHandler(r *Request, w *Response) {
	defer func() {
		w.Flush()
	}()
	if err := w.WriteStatusLine(); err != nil {
		log.Println(err.Error())
	}
	w.Header.Set("Server", "Slug")
	w.Header.Set("Date", time.Now().Format(time.RFC1123))
	w.Header.Set("Content-Type", "text/html")
	w.Body = []byte(
		`<!DOCTYPE html>
		<html>
			<body>
				<h1>` + w.StatusCode.GetStatus() + `</h1>
			</body>
		</html>`)
	w.Header.Set("Content-Length", strconv.Itoa(len(w.Body)))
	if err := w.WriteHeader(); err != nil {
		log.Println(err.Error())
	}
	if err := w.WriteBody(); err != nil {
		log.Println(err.Error())
	}

	if r.URL != nil {
		log.Printf("%s %s %s %v\n", r.Method, r.URL.Path, r.Proto, w.StatusCode)
	}

}
