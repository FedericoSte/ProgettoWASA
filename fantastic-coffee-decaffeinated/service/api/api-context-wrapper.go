package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
)

// httpRouterHandler is the signature for functions that accepts a reqcontext.RequestContext in addition to those
// required by the httprouter package.
type httpRouterHandler func(http.ResponseWriter, *http.Request, httprouter.Params, reqcontext.RequestContext)

// wrap converte una funzione personalizzata in un handler standard per httprouter.
// Questa versione NON richiede l'autenticazione (usata ad es. per /session).
func (rt *_router) wrap(fn httpRouterHandler) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		reqUUID, err := uuid.NewV4()
		if err != nil {
			rt.baseLogger.WithError(err).Error("can't generate a request UUID")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var ctx = reqcontext.RequestContext{
			ReqUUID: reqUUID.String(),
			UserID:  "", // Non autenticato
		}
		fn(w, r, ps, ctx)
	}
}

// wrapAuth converte una funzione personalizzata e APPLICA LA VERIFICA DEL BEARER TOKEN.
// Se il token è invalido o mancante, blocca la richiesta con un 401 Unauthorized.
func (rt *_router) wrapAuth(fn httpRouterHandler) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		reqUUID, err := uuid.NewV4()
		if err != nil {
			rt.baseLogger.WithError(err).Error("can't generate a request UUID")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// 1. Estrazione del token dall'header "Authorization"
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Header Authorization mancante"})
			return
		}

		// Il formato deve essere "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Formato Authorization invalido. Usa 'Bearer <token>'"})
			return
		}
		token := parts[1]

		// 2. Verifica del token nel Database
		isValid, err := rt.db.CheckToken(token)
		if err != nil {
			rt.baseLogger.WithError(err).Error("Errore durante la verifica del token nel database")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !isValid {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Token non valido o utente inesistente"})
			return
		}

		// 3. Il token è valido! Creiamo il contesto con l'ID utente.
		var ctx = reqcontext.RequestContext{
			ReqUUID: reqUUID.String(),
			UserID:  token, // Salviamo l'ID dell'utente loggato!
		}

		// 4. Passiamo l'esecuzione alla funzione finale (es. sendMessage)
		fn(w, r, ps, ctx)
	}
}
