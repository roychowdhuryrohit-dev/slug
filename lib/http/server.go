package http

import (
	"context"
	"fmt"
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

type Server struct {
	wg         sync.WaitGroup
	listener   net.Listener
	connection chan net.Conn
	shutdown   chan struct{}
	Router     Router
	Addr       string
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

	srv.wg.Add(2)
	go srv.acceptConnections()
	go srv.handleConnections()

	log.Printf("serving HTTP on %s", srv.Addr)
	return nil
}

func (srv *Server) acceptConnections() {
	defer srv.wg.Done()

	for {
		select {
		case <-srv.shutdown:
			return
		default:
			conn, err := srv.listener.Accept()
			if err != nil {
				log.Printf("unable to accept connect on address %s", srv.Addr)
				continue
			}
			srv.connection <- conn
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
			req := &Request{
				conn: conn,
			}
			req.ReadRequest()
			res := &Response{
				Request: req,
			}
			if c, ok := req.Header.Get("Connection"); ok && c == "keep-alive" && req.Proto == "HTTP/1.1" {
				conn.(*net.TCPConn).SetKeepAlive(true)
				conn.(*net.TCPConn).SetKeepAlivePeriod(time.Second * time.Duration(srv.getKeepAliveHeuristic(-1)))
			}

			handler, err := srv.Router.GetRoute(req.URL.Path)

			if err != nil {
				log.Println(err)
			}
			go handler(req, res)
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

func (srv *Server) getKeepAliveHeuristic(n int) int {
	if n == -1 {
		return 5
	}
	return 0
}