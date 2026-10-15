package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gofrs/uuid"
)

// CommentMessage aggiunge una reazione a un messaggio.
func (db *appdbimpl) CommentMessage(comment Comment, messageID string, conversationID string) error {
	// Controllo sicurezza: il chiamante deve essere nel gruppo
	var exists int
	err := db.c.QueryRow("SELECT 1 FROM participants WHERE conversation_id = ? AND user_id = ?", conversationID, comment.Author).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("unauthorized: caller is not in the group")
		}
		return fmt.Errorf("error checking permissions: %w", err)
	}

	// Controllo validità: il messaggio esiste ed appartiene a questa conversazione?
	var msgExists int
	err = db.c.QueryRow("SELECT 1 FROM messages WHERE id = ? AND conversation_id = ?", messageID, conversationID).Scan(&msgExists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("not found: message does not exist in this conversation")
		}
		return fmt.Errorf("error checking message: %w", err)
	}

	// Genera ID per il commento
	newUUID, _ := uuid.NewV4()
	comment.ID = newUUID.String()

	// Inserisce il commento
	_, err = db.c.Exec(`
		INSERT INTO comments (id, message_id, author, content) 
		VALUES (?, ?, ?, ?)
	`, comment.ID, messageID, comment.Author, comment.Content)

	if err != nil {
		return fmt.Errorf("error inserting comment: %w", err)
	}
	return nil
}

// UncommentMessage rimuove una reazione da un messaggio.
func (db *appdbimpl) UncommentMessage(commentID string, messageID string, conversationID string, callerID string) error {
	// Un utente può eliminare solo le proprie reazioni
	res, err := db.c.Exec("DELETE FROM comments WHERE id = ? AND message_id = ? AND author = ?", commentID, messageID, callerID)
	if err != nil {
		return fmt.Errorf("error deleting comment: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("unauthorized or not found: cannot delete comment")
	}
	return nil
}
