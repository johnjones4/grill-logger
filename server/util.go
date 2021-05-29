package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
)

func logRequestHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ri := &HTTPReqInfo{
			method:    r.Method,
			uri:       r.URL.String(),
			referer:   r.Header.Get("Referer"),
			userAgent: r.Header.Get("User-Agent"),
		}

		ri.ipaddr = requestGetRemoteAddress(r)

		m := httpsnoop.CaptureMetrics(h, w, r)

		ri.code = m.Code
		ri.size = m.Written
		ri.duration = m.Duration

		fmt.Printf("%s - - [%s] \"%s %s\" %d %d\n", ri.ipaddr, time.Now().Format(time.RFC3339), ri.method, ri.uri, ri.code, ri.size)
	}
	return http.HandlerFunc(fn)
}

func ipAddrFromRemoteAddr(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}

func requestGetRemoteAddress(r *http.Request) string {
	hdr := r.Header
	hdrRealIP := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrRealIP == "" && hdrForwardedFor == "" {
		return ipAddrFromRemoteAddr(r.RemoteAddr)
	}
	if hdrForwardedFor != "" {
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		return parts[0]
	}
	return hdrRealIP
}

func errorResponse(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	jsonResponse(w, Message{
		Error:   true,
		Message: fmt.Sprint(err),
	})
}

func jsonResponse(w http.ResponseWriter, info interface{}) {
	jsonInfo, err := json.Marshal(info)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Add("Content-type", "application/json")
	w.Write(jsonInfo)
}

func getLimitOffset(req *http.Request) (int, int, error) {
	limit, err := strconv.Atoi(req.URL.Query()["limit"][0])
	if err != nil {
		return 0, 0, err
	}
	offset, err := strconv.Atoi(req.URL.Query()["offset"][0])
	if err != nil {
		return 0, 0, err
	}
	return limit, offset, nil
}
