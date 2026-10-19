package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

// insertSystemMessage inserisce automaticamente un messaggio speciale di sistema ('system') nella conversazione.
func (db *appdbimpl) insertSystemMessage(conversationID string, actorID string, content string) error {
	newUUID, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("error generating system message UUID: %w", err)
	}
	now := time.Now().UTC()

	_, err = db.c.Exec(`
		INSERT INTO messages (id, conversation_id, sender, content, type, timestamp) 
		VALUES (?, ?, ?, ?, 'system', ?)
	`, newUUID.String(), conversationID, actorID, content, now)

	return err
}

// AddToGroup aggiunge un utente a un gruppo esistente e genera il messaggio di sistema.
func (db *appdbimpl) AddToGroup(conversationID string, userToAdd string, callerID string) error {
	// 1. Controllo sicurezza: il chiamante deve essere già nel gruppo per poter aggiungere altri
	var convName sql.NullString
	err := db.c.QueryRow(`
		SELECT c.name
		FROM conversations c
		INNER JOIN participants p ON c.id = p.conversation_id
		WHERE c.id = ? AND p.user_id = ?
	`, conversationID, callerID).Scan(&convName)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("unauthorized: caller is not in the group")
		}
		return fmt.Errorf("error checking caller permissions: %w", err)
	}

	// 2. Controllo validità: l'utente da aggiungere esiste?
	var userExists int
	err = db.c.QueryRow("SELECT 1 FROM users WHERE name = ?", userToAdd).Scan(&userExists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("not found: user to add does not exist")
		}
		return fmt.Errorf("error checking if user exists: %w", err)
	}

	// 3. Aggiungiamo l'utente al gruppo
	res, err := db.c.Exec(`
		INSERT OR IGNORE INTO participants (conversation_id, user_id) 
		VALUES (?, (SELECT id FROM users WHERE name = ?))
	`, conversationID, userToAdd)

	if err != nil {
		return fmt.Errorf("error adding user to group: %w", err)
	}

	rowsAffected, _ := res.RowsAffected()
	isPrivateChat := convName.Valid && strings.Contains(convName.String, "|")

	// 4. Se è un gruppo (non una chat privata) e l'utente non era già presente, inviamo il messaggio di sistema
	if rowsAffected > 0 && !isPrivateChat {
		_ = db.insertSystemMessage(conversationID, callerID, fmt.Sprintf("ha aggiunto %s.", userToAdd))
	}

	return nil
}

// RemoveFromGroup permette al proprietario del gruppo di rimuovere un partecipante e genera il messaggio di sistema.
func (db *appdbimpl) RemoveFromGroup(conversationID string, userToRemove string, callerID string) error {
	var ownerID, convName sql.NullString
	err := db.c.QueryRow(`
		SELECT owner, name FROM conversations WHERE id = ? AND type = 'group'
	`, conversationID).Scan(&ownerID, &convName)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("not found: group does not exist")
		}
		return fmt.Errorf("error checking group owner: %w", err)
	}

	if !ownerID.Valid || ownerID.String != callerID {
		return errors.New("unauthorized: caller is not the owner of the group")
	}

	var targetUserID string
	err = db.c.QueryRow("SELECT id FROM users WHERE name = ?", userToRemove).Scan(&targetUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("not found: user to remove does not exist")
		}
		return fmt.Errorf("error finding user to remove: %w", err)
	}

	if targetUserID == callerID {
		return errors.New("bad request: owner cannot remove themselves")
	}

	res, err := db.c.Exec(`
		DELETE FROM participants WHERE conversation_id = ? AND user_id = ?
	`, conversationID, targetUserID)
	if err != nil {
		return fmt.Errorf("error removing user from group: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("not found: user was not in the group")
	}

	isPrivateChat := convName.Valid && strings.Contains(convName.String, "|")
	if !isPrivateChat {
		_ = db.insertSystemMessage(conversationID, callerID, fmt.Sprintf("ha rimosso %s.", userToRemove))
	}

	return nil
}

