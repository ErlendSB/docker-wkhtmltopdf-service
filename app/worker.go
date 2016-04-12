package main

import (
	"net/http"
	"net/url"
	"fmt"
	"os/exec"
	"encoding/json"
	"strings"
	"strconv"
	"image/png"
	"image/jpeg"
	"bytes"
	"time"
)

type Request struct {
    ProgramFile string // changed
    Segments []string
    Html string
    ResultChan chan []byte // changed
}

func WorkerPool(n int) chan *Request {
    requests := make(chan *Request)

    for i:=0; i<n; i++ {
        go Worker(requests)
    }

    return requests
}

func Worker(requests chan *Request) {
    for request := range requests {
        path := CreateImage(request.ProgramFile, request.Segments, request.Html)
        request.ResultChan <- path // changed
    }
}

func CreateImage(programFile string, segments []string, html string) []byte {
	fmt.Println("[",time.Now().Format(time.RFC850),"]", programFile, strings.Join(segments, " "))

	cmd := exec.Command(programFile, segments...)
	if (html != "") {
		cmd.Stdin = strings.NewReader(html)
	}

	output, _ := cmd.CombinedOutput()
	return cleanupOutput(output, "png")

}

type Server struct {
    Requests chan *Request
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		logOutput(req, "404 not found")
		return
	}
	if req.Method != "POST" {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		logOutput(req, "405 not allowed")
		return
	}
	decoder := json.NewDecoder(req.Body)
	var payload documentRequest
	if err := decoder.Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logOutput(req, "400 bad request (invalid JSON)")
		return
	}
	segments := make([]string, 0)
	for key, element := range payload.Options {
		if element == true {
			// if it was parsed from the JSON as an actual boolean,
			// convert to command-line single argument	(--foo)
			segments = append(segments, fmt.Sprintf("--%v", key))
		} else if element != false {
			// Otherwise, use command-line argument with value (--foo bar)
			segments = append(segments, fmt.Sprintf("--%v", key), fmt.Sprintf("%v", element))
		}
	}
	for key, value := range payload.Cookies {
		segments = append(segments, "--cookie", key, url.QueryEscape(value))
	}
	var programFile string
	var contentType string
	switch payload.Output {
		case "jpg":
			programFile = "/usr/bin/wkhtmltoimage"
			contentType = "image/jpeg"
			segments = append(segments, "--format", "jpg", "-q")
		case "png":
			programFile = "/usr/bin/wkhtmltoimage"
			contentType = "image/png"
			segments = append(segments, "--format", "png", "-q")
		default:
			// defaults to pdf
			programFile = "/usr/bin/wkhtmltopdf"
			contentType = "application/pdf"
	}
	if (payload.Html != "") {
		segments = append(segments, "-", "-")
	} else {
		segments = append(segments, payload.Url, "-")
	}
    request := &Request{ProgramFile: programFile, Segments: segments, Html: payload.Html, ResultChan: make(chan []byte)}
    s.Requests <- request
    trimmed := <-request.ResultChan // this blocks
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(trimmed)))
	w.Write(trimmed);
	// if (err != nil) {
	// 	logOutput(req, err.Error())
	// } else {
	// 	logOutput(req, "200 OK")
	// }

    //http.ServeFile(w, req, path)
}
func main() {
    requests := WorkerPool(5)
    server := &Server{Requests: requests}
    http.ListenAndServe(":9090", server)
}

type documentRequest struct {
	Url string
	Output string
	Html string
	// TODO: whitelist options that can be passed to avoid errors,
	// log warning when different options get passed
	Options map[string]interface{}
	Cookies map[string]string
}

func logOutput(request *http.Request, message string) {
	ip := strings.Split(request.RemoteAddr, ":")[0]
	fmt.Println("[",time.Now().Format(time.RFC850),"]",ip, request.Method, request.URL, message)
}


func cleanupOutput(img []byte, format string) []byte {
	buf := new(bytes.Buffer)
	switch {
	case format == "png":
		decoded, err := png.Decode(bytes.NewReader(img))
		for err != nil {
			img = img[1:]
			if len(img) == 0 {
				break
			}
			decoded, err = png.Decode(bytes.NewReader(img))
		}
		png.Encode(buf, decoded)
		return buf.Bytes()
	case format == "jpg":
		decoded, err := jpeg.Decode(bytes.NewReader(img))
		for err != nil {
			img = img[1:]
			if len(img) == 0 {
				break
			}
			decoded, err = jpeg.Decode(bytes.NewReader(img))
		}
		jpeg.Encode(buf, decoded, nil)
		return buf.Bytes()
		// case format == "svg":
		// 	return img
	}
	return img
}