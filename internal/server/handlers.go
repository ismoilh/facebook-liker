package server

import (
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/tebeka/selenium"
)

func (srv *Server) handleRoot(rw http.ResponseWriter, r *http.Request) {
	log.Logger.Info().Str("path", "/").Msgf("RECEIVED request")
	defer log.Logger.Info().Str("path", "/").Msgf("FINISHED request")

	err := srv.templates.ExecuteTemplate(rw, "like.html", nil)
	if err != nil {
		log.Logger.Err(err).Msg("FAILED to execute template like.html")
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (srv *Server) handleLike(rw http.ResponseWriter, r *http.Request) {
	log.Logger.Info().Str("path", "/like").Msgf("RECEIVED request")
	defer log.Logger.Info().Str("path", "/like").Msgf("FINISHED request")

	subLogger := log.Logger.With().Str("path", "/like").Logger()

	subLogger.Info().Msg("GETTING browser from pool")
	poolItem := srv.driverPool.Get()
	switch p := poolItem.(type) {
	case error:
		subLogger.Err(p).Msg("FAILED to get browser from pool")
		err := srv.templates.ExecuteTemplate(rw, "error.html", nil)
		if err != nil {
			log.Logger.Err(err).Msg("FAILED to execute template error.html")
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}

	subLogger.Info().Msg("SUCCESSFULLY got browser from pool")
	wd := poolItem.(selenium.WebDriver)
	defer func() { // quit browser after we are done working with it
		if err := wd.Close(); err != nil {
			subLogger.Err(err).Msg("FAILED to quit browser")
			return
		}
		subLogger.Info().Msg("SUCCESSFULLY quit browser")
	}()
	defer srv.driverPool.Put(wd) // also return browser to pool after we are done working with it

	if err := wd.Get("https://google.com"); err != nil {
		subLogger.Err(err).Msg("FAILED to navigate to google.com")
		return
	}
}

func (srv *Server) handleCSV(rw http.ResponseWriter, r *http.Request) {
	log.Logger.Info().Str("path", "/csv").Msgf("RECEIVED request")
	defer log.Logger.Info().Str("path", "/csv").Msgf("FINISHED request")

	err := srv.templates.ExecuteTemplate(rw, "csv.html", nil)
	if err != nil {
		log.Logger.Err(err).Msg("FAILED to execute template csv.html")
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}