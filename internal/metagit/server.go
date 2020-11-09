package metagit

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	httpLogger "github.com/improbable-eng/go-httpwares/logging/logrus"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Server interface {
	Serve(addr string) error
	AddMapping(from string, to string) error
}

func NewServer() Server {
	return &server{
		proxys: map[string]*httputil.ReverseProxy{},
	}
}

type server struct {
	proxys map[string]*httputil.ReverseProxy
}

func (s *server) Serve(addr string) error {
	m := httpLogger.Middleware(logrus.NewEntry(logrus.StandardLogger()))(s)

	return http.ListenAndServe(addr, m)
}

func (s *server) AddMapping(from string, to string) error {
	target, err := url.Parse("https://" + to)
	if err != nil {
		return errors.Wrap(err, "failed to parse outgoing url")
	}

	rp := httputil.NewSingleHostReverseProxy(target)

	builtInDirector := rp.Director

	rp.Director = func(r *http.Request) {
		cleanHTTPRequestHeader(r)

		builtInDirector(r)

		r.Host = target.Host
	}

	s.proxys[from] = rp

	println(from + " " + to)

	return nil
}

func cleanHTTPRequestHeader(r *http.Request) {
	for k := range r.Header {
		if strings.HasPrefix(k, "X-Forwarded-") {
			continue
		}
		if strings.HasPrefix(k, "X-") {
			r.Header.Del(k)
		}
	}
}

func (s server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case isGitRequest(r):
		s.proxyHTTPGitRequest(w, r)
		return
	case isGoGetRequest(r):
		s.handleGoGetRequest(w, r)
		return
	default:
		http.Error(w, "", http.StatusBadRequest)
	}
}

func isGitRequest(r *http.Request) bool {
	switch {
	case strings.HasPrefix(r.URL.Query().Get("service"), "git-"):
		return true
	case strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-git-"):
		return true
	}

	return false
}

func isGoGetRequest(r *http.Request) bool {
	if r.URL.Query().Get("go-get") == "1" {
		return true
	}

	return false
}

func (s server) proxyHTTPGitRequest(w http.ResponseWriter, r *http.Request) {
	rp, ok := s.proxys[r.Host]
	if !ok {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	rp.ServeHTTP(w, r)
}

func (s server) handleGoGetRequest(w http.ResponseWriter, r *http.Request) {
	template :=
		`<!DOCTYPE html><html><head><meta name="go-import" content="%s git %s"></head><body>Nothing to see here.</body></html>`

	importPrefix := r.Host + r.URL.Path
	repoRoot := "https://" + importPrefix + ".git"

	result := fmt.Sprintf(template, importPrefix, repoRoot)

	w.Write([]byte(result))
}
