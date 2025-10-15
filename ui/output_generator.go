package ui

import (
	"context"
	"strings"
	"time"

	"github.com/jinford/coding-agent-example/session"
)

type OutputGenerator interface {
	GenerateResponse(ctx context.Context, userInput string, sessionID session.SessionID) (response string, err error)
}

type DummyOutputGenerator struct{}

func NewDummyOutputGenerator() *DummyOutputGenerator {
	return &DummyOutputGenerator{}
}

func (g *DummyOutputGenerator) GenerateResponse(ctx context.Context, userInput string, sessionID session.SessionID) (string, error) {
	time.Sleep(1 * time.Second)

	outputs := []string{
		"これはダミーの応答です。",
		"実際の実装では、ここでAIモデルからの応答します。",
		"ユーザーの入力: " + userInput,
	}

	return strings.Join(outputs, "\n"), nil
}
