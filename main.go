package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/roychowdhuryrohit-dev/slug/lib/config"
	"github.com/roychowdhuryrohit-dev/slug/lib/http"
)

func main() {
	config.SetupConfig()
	port, _ := config.ConfigMap.Load(config.Port)
	dirPath, _ := config.ConfigMap.Load(config.DocumentRoot)
	timeout, _ := config.ConfigMap.Load(config.Timeout)

	router := http.NewFileRouter()
	handler, err := http.FileServer(dirPath.(http.Dir))
	if err != nil {
		log.Panicln(err.Error())
	}
	router.AddRoute("/", handler)
	srv := &http.Server{
		Addr:   fmt.Sprintf(":%s", port),
		Router: router,
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Panicln(err.Error())
		}
	}()
	log.Println("server started")

	<-exit
	log.Println("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout.(int))*time.Second)
	defer func() {
		cancel()
	}()
	srv.Shutdown(ctx)
	
}
