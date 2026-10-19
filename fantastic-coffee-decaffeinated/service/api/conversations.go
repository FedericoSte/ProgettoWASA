package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/database"
	"github.com/julienschmidt/httprouter"
)

// getMyConversations ritorna l'array di conversazioni a cui partecipa l'utente.
// Endpoint: GET /conversazioni
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// L'ID dell'utente che sta facendo la richiesta è salvato nel contesto (grazie al wrapAuth)
	userID := ctx.UserID

	// Chiamiamo il DB
	conversations, err := rt.db.GetMyConversations(userID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error fetching conversations")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: "Errore interno del server"})
		return
	}

	// Restituiamo 200 OK con l'array JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(conversations)
}

// getConversation ritorna i dettagli di una singola conversazione e la sua cronologia.
// Endpoint: GET /conversazioni/:conversazioneID
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	conversationID := ps.ByName("conversazioneID")

	// Chiamiamo il DB
	conv, err := rt.db.GetConversation(conversationID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Se il DB non trova nulla, l'utente non fa parte della conversazione o essa non esiste (404 Not Found)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(errorResponse{Error: "Conversazione non trovata o non hai i permessi"})
			return
		}
		rt.baseLogger.WithError(err).Error("error fetching single conversation")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: "Errore interno del server"})
		return
	}

	// Restituiamo 200 OK con l'oggetto conversazione (che ora include l'array "messages")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(conv)
}

// sendMessage crea un nuovo messaggio nella conversazione.
// Endpoint: POST /conversazioni/:conversazioneID/messaggi
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	conversationID := ps.ByName("conversazioneID")

	// 1. Definiamo la struttura attesa nel body della richiesta
	// In OpenAPI abbiamo definito content, type e opzionalmente replyToMessageId
	type sendMessageRequest struct {
		Content          string `json:"content"`
		Type             string `json:"type"` // "text" o "photo"
		ReplyToMessageId string `json:"replyToMessageId,omitempty"`
	}

	var req sendMessageRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "Formato JSON non valido"})
		return
	}

	// Validazione dei campi obbligatori
	if req.Content == "" || (req.Type != "text" && req.Type != "photo") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "Content non può essere vuoto e Type deve essere 'text' o 'photo'"})
		return
	}

	// 2. Assembliamo l'oggetto Message da passare al Database
	// ID e Timestamp verranno generati dal database
	newMsg := database.Message{
		Sender:  userID,
		Content: req.Content,
		Type:    req.Type,
	}

	// 3. Chiamata al Database
	savedMsg, err := rt.db.SendMessage(newMsg, conversationID)
	if err != nil {
		// Gestiamo l'errore di sicurezza specifico restituito dal DB
		if err.Error() == "unauthorized: user not in conversation" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(errorResponse{Error: "Non fai parte di questa conversazione"})
			return
		}

		rt.baseLogger.WithError(err).Error("Errore durante il salvataggio del messaggio")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: "Errore interno del server"})
		return
	}

	// 4. Risposta di Successo
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created è il codice corretto in REST per una POST andata a buon fine
	json.NewEncoder(w).Encode(savedMsg)
}

// addToGroup aggiunge un nuovo utente ai partecipanti di un gruppo.
// Endpoint: POST /conversazioni/:conversazioneID/partecipanti
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	callerID := ctx.UserID
	conversationID := ps.ByName("conversazioneID")

	// 1. Decodifica il body JSON
	type addRequest struct {
		UserID string `json:"userId"`
	}
	var req addRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.UserID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "Devi specificare l'userId da aggiungere"})
		return
	}

	// 2. Esegui la logica nel database
	err = rt.db.AddToGroup(conversationID, req.UserID, callerID)
	if err != nil {
		if err.Error() == "unauthorized: caller is not in the group" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(errorResponse{Error: "Non sei un partecipante di questo gruppo"})
			return
		}
		if err.Error() == "not found: user to add does not exist" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(errorResponse{Error: "L'utente che stai cercando di aggiungere non esiste"})
			return
		}
		rt.baseLogger.WithError(err).Error("Errore nell'aggiunta al gruppo")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: "Errore interno del server"})
		return
	}

	// 3. Risposta di successo: 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// leaveGroup gestisce sia l'abbandono del gruppo da parte dell'utente loggato (:partecipanteID == "me"),
