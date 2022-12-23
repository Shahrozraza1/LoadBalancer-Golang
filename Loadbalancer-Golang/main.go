package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Server interface {
	Address() string
	IsAlive() bool
	Serve(w http.ResponseWriter, r *http.Request)
}

type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

func newSimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)
	handleErr(err)

	return &simpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}

}

type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []Server
}

func NewLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		port:            port,
		roundRobinCount: 0,
		servers:         servers,
	}
}

func handleErr(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
func (s *simpleServer) Address() string { return s.addr }
func (s *simpleServer) IsAlive() bool   { return true }
func (s *simpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
	s.proxy.ServeHTTP(rw, req)
}
func (Ib *LoadBalancer) getNexAvailableServer() Server {
	server := Ib.servers[Ib.roundRobinCount%len(Ib.servers)]
	for !server.IsAlive() {
		Ib.roundRobinCount++
		server = Ib.servers[Ib.roundRobinCount%len(Ib.servers)]
	}
	Ib.roundRobinCount++
	return server
}
func (Ib *LoadBalancer) serveProxy(rw http.ResponseWriter, req *http.Request) {
	targetServer := Ib.getNexAvailableServer()
	fmt.Printf("forwarding request to address %q\n", targetServer.Address())
	targetServer.Serve(rw, req)
}
func main() {
	servers := []Server{
		newSimpleServer("https://www.facebook.com"),
		newSimpleServer("https://www.bing.com"),
		newSimpleServer("http://www.duckduckgo.com"),
	}
	Ib := NewLoadBalancer("8009", servers)
	handleRedirect := func(rw http.ResponseWriter, req *http.Request) {
		Ib.serveProxy(rw, req)
	}
	http.HandleFunc("/", handleRedirect)

	fmt.Printf("serving requests at localhost:%s'\n", Ib.port)
	http.ListenAndServe(":"+Ib.port, nil)

}
