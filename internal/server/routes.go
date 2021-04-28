package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

// routes creates a new router(multiplexer) and
// registers all routes to it
func (srv *Server) routes() *mux.Router {
	router := mux.NewRouter()

	router.Handle("/", http.HandlerFunc(srv.handleRoot))

	router.Handle("/like", http.HandlerFunc(srv.handleLike))

	router.Handle("/csv", http.HandlerFunc(srv.handleCSV))

	router.PathPrefix("/static/").Handler(http.FileServer(http.FS(staticContents)))

	return router
}