// sia la rimozione di un partecipante da parte del proprietario del gruppo (:partecipanteID == "<username>").
// Endpoint: DELETE /conversazioni/:conversazioneID/partecipanti/:partecipanteID
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	callerID := ctx.UserID
	conversationID := ps.ByName("conversazioneID")
	targetParticipant := ps.ByName("partecipanteID")

	// Caso 1: L'utente loggato lascia il gruppo o elimina la chat privata ("me")
	if targetParticipant == "" || targetParticipant == "me" {
		err := rt.db.LeaveGroup(conversationID, callerID)
		if err != nil {
			if err.Error() == "not found: user was not in the group or group does not exist" {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(errorResponse{Error: "Non fai parte di questa conversazione o essa non esiste"})
				return
			}
			rt.baseLogger.WithError(err).Error("Errore nell'abbandono del gruppo")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(errorResponse{Error: "Errore interno del server"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Caso 2: Il proprietario del gruppo rimuove un altro utente dal gruppo
	err := rt.db.RemoveFromGroup(conversationID, targetParticipant, callerID)
	if err != nil {
		if err.Error() == "unauthorized: caller is not the owner of the group" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(errorResponse{Error: "Solo il proprietario può rimuovere membri dal gruppo"})
			return
		}
		if err.Error() == "not found: group does not exist" || err.Error() == "not found: user to remove does not exist" || err.Error() == "not found: user was not in the group" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(errorResponse{Error: "Utente o gruppo non trovato"})
			return
		}
		if err.Error() == "bad request: owner cannot remove themselves" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse{Error: "Il proprietario non può rimuovere se stesso"})
			return
		}
		rt.baseLogger.WithError(err).Error("Errore nella rimozione del partecipante dal gruppo")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: "Errore interno del server"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// forwardMessage inoltra un messaggio esistente in una conversazione.
// Endpoint: POST /conversazioni/:conversazioneID/messaggi/forward
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	callerID := ctx.UserID
	targetConversationID := ps.ByName("conversazioneID")

	// 1. Decodifica la richiesta
	type forwardRequest struct {
		ForwardedMessageId string `json:"forwardedMessageId"`
	}
	var req forwardRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.ForwardedMessageId == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "ID del messaggio da inoltrare mancante o non valido"})
		return
	}

	// 2. Chiamata al Database
	forwardedMsg, err := rt.db.ForwardMessage(req.ForwardedMessageId, targetConversationID, callerID)
	if err != nil {
		if err.Error() == "not found: original message does not exist" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(errorResponse{Error: "Il messaggio che cerchi di inoltrare non esiste"})
			return
		}
		if err.Error() == "unauthorized: caller is not in the target group" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(errorResponse{Error: "Non fai parte della conversazione di destinazione"})
			return
		}
		rt.baseLogger.WithError(err).Error("Errore durante l'inoltro del messaggio")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: "Errore interno del server"})
		return
	}

	// 3. Risposta (201 Created come da OpenAPI)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(forwardedMsg)
}

// deleteMessage elimina un messaggio precedentemente inviato.
// Endpoint: DELETE /conversazioni/:conversazioneID/messaggi/:messaggioID
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	callerID := ctx.UserID
	messageID := ps.ByName("messaggioID")

	// 1. Chiamata al Database
	err := rt.db.DeleteMessage(messageID, callerID)
	if err != nil {
		if err.Error() == "unauthorized or not found: cannot delete message" {
			w.WriteHeader(http.StatusForbidden) // L'utente non è il mittente o il messaggio non esiste
			json.NewEncoder(w).Encode(errorResponse{Error: "Messaggio non trovato o non hai i permessi per eliminarlo"})
			return
		}
		rt.baseLogger.WithError(err).Error("Errore durante l'eliminazione del messaggio")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: "Errore interno del server"})
		return
	}

	// 2. Risposta (204 No Content)
	w.WriteHeader(http.StatusNoContent)
}

