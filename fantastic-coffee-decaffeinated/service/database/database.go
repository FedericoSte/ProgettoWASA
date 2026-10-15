package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	Ping() error
	DoLogin(username string) (string, error)
	CheckToken(token string) (bool, error)
	GetMyConversations(userID string) ([]Conversation, error)
	GetConversation(conversationID string, userID string) (Conversation, error)
	SendMessage(msg Message, conversationID string) (Message, error)
	ForwardMessage(originalMessageID string, targetConversationID string, senderID string) (Message, error)
	DeleteMessage(messageID string, callerID string) error
	AddToGroup(conversationID string, userID string, callerID string) error
	LeaveGroup(conversationID string, userID string) error
	SetMyUserName(userID string, newName string) error
	SetMyPhoto(userID string, photoPath string) error
	// NUOVE FIRME:
	SetGroupName(conversationID string, newName string, callerID string) error
	SetGroupPhoto(conversationID string, photoPath string, callerID string) error
	CommentMessage(comment Comment, messageID string, conversationID string) error
	UncommentMessage(commentID string, messageID string, conversationID string, callerID string) error
}

type appdbimpl struct {
	c *sql.DB
}

// New initializes the database and creates tables if they don't exist.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Abilitiamo le Foreign Keys in SQLite (di default sono disabilitate)
	_, err := db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return nil, fmt.Errorf("error enabling foreign keys: %w", err)
	}

	// Controlliamo se la tabella 'users' esiste. Se non esiste, il DB è vuoto e creiamo le tabelle.
	var tableName string
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='users';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			photo TEXT
		);

		CREATE TABLE conversations (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL, -- 'single' o 'group'
			name TEXT,
			photo TEXT
		);

		CREATE TABLE participants (
			conversation_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			PRIMARY KEY (conversation_id, user_id),
			FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);

		CREATE TABLE messages (
			id TEXT PRIMARY KEY,
			conversation_id TEXT NOT NULL,
			sender TEXT NOT NULL,
			content TEXT NOT NULL,
			type TEXT NOT NULL,
			reply_to TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
			FOREIGN KEY (sender) REFERENCES users(id) ON DELETE CASCADE
		);

		CREATE TABLE comments (
			id TEXT PRIMARY KEY,
			message_id TEXT NOT NULL,
			author TEXT NOT NULL,
			content TEXT NOT NULL,
			FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
			FOREIGN KEY (author) REFERENCES users(id) ON DELETE CASCADE
		);

		CREATE TABLE message_status (
			message_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			status TEXT NOT NULL, -- 'received' o 'read'
			PRIMARY KEY (message_id, user_id),
			FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
		`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating database structure: %w", err)
		}
	}

	return &appdbimpl{
		c: db,
	}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
