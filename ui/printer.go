package ui

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

type Printer struct {
	userColor      *color.Color
	assistantColor *color.Color
	systemColor    *color.Color
	errorColor     *color.Color
	promptColor    *color.Color
	separatorColor *color.Color
	headerColor    *color.Color

	spinner *spinner.Spinner
}

func NewPrinter() *Printer {
	return &Printer{
		userColor:      color.New(color.FgCyan, color.Bold),
		assistantColor: color.New(color.FgGreen),
		systemColor:    color.New(color.FgYellow),
		errorColor:     color.New(color.FgRed, color.Bold),
		promptColor:    color.New(color.FgCyan, color.Bold),
		separatorColor: color.New(color.FgHiBlack),
		headerColor:    color.New(color.FgHiCyan, color.Bold),
		spinner:        spinner.New(spinner.CharSets[14], 100*time.Millisecond),
	}
}

// ウェルカムメッセージを表示
func (p *Printer) PrintWelcome() {
	fmt.Println()
	p.headerColor.Println("╔════════════════════════════════════════╗")
	p.headerColor.Println("║     Coding Agent CLI            ║")
	p.headerColor.Println("╚════════════════════════════════════════╝")
	fmt.Println()
	p.systemColor.Println("💡 使い方:")
	fmt.Println("  • 質問や指示を入力してEnterキーを押してください")
	fmt.Println("  • '/exit' で終了します")
	fmt.Println()
	p.separatorColor.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// ユーザー入力プロンプトを表示
func (p *Printer) PrintPrompt() {
	fmt.Println()
	p.promptColor.Print("❯ ")
}

// 考え中メッセージを表示
func (p *Printer) StartThinking() func() {
	p.spinner.Start()
	return func() {
		p.spinner.Stop()
		p.ClearLine()
	}
}

// 現在の行をクリア
func (p *Printer) ClearLine() {
	fmt.Print("\r\033[K")
}

// アシスタントのメッセージを表示
func (p *Printer) PrintAssistantMessage(message string) {
	p.assistantColor.Println(message)
}

// エラーメッセージを表示
func (p *Printer) PrintErrorMessage(message string) {
	p.errorColor.Printf("✗ エラー: %v\n", message)
}

// PrintSeparator 区切り線を表示
func (p *Printer) PrintSeparator() {
	p.separatorColor.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}
