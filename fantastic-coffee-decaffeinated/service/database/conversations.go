package database

import (
	"database/sql"
	"fmt"
)

// GetMyConversations ritorna la lista delle conversazioni a cui partecipa l'utente.
func (db *appdbimpl) GetMyConversations(userID string) ([]Conversation, error) {
	// Query: Prendi i dettagli della conversazione facendo JOIN con i partecipanti.
	// Inoltre, dobbiamo recuperare (con una subquery) l'ultimo messaggio per lo "snippet".
	query := `
		SELECT c.id, c.name, c.photo, 
			(SELECT content FROM messages m WHERE m.conversation_id = c.id ORDER BY timestamp DESC LIMIT 1) as last_message
		FROM conversations c
		INNER JOIN participants p ON c.id = p.conversation_id
		WHERE p.user_id = ?
	`
	rows, err := db.c.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying conversations: %w", err)
	}
	defer rows.Close()

	var convs []Conversation
	for rows.Next() {
		var conv Conversation
		var name, photo, lastMsg sql.NullString // Usiamo NullString perché potrebbero essere vuoti (NULL in DB)

		err := rows.Scan(&conv.ID, &name, &photo, &lastMsg)
		if err != nil {
			return nil, fmt.Errorf("error scanning conversation: %w", err)
		}

		if name.Valid {
			conv.Name = name.String
		}
		if photo.Valid {
			conv.Photo = photo.String
		}
		if lastMsg.Valid {
			conv.LastMessage = lastMsg.String
		}

		convs = append(convs, conv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// Se convs è vuoto (nil), restituiamo una slice vuota invece di nil per il JSON
	if convs == nil {
		convs = make([]Conversation, 0)
	}

	return convs, nil
}

// GetConversation ritorna i dettagli di una specifica conversazione e i suoi messaggi, inclusi i commenti.
func (db *appdbimpl) GetConversation(conversationID string, userID string) (Conversation, error) {
	var conv Conversation
	var name, photo sql.NullString

	// 1. Controlla se l'utente è nel gruppo e prendi i metadati
	err := db.c.QueryRow(`
		SELECT c.id, c.name, c.photo 
		FROM conversations c
		INNER JOIN participants p ON c.id = p.conversation_id
		WHERE c.id = ? AND p.user_id = ?
	`, conversationID, userID).Scan(&conv.ID, &name, &photo)

	if err != nil {
		return conv, err // Potrebbe essere sql.ErrNoRows
	}

	if name.Valid {
		conv.Name = name.String
	}
	if photo.Valid {
		conv.Photo = photo.String
	}

	// 2. Prendi tutti i messaggi della conversazione
	msgQuery := `SELECT id, sender, content, type, timestamp FROM messages WHERE conversation_id = ? ORDER BY timestamp ASC`
	rows, err := db.c.Query(msgQuery, conversationID)
	if err != nil {
		return conv, fmt.Errorf("error querying messages: %w", err)
	}
	defer rows.Close()

	// Usiamo una mappa per tenere traccia dei messaggi tramite il loro ID,
	// così possiamo associare rapidamente i commenti
	msgMap := make(map[string]*Message)
	conv.Messages = make([]Message, 0)

	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.Sender, &msg.Content, &msg.Type, &msg.Timestamp)
		if err != nil {
			return conv, fmt.Errorf("error scanning message: %w", err)
		}
		// Inizializziamo l'array dei commenti vuoto, altrimenti il JSON lo formatterebbe a null
		msg.Reactions = make([]Comment, 0)

		// Inseriamo il messaggio nell'array della conversazione
		conv.Messages = append(conv.Messages, msg)
	}

	// Dopo aver letto tutti i messaggi, popoliamo i puntatori nella mappa
	// Dobbiamo farlo DOPO aver completato l'append, altrimenti i puntatori cambiano
	for i := range conv.Messages {
		msgMap[conv.Messages[i].ID] = &conv.Messages[i]
	}

	// 3. Prendi tutti i commenti associati ai messaggi di questa conversazione
	// Facciamo un JOIN per assicurarci di prendere solo i commenti rilevanti
	commentQuery := `
		SELECT c.id, c.message_id, c.author, c.content 
		FROM comments c
		INNER JOIN messages m ON c.message_id = m.id
		WHERE m.conversation_id = ?
	`
	commentRows, err := db.c.Query(commentQuery, conversationID)
	if err != nil {
		return conv, fmt.Errorf("error querying comments: %w", err)
	}
	defer commentRows.Close()

	for commentRows.Next() {
		var c Comment
		var messageID string
		err := commentRows.Scan(&c.ID, &messageID, &c.Author, &c.Content)
		if err != nil {
			return conv, fmt.Errorf("error scanning comment: %w", err)
		}

		// Se troviamo il messaggio padre nella mappa, aggiungiamo il commento al suo array di Reactions
		if msgPtr, exists := msgMap[messageID]; exists {
			msgPtr.Reactions = append(msgPtr.Reactions, c)
		}
	}

	return conv, nil
}
