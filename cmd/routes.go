package cmd

import "github.com/gorilla/mux"

func (app *application) setupRoutes() *mux.Router {
	mux := mux.NewRouter()
	mux.HandleFunc("/healthz", GetHealth)
	mux.HandleFunc("/healthcheck", GetHealth)
	mux.HandleFunc("/validate", app.Validate)
	mux.HandleFunc("/mutate", app.Mutate)
	return mux
}
