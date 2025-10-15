package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jinford/coding-agent-example/ai"
	"github.com/jinford/coding-agent-example/session"
	"github.com/jinford/coding-agent-example/ui"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 環境変数からAPIキーを取得
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable is not set")
		fmt.Println("Please set your OpenAI API key: export OPENAI_API_KEY=your_api_key_here")
		os.Exit(1)
	}

	// セッションストアを初期化（SQLite）
	sessionStore, err := session.NewSQLiteStore("./sessions.db")
	if err != nil {
		fmt.Printf("Error: Failed to initialize session store: %v\n", err)
		os.Exit(1)
	}
	defer sessionStore.Close()

	// OpenAIクライアントを初期化
	client := ai.NewOpenAIClient(apiKey, sessionStore, ai.WithDebugOutput(os.Stdout))

	// UIコンポーネントを初期化
	conversation := ui.NewConversation(bufio.NewScanner(os.Stdin), client)

	// 会話を開始
	conversation.Run(ctx)
}
