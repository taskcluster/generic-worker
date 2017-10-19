package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/taskcluster/slugid-go/slugid"
	"github.com/taskcluster/webhooktunnel/whclient"
)

type WebhookServer struct {
	Client *whclient.Client
	m      sync.Mutex
	hooks  map[string]http.Handler
}

func NewWebhookServer(client *whclient.Client) *WebhookServer {
	return &WebhookServer{
		Client: client,
		hooks:  make(map[string]http.Handler),
	}
}

func (wh *WebhookServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) < 24 || r.URL.Path[23] != '/' {
		http.NotFound(w, r)
		return
	}
	id, path := r.URL.Path[1:23], r.URL.Path[23:]
	log.Printf("Request for id: %s, path: %s url.Path:%s\n", id, path, r.URL.Path)

	wh.m.Lock()
	handler, ok := wh.hooks[id]
	wh.m.Unlock()

	if !ok {
		log.Printf("Hook not found")
		http.NotFound(w, r)
		return
	}

	r.URL.Path = path
	handler.ServeHTTP(w, r)
}

func (wh *WebhookServer) AttachHook(handler http.Handler) (string, func()) {
	id := slugid.Nice()
	wh.m.Lock()
	defer wh.m.Unlock()
	wh.hooks[id] = handler

	url := wh.Client.URL() + "/" + id + "/"
	detach := func() {
		wh.m.Lock()
		defer wh.m.Unlock()
		delete(wh.hooks, id)
	}

	return url, detach
}

func (wh *WebhookServer) Initialise() {
	server := &http.Server{Handler: wh}
	go server.Serve(wh.Client)
}
