package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
)

// GetMyConversations ritorna la lista delle conversazioni con il conteggio dei messaggi NON letti
func (db *appdbimpl) GetMyConversations(userID string) ([]Conversation, error) {
	query := `
		SELECT c.id, c.name, c.photo, u_owner.name,
			(SELECT 
				CASE 
					WHEN m.type = 'system' AND m.sender = ? AND m.content LIKE 'ha %' THEN 'Hai ' || SUBSTR(m.content, 4)
					WHEN m.type = 'system' AND m.content LIKE 'ha %' THEN u_msg.name || ' ' || m.content
					ELSE m.content
				END
			 FROM messages m
			 JOIN users u_msg ON m.sender = u_msg.id
			 WHERE m.conversation_id = c.id
			 ORDER BY m.timestamp DESC LIMIT 1) as last_message,
			(SELECT COUNT(*) FROM messages m 
			 WHERE m.conversation_id = c.id AND m.sender != ? AND m.type != 'system'
			 AND m.id NOT IN (SELECT message_id FROM message_status WHERE user_id = ? AND status = 'read')
			) as unread_count
		FROM conversations c
		INNER JOIN participants p ON c.id = p.conversation_id
		LEFT JOIN users u_owner ON c.owner = u_owner.id
		WHERE p.user_id = ?
	`
	// Passiamo userID quattro volte per riempire i quattro "?" nella query
	rows, err := db.c.Query(query, userID, userID, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying conversations: %w", err)
	}
	defer rows.Close()

	var convs []Conversation
	for rows.Next() {
		var conv Conversation
		var name, photo, ownerName, lastMsg sql.NullString
		var unreadCount int

		err := rows.Scan(&conv.ID, &name, &photo, &ownerName, &lastMsg, &unreadCount)
		if err != nil {
			return nil, fmt.Errorf("error scanning conversation: %w", err)
		}

		if name.Valid {
			conv.Name = name.String
		}
		if photo.Valid {
			conv.Photo = photo.String
		}
		if ownerName.Valid {
			conv.Owner = ownerName.String
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

// GetConversation ritorna i dettagli della conversazione, i suoi partecipanti, il proprietario e i messaggi
func (db *appdbimpl) GetConversation(conversationID string, userID string) (Conversation, error) {
	var conv Conversation
	var name, photo, ownerName sql.NullString

	err := db.c.QueryRow(`
		SELECT c.id, c.name, c.photo, u_owner.name
		FROM conversations c
		INNER JOIN participants p ON c.id = p.conversation_id
		LEFT JOIN users u_owner ON c.owner = u_owner.id
		WHERE c.id = ? AND p.user_id = ?
	`, conversationID, userID).Scan(&conv.ID, &name, &photo, &ownerName)

	if err != nil {
		return conv, err
	}

	if name.Valid {
		conv.Name = name.String
	}
	if photo.Valid {
		conv.Photo = photo.String
	}
	if ownerName.Valid {
		conv.Owner = ownerName.String
	}

	// Estrazione della lista reale dei partecipanti dalla tabella participants
	memberRows, err := db.c.Query(`
		SELECT u.name
		FROM participants p
		INNER JOIN users u ON p.user_id = u.id
		LEFT JOIN conversations c ON c.id = p.conversation_id
		WHERE p.conversation_id = ?
		ORDER BY CASE WHEN p.user_id = c.owner THEN 0 ELSE 1 END, p.rowid ASC
	`, conversationID)
	if err != nil {
		return conv, fmt.Errorf("error querying conversation members: %w", err)
	}
	defer memberRows.Close()

	conv.Members = make([]string, 0)
	for memberRows.Next() {
		var memberName string
		if err := memberRows.Scan(&memberName); err != nil {
			return conv, fmt.Errorf("error scanning member name: %w", err)
		}
		conv.Members = append(conv.Members, memberName)
	}

	// MEGA-QUERY: Estrae i messaggi, unisce il vero nome dell'utente e calcola lo stato (sent, received, read)
	msgQuery := `
		SELECT m.id, u.name as sender, m.content, m.type, m.timestamp,
			CASE
				WHEN m.type = 'system' OR m.sender != ? THEN '' 
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
		WHERE conversation_id = ? AND sender != ? AND type != 'system'
		ON CONFLICT(message_id, user_id) DO UPDATE SET status = ?
	`
	_, err := db.c.Exec(query, userID, status, conversationID, userID, status)
	return err
}

// CreateConversation crea un nuovo record nella tabella conversations impostando creatorID come owner
// e aggiunge il creatore in participants.
func (db *appdbimpl) CreateConversation(name string, creatorID string) (Conversation, error) {
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
	_, err = tx.Exec("INSERT INTO conversations (id, name, type, owner) VALUES (?, ?, ?, ?)", newID, name, chatType, creatorID)
	if err != nil {
		tx.Rollback()
		return Conversation{}, fmt.Errorf("failed to insert conversation: %w", err)
	}

	_, err = tx.Exec("INSERT INTO participants (conversation_id, user_id) VALUES (?, ?)", newID, creatorID)
	if err != nil {
		tx.Rollback()
		return Conversation{}, fmt.Errorf("failed to insert participant: %w", err)
	}

	var creatorName string
	_ = tx.QueryRow("SELECT name FROM users WHERE id = ?", creatorID).Scan(&creatorName)

	err = tx.Commit()
	if err != nil {
		return Conversation{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Se è un gruppo (non una chat privata), inseriamo il messaggio di sistema di creazione gruppo
	if !strings.Contains(name, "|") {
		_ = db.insertSystemMessage(newID, creatorID, fmt.Sprintf("ha creato il gruppo \"%s\".", name))
	}

	members := make([]string, 0)
	if creatorName != "" {
		members = append(members, creatorName)
	}

	newConv := Conversation{
		ID:       newID,
		Name:     name,
		Owner:    creatorName,
		Members:  members,
		Messages: make([]Message, 0),
	}

	return newConv, nil
}
