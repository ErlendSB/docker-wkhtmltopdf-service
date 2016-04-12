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

func main() {
	const bindAddress = ":9090"
	http.HandleFunc("/", requestHandler)
	fmt.Println("Http server listening on", bindAddress)
	http.ListenAndServe(bindAddress, nil)
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

func requestHandler(response http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
		response.WriteHeader(http.StatusNotFound)
		logOutput(request, "404 not found")
		return
	}
	if request.Method != "POST" {
		response.Header().Set("Allow", "POST")
		response.WriteHeader(http.StatusMethodNotAllowed)
		logOutput(request, "405 not allowed")
		return
	}
	decoder := json.NewDecoder(request.Body)
	var req documentRequest
	if err := decoder.Decode(&req); err != nil {
		response.WriteHeader(http.StatusBadRequest)
		logOutput(request, "400 bad request (invalid JSON)")
		return
	}
	segments := make([]string, 0)
	for key, element := range req.Options {
		if element == true {
			// if it was parsed from the JSON as an actual boolean,
			// convert to command-line single argument	(--foo)
			segments = append(segments, fmt.Sprintf("--%v", key))
		} else if element != false {
			// Otherwise, use command-line argument with value (--foo bar)
			segments = append(segments, fmt.Sprintf("--%v", key), fmt.Sprintf("%v", element))
		}
	}
	for key, value := range req.Cookies {
		segments = append(segments, "--cookie", key, url.QueryEscape(value))
	}
	var programFile string
	var contentType string
	switch req.Output {
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
	if (req.Html != "") {
		segments = append(segments, "-", "-")
	} else {
		segments = append(segments, req.Url, "-")
	}
	fmt.Println("[",time.Now().Format(time.RFC850),"]", programFile, strings.Join(segments, " "))

	cmd := exec.Command(programFile, segments...)
	if (req.Html != "") {
		cmd.Stdin = strings.NewReader(req.Html)
	}

	response.Header().Set("Content-Type", contentType)
	output, err := cmd.CombinedOutput()
	trimmed := cleanupOutput(output, req.Output)
	response.Header().Set("Content-Length", strconv.Itoa(len(trimmed)))
	response.Write(trimmed);
	if (err != nil) {
		logOutput(request, err.Error())
	} else {
		logOutput(request, "200 OK")
	}

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