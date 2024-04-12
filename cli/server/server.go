package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"numerous/cli/appdev"
	"numerous/cli/graphql"

	graphqlHandler "github.com/99designs/gqlgen/graphql/handler"
	graphqlPlayground "github.com/99designs/gqlgen/graphql/playground"
	"github.com/rs/cors"
)

type HandlerRegister = func(mux *http.ServeMux)

type ServerOptions struct {
	HTTPPort          string
	AppSessions       appdev.AppSessionRepository
	AppSessionService appdev.AppSessionService
	Registers         []HandlerRegister
	GQLPath           string
	PlaygroundPath    string
}

func CreateServer(opts ServerOptions) *http.Server {
	mux := http.NewServeMux()
	corsOptions := cors.Options{
		AllowCredentials: true,
		AllowedHeaders:   []string{"*"},
		Debug:            false,
	}
	handler := cors.New(corsOptions).Handler(mux)
	addGraphQLHandlers(opts, mux)
	for _, register := range opts.Registers {
		register(mux)
	}

	return &http.Server{Addr: ":" + opts.HTTPPort, Handler: handler}
}

func RunServer(server *http.Server) {
	if err := server.ListenAndServe(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func addGraphQLHandlers(opts ServerOptions, mux *http.ServeMux) {
	cfg := graphql.Config{
		Resolvers: &graphql.Resolver{
			AppSessionsRepo:    opts.AppSessions,
			ToolSessionService: opts.AppSessionService,
		},
	}

	srv := graphqlHandler.New(graphql.NewExecutableSchema(cfg))

	setServerSettings(srv)

	mux.Handle(opts.PlaygroundPath, graphqlPlayground.Handler("GraphQL playground", opts.GQLPath))
	mux.Handle(opts.GQLPath, srv)

	slog.Info(fmt.Sprintf("Query GraphQL API at http://localhost:%s%s", opts.HTTPPort, opts.GQLPath))
	slog.Info(fmt.Sprintf("Connect to http://localhost:%s%s for GraphQL playground", opts.HTTPPort, opts.PlaygroundPath))
}
