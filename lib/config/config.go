package config

import (
	"flag"
	"sync"
)

const (
	Port         = "port"
	DocumentRoot = "document_root"
	Timeout      = "timeout"
)

var ConfigMap sync.Map

func SetupConfig() {
	port := flag.String(Port, "8080", "port to serve on")
	dirPath := flag.String(DocumentRoot, ".", "the absolute directory path to host")
	timeout := flag.Int(Timeout, 5, "timeout for graceful shutdown (in seconds)")
	

	ConfigMap.Store(Port, port)
	ConfigMap.Store(DocumentRoot, dirPath)
	ConfigMap.Store(Timeout, timeout)
}
