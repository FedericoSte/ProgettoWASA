package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// loginRequest rappresenta il JSON inviato dal client nella POST /session
type loginRequest struct {
	Name string `json:"name"`
}

// loginResponse rappresenta il JSON restituito dal server in caso di successo
type loginResponse struct {
	Identifier string `json:"identifier"`
}

// errorResponse rappresenta il JSON restituito in caso di errore (come definito nell'OpenAPI)
type errorResponse struct {
	Error string `json:"error"`
}

// doLogin gestisce il login semplificato. Ritorna l'identificatore utente.
// Endpoint: POST /session
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1. Decodifica il body JSON in una struct Go
	var req loginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "Formato JSON non valido"})
		return
	}

	// 2. Validazione dei dati (come definito in OpenAPI: minLength 3, maxLength 16)
	username := strings.TrimSpace(req.Name)
	if len(username) < 3 || len(username) > 16 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "Lo username deve essere compreso tra 3 e 16 caratteri"})
		return
	}

	// 3. Esegui la logica di business chiamando il Database
	identifier, err := rt.db.DoLogin(username)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Errore durante il login nel database")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: "Si è verificato un errore interno al server"})
		return
	}

	// 4. Prepara e invia la risposta JSON di successo (201 Created)
	resp := loginResponse{
		Identifier: identifier,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
