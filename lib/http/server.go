package http

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type Handler func(*Request, *Response)

type Router interface {
	AddRoute(string, Handler) error
	GetRoute(string) (Handler, error)
}

type ConnectionWatcher struct {
	mu        sync.RWMutex
	connCount int64
}

func (cw *ConnectionWatcher) UpdateCount(c int) {
	cw.mu.Lock()
	cw.connCount += int64(c)
	cw.mu.Unlock()
}

func (cw *ConnectionWatcher) GetCount() int64 {
	cw.mu.RLocker().Lock()
	c := cw.connCount
	cw.mu.RLocker().Unlock()
	return c
}

type Server struct {
	wg         sync.WaitGroup
	listener   net.Listener
	connection chan net.Conn
	shutdown   chan struct{}
	Router     Router
	Addr       string
	cw         *ConnectionWatcher
}

func (srv *Server) ListenAndServe() error {
	if srv.Addr == "" {
		return fmt.Errorf("address is not initialised")
	}
	if srv.Router == nil {
		return fmt.Errorf("router is not initialised")
	}
	if srv.listener != nil {
		return fmt.Errorf("server already listening on address %s", srv.Addr)
	}

	listener, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on address %s: %w", srv.Addr, err)
	}

	srv.listener = listener
	if srv.connection == nil {
		srv.connection = make(chan net.Conn)
	}
	if srv.shutdown == nil {
		srv.shutdown = make(chan struct{})
	}
	if srv.cw == nil {
		srv.cw = &ConnectionWatcher{}
	}

	srv.wg.Add(2)
	go srv.acceptConnections()
	go srv.handleConnections()

	log.Printf("serving HTTP on %s", srv.Addr)
	return nil
}

func (srv *Server) acceptConnections() {
	defer srv.wg.Done()
	defer srv.listener.Close()

	for {
		select {
		case <-srv.shutdown:
			return
		default:
			conn, err := srv.listener.Accept()
			if err != nil {
				log.Printf("unable to accept connection on address %s", srv.Addr)
				continue
			}
			srv.connection <- conn
			srv.cw.UpdateCount(1)
		}
	}
}

func (srv *Server) handleConnections() {
	defer srv.wg.Done()

	for {
		select {
		case <-srv.shutdown:
			log.Printf("closing all connections on %s", srv.Addr)
			return
		case conn := <-srv.connection:
			go srv.handleConnection(conn)
		}
	}
}

func (srv *Server) handleConnection(conn net.Conn) {
	defer srv.cw.UpdateCount(-1)
	defer conn.Close()

	handleRequest := func() int {
		req := &Request{
			conn:   conn,
			Header: make(Header),
		}
		res := &Response{
			Request: req,
			Header:  make(Header),
		}

		if err := req.ReadRequest(); err != nil {
			if err == io.EOF {
				return -1
			}
			log.Println(err.Error())
			res.StatusCode = StatusInternalServerError
			ErrorHandler(req, res)
			return 0
		}
		var timeout, max int
		//if c, ok := req.Header.Get("Connection"); ok && c == "keep-alive" && req.Proto == "HTTP/1.1" {
		if req.Proto == "HTTP/1.1" {
			conCount := srv.cw.GetCount()
			timeout, max = getKeepAliveHeuristic(conCount)
			conn.(*net.TCPConn).SetKeepAlive(true)
			conn.(*net.TCPConn).SetKeepAlivePeriod(time.Second * time.Duration(timeout))
			res.Header.Add("Keep-Alive", fmt.Sprintf("timeout=%d, max=%d", timeout, max))
		}
		handler, err := srv.Router.GetRoute(req.URL.Path)

		if err != nil {
			res.StatusCode = StatusBadRequest
			ErrorHandler(req, res)
		}
		handler(req, res)
		return timeout
	}
	var timer <-chan time.Time

	for {
		if timeout := handleRequest(); timeout != -1 {
			timer = time.After(time.Duration(timeout) * time.Second)
		}
		select {
		case <-timer:
			return
		default:
		}
	}

}

func (srv *Server) Shutdown(ctx context.Context) {
	close(srv.shutdown)
	srv.listener.Close()
	connDone := make(chan struct{})
	go func() {
		srv.wg.Wait()
		close(connDone)
	}()

	select {
	case <-ctx.Done():
		log.Println("timed out waiting for connections to finish")
		return
	case <-connDone:
		log.Print("server exited successfully")
		return
	}
}

func getKeepAliveHeuristic(connCount int64) (int, int) {
	
	return 5, 100
}
