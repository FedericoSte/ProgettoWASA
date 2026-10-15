package api

import (
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// doLogin gestisce il login semplificato. Ritorna l'identificatore utente.
// Endpoint: POST /session
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// TODO: Implementare l'estrazione dello username dal body, la creazione nel DB, e la risposta 201 Created.
	w.WriteHeader(http.StatusNotImplemented)
}
