package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// setMyUserName permette all'utente autenticato di cambiare il proprio nome.
// Endpoint: PATCH /utenti/me/username
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID

	// 1. Decodifica la richiesta
	type nameRequest struct {
		Name string `json:"name"`
	}
	var req nameRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "Formato JSON non valido"})
		return
	}

	// 2. Validazione (minLength 3, maxLength 16 come da OpenAPI)
	newName := strings.TrimSpace(req.Name)
	if len(newName) < 3 || len(newName) > 16 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "Lo username deve essere compreso tra 3 e 16 caratteri"})
		return
	}

	// 3. Aggiornamento nel database
	err = rt.db.SetMyUserName(userID, newName)
	if err != nil {
		if err.Error() == "conflict: username already in use" {
			w.WriteHeader(http.StatusConflict) // 409 Conflict
			json.NewEncoder(w).Encode(errorResponse{Error: "Questo username è già in uso da un altro utente"})
			return
		}
		rt.baseLogger.WithError(err).Error("Errore nell'aggiornamento dello username")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: "Errore interno del server"})
		return
	}

	// 4. Risposta di successo: 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// setMyPhoto permette all'utente autenticato di aggiornare la propria foto profilo.
// Endpoint: PATCH /utenti/me/photo
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID

	// 1. Decodifica la richiesta
	type photoRequest struct {
		Photo string `json:"photo"`
	}
	var req photoRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "Formato JSON non valido"})
		return
	}

	// Controllo che la stringa della foto non sia vuota
	if strings.TrimSpace(req.Photo) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "L'immagine fornita non è valida"})
		return
	}

	// 2. Aggiornamento nel database
	err = rt.db.SetMyPhoto(userID, req.Photo)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Errore nell'aggiornamento della foto")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: "Errore interno del server"})
		return
	}

	// 3. Risposta di successo: 204 No Content
	w.WriteHeader(http.StatusNoContent)
}
