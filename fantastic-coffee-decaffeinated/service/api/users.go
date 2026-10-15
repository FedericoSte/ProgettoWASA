package api

import (
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// setMyUserName permette all'utente autenticato di cambiare il proprio nome.
// Endpoint: PATCH /utenti/me/username
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// TODO: Implementare controllo Bearer token e logica di update nome utente.
	w.WriteHeader(http.StatusNotImplemented)
}

// setMyPhoto permette all'utente autenticato di aggiornare la propria foto profilo.
// Endpoint: PATCH /utenti/me/photo
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// TODO: Implementare controllo Bearer token e logica di update foto.
	w.WriteHeader(http.StatusNotImplemented)
}
