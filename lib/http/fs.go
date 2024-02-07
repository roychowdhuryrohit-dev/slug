package http

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Dir string

func (d Dir) Open(fpath string) (*os.File, error) {
	path := filepath.FromSlash(path.Clean("/" + fpath))
	dir := string(d)
	if dir == "" {
		dir = "."
	}
	fullPath := filepath.Join(dir, path)
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file (%s)", fullPath)
	}
	return f, nil
}

func Read(f *os.File) ([]byte, error) {
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to read file")
	}

	bs := make([]byte, stat.Size())
	_, err = bufio.NewReader(f).Read(bs)
	if err != nil && err != io.EOF {
		fmt.Println(err)
		return nil, fmt.Errorf("failed to read file")
	}
	return bs, nil
}

func GetContentType(filename string) (contentType string) {
	idx := strings.LastIndex(filename, ".")
	if idx < 0 || (idx+1) == len(filename) {
		return "application/octet-stream"
	}
	fileExt := filename[idx+1:]
	switch fileExt {
	case "htm", "html", "shtml":
		return "text/html"
	case "css", "csv", "rtf", "xml":
		return fmt.Sprintf("text/%s", fileExt)
	case "txt":
		return "text/plain"
	case "js", "mjs":
		return "text/javascript"
	case "jpeg", "jpg":
		return "image/jpeg"
	case "heic", "png", "apng", "tiff", "webp", "bmp", "gif", "avif":
		return fmt.Sprintf("image/%s", fileExt)
	case "json", "gz", "pdf":
		return fmt.Sprintf("application/%s", fileExt)
	case "mp3", "aac":
		return fmt.Sprintf("audio/%s", fileExt)
	case "mp4", "mpeg":
		return fmt.Sprintf("video/%s", fileExt)
	default:
		return "application/octet-stream"
	}
}

type FileRouter struct {
	routesMap map[string]Handler
}

func (fr *FileRouter) AddRoute(routePath string, routeHandler Handler) error {
	if fr.routesMap == nil {
		return fmt.Errorf("router not initialised")
	}
	if _, ok := fr.routesMap[routePath]; ok {
		return fmt.Errorf("route path already in use")
	}
	fr.routesMap[routePath] = routeHandler
	return nil
}

func (fr *FileRouter) GetRoute(routePath string) (Handler, error) {
	for k, v := range fr.routesMap {
		if strings.HasPrefix(routePath, k) {
			return v, nil
		}
	}

	return nil, fmt.Errorf("invalid route path: %s", routePath)
}

func NewFileRouter() *FileRouter {
	var r FileRouter
	r.routesMap = make(map[string]Handler)
	return &r
}

func FileServer(d Dir) (Handler, error) {
	var fr FileRouter
	if _, err := os.Stat(string(d)); os.IsNotExist(err) {
		return nil, fmt.Errorf("invalid dir path: %s", d)
	}

	errorHandler := func(r *Request, w *Response) {
		if err := w.WriteStatusLine(); err != nil {
			log.Println(err.Error())
		}
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

		host, ok := r.Header.Get("Host")
		if !ok {
			host = ""
		}
		log.Printf("%s %s %s %s %v\n", host, r.Method, r.URL.Path, r.Proto, w.StatusCode)
	}

	handler := func(r *Request, w *Response) {
		d := d
		eh := errorHandler
		defer w.Flush()
		reqPath := r.URL.Path
		if strings.HasSuffix(reqPath, "/") {
			w.StatusCode = StatusMovedPermanently
			w.Header.Set("Location", fmt.Sprintf("%sindex.html", reqPath))
			if err := w.WriteStatusLine(); err != nil {
				log.Println(err.Error())
			}
			if err := w.WriteHeader(); err != nil {
				log.Println(err.Error())
			}
			host, ok := r.Header.Get("Host")
			if !ok {
				host = ""
			}
			log.Printf("%s %s %s %s %v\n", host, r.Method, r.URL.Path, r.Proto, w.StatusCode)
			return
		}
		w.Header.Set("Server", "Slug")
		w.Header.Set("Date", time.Now().Format(time.RFC1123))
		file, err := d.Open(r.URL.Path)
		if err != nil {
			log.Println(err.Error())
			w.StatusCode = StatusNotFound
			eh(r, w)
			return
		}
		if r.Proto != "HTTP/1.1" && r.Proto != "HTTP/1.0" {
			w.StatusCode = StatusHTTPVersionNotSupported
			eh(r, w)
			return
		} else {
			w.Proto = r.Proto
		}
		if r.Method != "GET" {
			w.StatusCode = StatusMethodNotAllowed
			eh(r, w)
			return
		}
		fileData, err := Read(file)
		if err != nil {
			log.Println(err.Error())
			w.StatusCode = StatusNotFound
			eh(r, w)
			return
		}
		w.Body = fileData
		contentType := GetContentType(r.URL.Path)
		w.StatusCode = StatusOK
		w.Header.Set("Content-Type", contentType)
		w.Header.Set("Content-Length", strconv.Itoa(len(w.Body)))
		if err := w.WriteStatusLine(); err != nil {
			log.Println(err.Error())
		}
		if err := w.WriteHeader(); err != nil {
			log.Println(err.Error())
		}
		if err := w.WriteBody(); err != nil {
			log.Println(err.Error())
		}

		host, ok := r.Header.Get("Host")
		if !ok {
			host = ""
		}
		log.Printf("%s %s %s %s %v\n", host, r.Method, r.URL.Path, r.Proto, w.StatusCode)
	}

	fr.routesMap = make(map[string]Handler)
	return handler, nil
}
