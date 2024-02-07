package http

import (
	"fmt"
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

//TODO Register all possible paths for serving files
