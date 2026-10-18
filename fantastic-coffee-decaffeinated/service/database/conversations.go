package database

import (
	"database/sql"
	"fmt"

	"github.com/gofrs/uuid"
)

// GetMyConversations ritorna la lista delle conversazioni con il conteggio dei messaggi NON letti
func (db *appdbimpl) GetMyConversations(userID string) ([]Conversation, error) {
	query := `
		SELECT c.id, c.name, c.photo, 
			(SELECT content FROM messages m WHERE m.conversation_id = c.id ORDER BY timestamp DESC LIMIT 1) as last_message,
			(SELECT COUNT(*) FROM messages m 
			 WHERE m.conversation_id = c.id AND m.sender != ? 
			 AND m.id NOT IN (SELECT message_id FROM message_status WHERE user_id = ? AND status = 'read')
			) as unread_count
		FROM conversations c
		INNER JOIN participants p ON c.id = p.conversation_id
		WHERE p.user_id = ?
	`
	// Passiamo userID tre volte per riempire i tre "?" nella query
	rows, err := db.c.Query(query, userID, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying conversations: %w", err)
	}
	defer rows.Close()

	var convs []Conversation
	for rows.Next() {
		var conv Conversation
		var name, photo, lastMsg sql.NullString
		var unreadCount int

		err := rows.Scan(&conv.ID, &name, &photo, &lastMsg, &unreadCount)
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
		conv.UnreadCount = unreadCount

		convs = append(convs, conv)
	}

	if convs == nil {
		convs = make([]Conversation, 0)
	}
	return convs, nil
}

// GetConversation ritorna i messaggi calcolando lo stato delle spunte e unendo i nomi utente
func (db *appdbimpl) GetConversation(conversationID string, userID string) (Conversation, error) {
	var conv Conversation
	var name, photo sql.NullString

	err := db.c.QueryRow(`
		SELECT c.id, c.name, c.photo 
		FROM conversations c
		INNER JOIN participants p ON c.id = p.conversation_id
		WHERE c.id = ? AND p.user_id = ?
	`, conversationID, userID).Scan(&conv.ID, &name, &photo)

	if err != nil {
		return conv, err
	}

	if name.Valid {
		conv.Name = name.String
	}
	if photo.Valid {
		conv.Photo = photo.String
	}

	// MEGA-QUERY: Estrae i messaggi, unisce il vero nome dell'utente e calcola lo stato (sent, received, read)
	msgQuery := `
		SELECT m.id, u.name as sender, m.content, m.type, m.timestamp,
			CASE
				WHEN m.sender != ? THEN '' 
				ELSE (
					SELECT CASE
						WHEN COUNT(p.user_id) = 0 THEN 'sent'
						WHEN SUM(CASE WHEN ms.status = 'read' THEN 1 ELSE 0 END) = COUNT(p.user_id) THEN 'read'
						WHEN SUM(CASE WHEN ms.status IN ('received', 'read') THEN 1 ELSE 0 END) = COUNT(p.user_id) THEN 'received'
						ELSE 'sent'
					END
					FROM participants p
					LEFT JOIN message_status ms ON p.user_id = ms.user_id AND ms.message_id = m.id
					WHERE p.conversation_id = m.conversation_id AND p.user_id != m.sender
				)
			END as status
		FROM messages m
		JOIN users u ON m.sender = u.id
		WHERE m.conversation_id = ? 
		ORDER BY m.timestamp ASC
	`

	rows, err := db.c.Query(msgQuery, userID, conversationID)
	if err != nil {
		return conv, fmt.Errorf("error querying messages: %w", err)
	}
	defer rows.Close()

	msgMap := make(map[string]*Message)
	conv.Messages = make([]Message, 0)

	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.Sender, &msg.Content, &msg.Type, &msg.Timestamp, &msg.Status)
		if err != nil {
			return conv, fmt.Errorf("error scanning message: %w", err)
		}
		msg.Reactions = make([]Comment, 0)
		conv.Messages = append(conv.Messages, msg)
	}

	for i := range conv.Messages {
		msgMap[conv.Messages[i].ID] = &conv.Messages[i]
	}

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
		if msgPtr, exists := msgMap[messageID]; exists {
			msgPtr.Reactions = append(msgPtr.Reactions, c)
		}
	}

	return conv, nil
}

// UpdateConversationStatus aggiorna lo stato ('received' o 'read') per i messaggi altrui nella chat
func (db *appdbimpl) UpdateConversationStatus(conversationID string, userID string, status string) error {
	query := `
		INSERT INTO message_status (message_id, user_id, status)
		SELECT id, ?, ?
		FROM messages
		WHERE conversation_id = ? AND sender != ?
		ON CONFLICT(message_id, user_id) DO UPDATE SET status = ?
	`
	_, err := db.c.Exec(query, userID, status, conversationID, userID, status)
	return err
}

// CreateConversation crea un nuovo record nella tabella conversations e aggiunge il creatore in participants.
func (db *appdbimpl) CreateConversation(name string, creatorID string) (Conversation, error) {
	// Genera un nuovo UUID per la conversazione
	u, err := uuid.NewV4()
	if err != nil {
		return Conversation{}, fmt.Errorf("failed to generate UUID: %w", err)
	}
	newID := u.String()

	tx, err := db.c.Begin()
	if err != nil {
		return Conversation{}, fmt.Errorf("failed to begin transaction: %w", err)
	}

	chatType := "group"
	_, err = tx.Exec("INSERT INTO conversations (id, name, type) VALUES (?, ?, ?)", newID, name, chatType)
	if err != nil {
		tx.Rollback()
		return Conversation{}, fmt.Errorf("failed to insert conversation: %w", err)
	}

	// Aggiungi il creatore alla tabella partecipanti
	_, err = tx.Exec("INSERT INTO participants (conversation_id, user_id) VALUES (?, ?)", newID, creatorID)
	if err != nil {
		tx.Rollback()
		return Conversation{}, fmt.Errorf("failed to insert participant: %w", err)
	}

	// Conferma le scritture sul DB
	err = tx.Commit()
	if err != nil {
		return Conversation{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	newConv := Conversation{
		ID:       newID,
		Name:     name,
		Messages: make([]Message, 0),
	}

	return newConv, nil
}