// setGroupName modifica il nome di una conversazione di gruppo (solo proprietario).
// Endpoint: PATCH /conversazioni/:conversazioneID/name
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	callerID := ctx.UserID
	conversationID := ps.ByName("conversazioneID")

	type nameReq struct {
		Name string `json:"name"`
	}
	var req nameReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "Formato JSON non valido o nome mancante"})
		return
	}

	err = rt.db.SetGroupName(conversationID, req.Name, callerID)
	if err != nil {
		if err.Error() == "unauthorized: caller is not the owner of the group" || err.Error() == "unauthorized: caller is not in the group" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(errorResponse{Error: "Solo il proprietario può modificare il nome del gruppo"})
			return
		}
		if err.Error() == "not found: group does not exist" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(errorResponse{Error: "Gruppo non trovato"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// setGroupPhoto modifica la foto di una conversazione di gruppo (solo proprietario).
// Endpoint: PATCH /conversazioni/:conversazioneID/photo
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	callerID := ctx.UserID
	conversationID := ps.ByName("conversazioneID")

	type photoReq struct {
		Photo string `json:"photo"`
	}
	var req photoReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Photo == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "Formato JSON non valido o foto mancante"})
		return
	}

	err = rt.db.SetGroupPhoto(conversationID, req.Photo, callerID)
	if err != nil {
		if err.Error() == "unauthorized: caller is not the owner of the group" || err.Error() == "unauthorized: caller is not in the group" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(errorResponse{Error: "Solo il proprietario può modificare la foto del gruppo"})
			return
		}
		if err.Error() == "not found: group does not exist" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(errorResponse{Error: "Gruppo non trovato"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// commentMessage aggiunge una reazione a un messaggio.
// Endpoint: POST /conversazioni/:conversazioneID/messaggi/:messaggioID/commenti
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	callerID := ctx.UserID
	conversationID := ps.ByName("conversazioneID")
	messageID := ps.ByName("messaggioID")

	type commentReq struct {
		Content string `json:"content"`
	}
	var req commentReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Content == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "Formato JSON non valido o contenuto mancante"})
		return
	}

	newComment := database.Comment{
		Author:  callerID,
		Content: req.Content,
	}

	err = rt.db.CommentMessage(newComment, messageID, conversationID)
	if err != nil {
		if err.Error() == "unauthorized: caller is not in the group" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(errorResponse{Error: "Non fai parte di questo gruppo"})
			return
		}
		if err.Error() == "not found: message does not exist in this conversation" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(errorResponse{Error: "Messaggio non trovato in questa conversazione"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// uncommentMessage elimina una reazione precedentemente aggiunta a un messaggio.
// Endpoint: DELETE /conversazioni/:conversazioneID/messaggi/:messaggioID/commenti/:commentoID
func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	callerID := ctx.UserID
	conversationID := ps.ByName("conversazioneID")
	messageID := ps.ByName("messaggioID")
	commentID := ps.ByName("commentoID")

	err := rt.db.UncommentMessage(commentID, messageID, conversationID, callerID)
	if err != nil {
		if err.Error() == "unauthorized or not found: cannot delete comment" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(errorResponse{Error: "Commento non trovato o non hai i permessi per eliminarlo"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// createConversation crea una nuova conversazione e aggiunge il creatore come partecipante.
// Endpoint: POST /conversazioni
func (rt *_router) createConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	creatorID := ctx.UserID

	// 1. Decodifica il body JSON. Ci aspettiamo opzionalmente un "name"
	type createReq struct {
		Name string `json:"name"`
	}
	var req createReq
	// Ignoriamo l'errore qui, perché il nome è opzionale
	_ = json.NewDecoder(r.Body).Decode(&req)

	// 2. Chiamata al Database per creare la conversazione
	// ATTENZIONE: Assicurati che il tuo database.go abbia un metodo CreateConversation(name, creatorID)
	newGroup, err := rt.db.CreateConversation(req.Name, creatorID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Errore durante la creazione della conversazione")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: "Errore interno del server"})
		return
	}

	// 3. Risposta di successo (201 Created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newGroup)
}

// updateChatStatus permette a un utente di segnalare che ha ricevuto o letto i messaggi
// Endpoint: PATCH /conversazioni/:conversazioneID/status
func (rt *_router) updateChatStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	conversationID := ps.ByName("conversazioneID")

	type statusReq struct {
		Status string `json:"status"` // 'received' o 'read'
	}
	var req statusReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Status != "received" && req.Status != "read" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "Lo status deve essere 'received' o 'read'"})
		return
	}

	err := rt.db.UpdateConversationStatus(conversationID, userID, req.Status)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Errore durante l'aggiornamento dello status")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
