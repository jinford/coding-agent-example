package session

import (
	"errors"
	"sync"

	"github.com/google/uuid"
)

// SessionID はセッションを識別するための型安全なID
type SessionID string

// NewSessionID は新しいセッションIDを生成する
func NewSessionID() SessionID {
	id := uuid.New()
	return SessionID(id.String())
}

// String はSessionIDを文字列に変換する
func (s SessionID) String() string {
	return string(s)
}

// IsEmpty はSessionIDが空かどうかを判定する
func (s SessionID) IsEmpty() bool {
	return s == ""
}

// ErrSessionNotFound はセッションが見つからない場合のエラー
var ErrSessionNotFound = errors.New("session not found")

// ToolCall はツール呼び出し情報を表す
type ToolCall struct {
	Name      string `json:"name"`      // ツール名
	Arguments string `json:"arguments"` // 引数（JSON文字列）
	Result    string `json:"result"`    // 実行結果
}

// ConversationTurn は会話のターン（ユーザーまたはアシスタントの発言）を表す
type ConversationTurn struct {
	Role      string            `json:"role"`                 // "user", "assistant", "tool"
	Content   string            `json:"content"`              // 発言内容
	ToolCalls []ToolCall        `json:"tool_calls,omitempty"` // ツール呼び出し（assistantロールの場合）
	Metadata  map[string]string `json:"metadata,omitempty"`   // ベンダー固有のメタデータ
}

// Store はセッションデータを保存・取得するインターフェース
type Store interface {
	// List はセッションIDから会話履歴を取得する
	List(sessionID SessionID) ([]*ConversationTurn, error)

	// Append は会話履歴に新しいターンを追加する
	Append(sessionID SessionID, turn *ConversationTurn) error

	// Delete はセッションを削除する
	Delete(sessionID SessionID) error
}

// InMemoryStore はメモリ内にセッションを保存する実装
type InMemoryStore struct {
	mu   sync.RWMutex
	data map[SessionID][]*ConversationTurn
}

// NewInMemoryStore は新しいInMemoryStoreを作成する
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[SessionID][]*ConversationTurn),
	}
}

// List はセッションIDから会話履歴を取得する
func (s *InMemoryStore) List(sessionID SessionID) ([]*ConversationTurn, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	turns, ok := s.data[sessionID]
	if !ok {
		// セッションが存在しない場合は空のスライスを返す
		return []*ConversationTurn{}, nil
	}

	// コピーを返す（元のデータを保護）
	result := make([]*ConversationTurn, len(turns))
	copy(result, turns)

	return result, nil
}

// Append は会話履歴に新しいターンを追加する
func (s *InMemoryStore) Append(sessionID SessionID, turn *ConversationTurn) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[sessionID] = append(s.data[sessionID], turn)
	return nil
}

// Delete はセッションを削除する
func (s *InMemoryStore) Delete(sessionID SessionID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, sessionID)
	return nil
}
