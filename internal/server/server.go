package server

import (
	"context"
	"embed"
	"facebook-liker/internal/scrapper"
	"net/http"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

// Server is a structs to hold all dependencies for server
type Server struct {
	httpServer *http.Server

	templates *template.Template

	// a pool of selenium drivers(browsers)
	scrp *scrapper.Scrapper
}

//go:embed templates/*.html
var templatesContents embed.FS

//go:embed static
var staticContents embed.FS

// New constructs http server, configures it,
// and returns it
func New(addr string) (*Server, error) {
	var (
		err error
		srv = &Server{}
	)

	srv.httpServer = &http.Server{
		Addr:    addr,
		Handler: srv.routes(),
	}

	srv.templates, err = template.New("").ParseFS(templatesContents, "templates/*")
	if err != nil {
		return nil, errors.WithMessage(err, "failed to parse templates")
	}

	srv.scrp, err = scrapper.New("selenium-hub:4444")
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create new scrapper")
	}

	return srv, nil
}

// Run runs http server,
// it returns if context is canceled or error is returned from server
func (srv *Server) Run(ctx context.Context, errChan chan<- error) {
	// run on separate goroutine our http server
	// so it doesn't block Run function
	go func(errChan chan<- error) {
		if err := srv.httpServer.ListenAndServe(); err != nil {
			errChan <- err
		}
	}(errChan)

	// wait until calling func says that the work is done on the behalf of Run function
	<-ctx.Done()
	newCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// gracefully shutdown the server
	err := srv.httpServer.Shutdown(newCtx)
	if err != nil {
		errChan <- err
	}
}
