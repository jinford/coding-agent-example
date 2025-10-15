package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/jinford/coding-agent-example/session"
)

type Conversation struct {
	inputScanner    InputScanner
	outputGenerator OutputGenerator
	printer         *Printer
	currentSession  session.SessionID
}

func NewConversation(inputScanner InputScanner, outputGenerator OutputGenerator) *Conversation {
	return &Conversation{
		inputScanner:    inputScanner,
		outputGenerator: outputGenerator,
		printer:         NewPrinter(),
	}
}

func (c *Conversation) Run(ctx context.Context) {
	// Welcome メッセージを表示
	c.printer.PrintWelcome()

	// セッションIDを初期化（最初の1回だけ）
	if c.currentSession.IsEmpty() {
		c.currentSession = session.NewSessionID()
	}

	// ユーザー入力用のチャンネル
	inputChan := make(chan string)

	// コンテキストがキャンセルされたら処理を抜けられるようにユーザーの入力をチャネルで処理
	go func() {
		defer close(inputChan)
		for {
			// ユーザー入力を取得
			if !c.inputScanner.Scan() {
				return
			}
			userInput := strings.TrimSpace(c.inputScanner.Text())

			select {
			case inputChan <- userInput:
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		// プロンプト表示
		c.printer.PrintPrompt()

		// ユーザー入力またはコンテキストキャンセルを待つ
		select {
		case userInput, ok := <-inputChan:
			if !ok {
				return
			}

			// 終了コマンドかチェック
			if userInput == "/exit" {
				return
			}

			if userInput == "" {
				continue
			}

			// API呼び出し中の表示
			stopThinking := c.printer.StartThinking()

			// 応答を生成
			out, err := c.outputGenerator.GenerateResponse(ctx, userInput, c.currentSession)
			stopThinking()

			// エラーがあれば表示して次の入力へ
			if err != nil {
				c.printer.PrintErrorMessage(err.Error())
				continue
			}

			// アシスタントの応答を表示
			c.printer.PrintAssistantMessage(out)

			// 応答後に改行
			fmt.Println()

		case <-ctx.Done():
			return
		}
	}
}
