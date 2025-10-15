package session

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore はSQLiteを使用してセッションを保存する実装
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore は新しいSQLiteStoreを作成する
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// テーブルを作成
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS conversation_turns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			tool_calls TEXT,
			metadata TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_session_id ON conversation_turns(session_id);
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

// Close はデータベース接続を閉じる
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// List はセッションIDから会話履歴を取得する
func (s *SQLiteStore) List(sessionID SessionID) ([]*ConversationTurn, error) {
	rows, err := s.db.Query(`
		SELECT role, content, tool_calls, metadata
		FROM conversation_turns
		WHERE session_id = ?
		ORDER BY id ASC
	`, sessionID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query turns: %w", err)
	}
	defer rows.Close()

	var turns []*ConversationTurn
	for rows.Next() {
		var (
			role         string
			content      string
			toolCallsStr sql.NullString
			metadataStr  sql.NullString
		)

		if err := rows.Scan(&role, &content, &toolCallsStr, &metadataStr); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		turn := &ConversationTurn{
			Role:    role,
			Content: content,
		}

		// ToolCallsをデシリアライズ
		if toolCallsStr.Valid && toolCallsStr.String != "" {
			if err := json.Unmarshal([]byte(toolCallsStr.String), &turn.ToolCalls); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tool_calls: %w", err)
			}
		}

		// Metadataをデシリアライズ
		if metadataStr.Valid && metadataStr.String != "" {
			if err := json.Unmarshal([]byte(metadataStr.String), &turn.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		turns = append(turns, turn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return turns, nil
}

// Append は会話履歴に新しいターンを追加する
func (s *SQLiteStore) Append(sessionID SessionID, turn *ConversationTurn) error {
	// ToolCallsをシリアライズ
	var toolCallsStr sql.NullString
	if len(turn.ToolCalls) > 0 {
		toolCallsJSON, err := json.Marshal(turn.ToolCalls)
		if err != nil {
			return fmt.Errorf("failed to marshal tool_calls: %w", err)
		}
		toolCallsStr = sql.NullString{String: string(toolCallsJSON), Valid: true}
	}

	// Metadataをシリアライズ
	var metadataStr sql.NullString
	if len(turn.Metadata) > 0 {
		metadataJSON, err := json.Marshal(turn.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataStr = sql.NullString{String: string(metadataJSON), Valid: true}
	}

	_, err := s.db.Exec(`
		INSERT INTO conversation_turns (session_id, role, content, tool_calls, metadata)
		VALUES (?, ?, ?, ?, ?)
	`, sessionID.String(), turn.Role, turn.Content, toolCallsStr, metadataStr)
	if err != nil {
		return fmt.Errorf("failed to insert turn: %w", err)
	}

	return nil
}

// Delete はセッションを削除する
func (s *SQLiteStore) Delete(sessionID SessionID) error {
	_, err := s.db.Exec(`
		DELETE FROM conversation_turns
		WHERE session_id = ?
	`, sessionID.String())
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}
