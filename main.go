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
	portConfig, _ := config.ConfigMap.Load(config.Port)
	port, _ := portConfig.(*string)
	dirPathConfig, _ := config.ConfigMap.Load(config.DocumentRoot)
	dirPath, _ := dirPathConfig.(*string)
	timeoutEnvConfig, _ := config.ConfigMap.Load(config.Timeout)
	timeout, _ := timeoutEnvConfig.(*int)

	router := http.NewFileRouter()
	handler, err := http.FileServer(http.Dir(*dirPath))
	if err != nil {
		log.Panicln(err.Error())
	}
	router.AddRoute("/", handler)
	srv := &http.Server{
		Addr:   fmt.Sprintf(":%s", *port),
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)
	defer func() {
		cancel()
	}()
	srv.Shutdown(ctx)

}
