package gee

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"

	"golang.org/x/crypto/acme/autocert"
)

var (
	m       *autocert.Manager
	tlsconf *tls.Config
)

func init() {

	m = &autocert.Manager{
		Cache:      autocert.DirCache("secret-dir"),
		Prompt:     autocert.AcceptTOS,
		Email:      "xxxx@gmail.com",
		HostPolicy: autocert.HostWhitelist(os.Args[1:]...),
	}

	tlsconf = m.TLSConfig()
	tlsconf.MinVersion = tls.VersionTLS13

}

func (engine *Engine) Run(addr string) error {

	tlsListen, err := tls.Listen("tcp", ":443", tlsconf)
	if err != nil {
		log.Println("tls Listen 443", err)
	}
	defer tlsListen.Close()

	return http.Serve(tlsListen, m.HTTPHandler(engine))

}

//go build -ldflags="-s -w" ./
