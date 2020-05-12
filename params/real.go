package params

import (
	"net/http"
	"net/url"
	"strings"
)

func RealIp(req *http.Request) (realIp string) {
	realIp = req.Header.Get("X-Real-Ip")
	if realIp != "" {
		return
	}

	ips := strings.Split(req.Header.Get("X-Forwarded-For"), ",")
	if len(ips) > 0 && ips[0] != "" {
		rip := strings.Split(ips[0], ":")
		realIp = rip[0]
		return
	}

	ip := strings.Split(req.RemoteAddr, ":")
	if len(ip) > 0 {
		if ip[0] != "[" {
			realIp = ip[0]
			return
		}
	}
	return "127.0.0.1"
}

func RealHost(req *http.Request) (host string) {
	host = req.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = req.Host
	}
	parts := strings.Split(host, ":")
	if len(parts) > 0 {
		host = parts[0]
	}
	return
}

func RealProto(req *http.Request) (proto string) {
	proto = req.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		proto = req.URL.Scheme
	}
	if proto == "" {
		proto = "http"
	}
	return
}

func RealURI(req *http.Request) (uri string) {
	uri = req.Header.Get("X-Original-URI")
	if uri == "" {
		uri = req.URL.RequestURI()
	}
	return uri
}

func RealURL(req *http.Request) (u *url.URL, err error) {
	ins, err := url.ParseRequestURI(RealURI(req))
	if err != nil {
		return
	}
	d := *req.URL
	d.Host = RealHost(req)
	d.Scheme = RealProto(req)
	d.Opaque = ins.Opaque
	d.Path = ins.Path
	d.RawQuery = ins.RawQuery
	return &d, nil
}
