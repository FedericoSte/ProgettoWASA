package api

import (
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getMyConversations ritorna l'array di conversazioni a cui partecipa l'utente.
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.WriteHeader(http.StatusNotImplemented)
}

// getConversation ritorna i dettagli di una singola conversazione e la sua cronologia.
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.WriteHeader(http.StatusNotImplemented)
}

// setGroupName modifica il nome di una conversazione di gruppo.
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.WriteHeader(http.StatusNotImplemented)
}

// setGroupPhoto modifica la foto di una conversazione di gruppo.
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.WriteHeader(http.StatusNotImplemented)
}

// addToGroup aggiunge un nuovo utente ai partecipanti di un gruppo.
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.WriteHeader(http.StatusNotImplemented)
}

// leaveGroup rimuove l'utente autenticato dai partecipanti del gruppo.
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.WriteHeader(http.StatusNotImplemented)
}

// sendMessage crea un nuovo messaggio nella conversazione.
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.WriteHeader(http.StatusNotImplemented)
}

// forwardMessage inoltra un messaggio esistente in una conversazione.
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.WriteHeader(http.StatusNotImplemented)
}

// deleteMessage elimina un messaggio precedentemente inviato.
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.WriteHeader(http.StatusNotImplemented)
}

// commentMessage aggiunge una reazione a un messaggio.
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.WriteHeader(http.StatusNotImplemented)
}

// uncommentMessage elimina una reazione precedentemente aggiunta a un messaggio.
func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.WriteHeader(http.StatusNotImplemented)
}
