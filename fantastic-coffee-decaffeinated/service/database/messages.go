package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
)

// SendMessage salva un nuovo messaggio nel database.
func (db *appdbimpl) SendMessage(msg Message, conversationID string) (Message, error) {
	// 1. Controllo di sicurezza: l'utente fa parte della conversazione?
	var exists int
	err := db.c.QueryRow(`
		SELECT 1 FROM participants 
		WHERE conversation_id = ? AND user_id = ?
	`, conversationID, msg.Sender).Scan(&exists)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// L'utente non è nei partecipanti (ritorniamo un errore specifico che gestiremo nell'API)
			return msg, errors.New("unauthorized: user not in conversation")
		}
		return msg, fmt.Errorf("error checking participant status: %w", err)
	}

	// 2. Generiamo l'ID del nuovo messaggio e il timestamp attuale
	newUUID, err := uuid.NewV4()
	if err != nil {
		return msg, fmt.Errorf("error generating message UUID: %w", err)
	}
	msg.ID = newUUID.String()

	// Salviamo l'ora esatta del server (UTC)
	msg.Timestamp = time.Now().UTC()

	// 3. Salviamo il messaggio nel Database
	// La tabella 'messages' ha: id, conversation_id, sender, content, type, reply_to, timestamp
	_, err = db.c.Exec(`
		INSERT INTO messages (id, conversation_id, sender, content, type, timestamp) 
		VALUES (?, ?, ?, ?, ?, ?)
	`, msg.ID, conversationID, msg.Sender, msg.Content, msg.Type, msg.Timestamp)

	if err != nil {
		return msg, fmt.Errorf("error inserting message: %w", err)
	}

	// Ritorniamo il messaggio completo (ora dotato di ID e Timestamp)
	return msg, nil
}

// ForwardMessage clona un messaggio esistente in una nuova conversazione.
func (db *appdbimpl) ForwardMessage(originalMessageID string, targetConversationID string, senderID string) (Message, error) {
	var forwardedMsg Message

	// 1. Controlla che il chiamante faccia parte della conversazione di destinazione
	var exists int
	err := db.c.QueryRow(`
		SELECT 1 FROM participants 
		WHERE conversation_id = ? AND user_id = ?
	`, targetConversationID, senderID).Scan(&exists)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return forwardedMsg, errors.New("unauthorized: caller is not in the target group")
		}
		return forwardedMsg, fmt.Errorf("error checking participant status: %w", err)
	}

	// 2. Recupera il contenuto e il tipo del messaggio originale
	var originalContent, originalType string
	err = db.c.QueryRow("SELECT content, type FROM messages WHERE id = ?", originalMessageID).Scan(&originalContent, &originalType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return forwardedMsg, errors.New("not found: original message does not exist")
		}
		return forwardedMsg, fmt.Errorf("error retrieving original message: %w", err)
	}

	// 3. Genera ID e Timestamp per il nuovo messaggio clonato
	newUUID, _ := uuid.NewV4()
	forwardedMsg.ID = newUUID.String()
	forwardedMsg.Sender = senderID
	forwardedMsg.Content = originalContent
	forwardedMsg.Type = originalType
	forwardedMsg.Timestamp = time.Now().UTC()

	// 4. Salva il messaggio inoltrato nel database
	_, err = db.c.Exec(`
		INSERT INTO messages (id, conversation_id, sender, content, type, timestamp) 
		VALUES (?, ?, ?, ?, ?, ?)
	`, forwardedMsg.ID, targetConversationID, forwardedMsg.Sender, forwardedMsg.Content, forwardedMsg.Type, forwardedMsg.Timestamp)

	if err != nil {
		return forwardedMsg, fmt.Errorf("error inserting forwarded message: %w", err)
	}

	return forwardedMsg, nil
}

// DeleteMessage elimina un messaggio. Un utente può eliminare solo i propri messaggi.
func (db *appdbimpl) DeleteMessage(messageID string, callerID string) error {
	// Elimina il messaggio SOLO SE il 'sender' corrisponde all'utente che sta facendo la richiesta.
	res, err := db.c.Exec("DELETE FROM messages WHERE id = ? AND sender = ?", messageID, callerID)
	if err != nil {
		return fmt.Errorf("error executing delete query: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("unauthorized or not found: cannot delete message")
	}

	return nil
}
