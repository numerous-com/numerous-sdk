package server

import (
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gorilla/websocket"
)

const (
	maxQueryCacheSize     = 1000
	maxUploadCacheSize    = 100
	keepAlivePingInterval = 10
	maxUploadSize         = 1 << 20   // 256MB
	maxUploadMemory       = 128 << 20 // 128MB
)

func setServerSettings(srv *handler.Server) {
	// similar to NewDefaultServer but with customized websocket
	srv.AddTransport(&transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // TODO: make configurable and update the middleware too
			},
		},
		KeepAlivePingInterval: keepAlivePingInterval * time.Second,
	})

	// from NewDefaultServer
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{
		MaxUploadSize: maxUploadSize,
		MaxMemory:     maxUploadMemory,
	})

	srv.SetQueryCache(lru.New(maxQueryCacheSize))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(maxUploadCacheSize),
	})
}