// LeaveGroup rimuove l'utente dalla conversazione.
// Se si tratta di un gruppo e l'utente che esce è il proprietario (owner),
// l'intero gruppo viene eliminato per tutti i partecipanti.
// Se esce un membro non proprietario, viene generato il messaggio di sistema "ha lasciato il gruppo.".
func (db *appdbimpl) LeaveGroup(conversationID string, userID string) error {
	var ownerID, convName sql.NullString
	err := db.c.QueryRow(`
		SELECT c.owner, c.name
		FROM conversations c
		INNER JOIN participants p ON c.id = p.conversation_id
		WHERE c.id = ? AND p.user_id = ?
	`, conversationID, userID).Scan(&ownerID, &convName)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("not found: user was not in the group or group does not exist")
		}
		return fmt.Errorf("error checking group participation: %w", err)
	}

	isPrivateChat := convName.Valid && strings.Contains(convName.String, "|")

	// Se è un gruppo e chi esce è il proprietario, cancelliamo l'intero gruppo per tutti
	if !isPrivateChat && ownerID.Valid && ownerID.String == userID {
		tx, err := db.c.Begin()
		if err != nil {
			return fmt.Errorf("error starting transaction: %w", err)
		}

		_, _ = tx.Exec(`DELETE FROM comments WHERE message_id IN (SELECT id FROM messages WHERE conversation_id = ?)`, conversationID)
		_, _ = tx.Exec(`DELETE FROM message_status WHERE message_id IN (SELECT id FROM messages WHERE conversation_id = ?)`, conversationID)
		_, _ = tx.Exec(`DELETE FROM messages WHERE conversation_id = ?`, conversationID)
		_, _ = tx.Exec(`DELETE FROM participants WHERE conversation_id = ?`, conversationID)

		if _, err := tx.Exec(`DELETE FROM conversations WHERE id = ?`, conversationID); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("error deleting group conversation: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("error committing group deletion: %w", err)
		}
		return nil
	}

	// Altrimenti (membro non proprietario o uscita da chat privata), rimuoviamo solo l'utente dai partecipanti
	res, err := db.c.Exec(`
		DELETE FROM participants 
		WHERE conversation_id = ? AND user_id = ?
	`, conversationID, userID)

	if err != nil {
		return fmt.Errorf("error leaving group: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("not found: user was not in the group or group does not exist")
	}

	// Controlliamo quanti partecipanti restano
	var remaining int
	if err := db.c.QueryRow(`SELECT COUNT(*) FROM participants WHERE conversation_id = ?`, conversationID).Scan(&remaining); err == nil {
		if remaining == 0 {
			_, _ = db.c.Exec(`DELETE FROM comments WHERE message_id IN (SELECT id FROM messages WHERE conversation_id = ?)`, conversationID)
			_, _ = db.c.Exec(`DELETE FROM message_status WHERE message_id IN (SELECT id FROM messages WHERE conversation_id = ?)`, conversationID)
			_, _ = db.c.Exec(`DELETE FROM messages WHERE conversation_id = ?`, conversationID)
			_, _ = db.c.Exec(`DELETE FROM conversations WHERE id = ?`, conversationID)
		} else if !isPrivateChat {
			// Se il gruppo esiste ancora ed è un gruppo, inseriamo il messaggio speciale di uscita
			_ = db.insertSystemMessage(conversationID, userID, "ha lasciato il gruppo.")
		}
	}

	return nil
}

// SetGroupName aggiorna il nome di un gruppo (solo proprietario) e genera il messaggio di sistema.
func (db *appdbimpl) SetGroupName(conversationID string, newName string, callerID string) error {
	var ownerID sql.NullString
	err := db.c.QueryRow("SELECT owner FROM conversations WHERE id = ? AND type = 'group'", conversationID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("not found: group does not exist")
		}
		return fmt.Errorf("error checking group owner: %w", err)
	}

	if !ownerID.Valid || ownerID.String != callerID {
		return errors.New("unauthorized: caller is not the owner of the group")
	}

	_, err = db.c.Exec("UPDATE conversations SET name = ? WHERE id = ? AND type = 'group'", newName, conversationID)
	if err != nil {
		return fmt.Errorf("error updating group name: %w", err)
	}

	_ = db.insertSystemMessage(conversationID, callerID, fmt.Sprintf("ha cambiato il nome del gruppo in \"%s\".", newName))
	return nil
}

// SetGroupPhoto aggiorna la foto di un gruppo (solo proprietario) e genera il messaggio di sistema.
func (db *appdbimpl) SetGroupPhoto(conversationID string, photo string, callerID string) error {
	var ownerID sql.NullString
	err := db.c.QueryRow("SELECT owner FROM conversations WHERE id = ? AND type = 'group'", conversationID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("not found: group does not exist")
		}
		return fmt.Errorf("error checking group owner: %w", err)
	}

	if !ownerID.Valid || ownerID.String != callerID {
		return errors.New("unauthorized: caller is not the owner of the group")
	}

	_, err = db.c.Exec("UPDATE conversations SET photo = ? WHERE id = ? AND type = 'group'", photo, conversationID)
	if err != nil {
		return fmt.Errorf("error updating group photo: %w", err)
	}

	_ = db.insertSystemMessage(conversationID, callerID, "ha cambiato l'immagine del gruppo.")
	return nil
}
