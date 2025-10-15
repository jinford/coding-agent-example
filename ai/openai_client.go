package ai

import (
	"context"
	"fmt"

	"github.com/jinford/coding-agent-example/ai/tools"
	"github.com/jinford/coding-agent-example/session"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
)

type OpenAIClient struct {
	client       openai.Client
	config       *Config
	sessionStore session.Store
}

func NewOpenAIClient(apiKey string, sessionStore session.Store, opts ...OptionFunc) *OpenAIClient {
	config := defaultConfig()
	for _, f := range opts {
		f(config)
	}

	return &OpenAIClient{
		client:       openai.NewClient(option.WithAPIKey(apiKey)),
		config:       config,
		sessionStore: sessionStore,
	}
}

// GenerateResponse implements ui.OutputGenerator.
func (c *OpenAIClient) GenerateResponse(ctx context.Context, userInput string, sessionID session.SessionID) (string, error) {
	// セッションから会話履歴を取得
	conversationHistory, err := c.sessionStore.List(sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to list conversation history: %w", err)
	}

	// 前回のresponse IDを取得（最後のassistantターンのMetadataから）
	var previousResponseID string
	for i := len(conversationHistory) - 1; i >= 0; i-- {
		if conversationHistory[i].Role == "assistant" {
			if respID, ok := conversationHistory[i].Metadata["previous_response_id"]; ok {
				previousResponseID = respID
				break
			}
		}
	}

	// response ID が有効か確認
	if previousResponseID != "" {
		if _, err := c.client.Responses.Get(ctx, previousResponseID, responses.ResponseGetParams{}); err != nil {
			previousResponseID = ""
		}
	}

	params := responses.ResponseNewParams{
		Model:        shared.ChatModelGPT4_1,
		Instructions: openai.String(systemPrompt),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: responses.ResponseInputParam{
				responses.ResponseInputItemParamOfMessage(userInput, responses.EasyInputMessageRoleUser),
			},
		},
		Tools: tools.GetAllToolParams(),
	}

	if previousResponseID != "" {
		params.PreviousResponseID = openai.String(previousResponseID)
	}

	resp, err := c.client.Responses.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to call response API: %w", err)
	}

	responseText, toolCalls, lastResponseID, err := c.resolveToolCalls(ctx, resp)
	if err != nil {
		return "", err
	}

	// ユーザーのターンを追加
	userTurn := &session.ConversationTurn{
		Role:    "user",
		Content: userInput,
	}
	if err := c.sessionStore.Append(sessionID, userTurn); err != nil {
		return "", fmt.Errorf("failed to append user turn: %w", err)
	}

	// アシスタントのターンを追加
	assistantTurn := &session.ConversationTurn{
		Role:      "assistant",
		Content:   responseText,
		ToolCalls: toolCalls,
		Metadata: map[string]string{
			"previous_response_id": lastResponseID,
		},
	}
	if err := c.sessionStore.Append(sessionID, assistantTurn); err != nil {
		return "", fmt.Errorf("failed to append assistant turn: %w", err)
	}

	return responseText, nil
}

func (c *OpenAIClient) resolveToolCalls(ctx context.Context, resp *responses.Response) (string, []session.ToolCall, string, error) {
	// ツール呼び出し情報を記録
	var toolCalls []session.ToolCall

	// ツールコールがあるか確認し、あれば実行して結果を取得
	toolOutputs := make([]responses.ResponseInputItemUnionParam, 0, len(resp.Output))
	for _, outputItem := range resp.Output {
		if outputItem.Type != "function_call" {
			continue
		}

		item := outputItem.AsFunctionCall()
		result, err := c.handleFunctionCall(ctx, item)
		if err != nil {
			// エラーの場合も結果として返す
			result = fmt.Sprintf("Error: %v", err)
			toolOutputs = append(toolOutputs, responses.ResponseInputItemParamOfFunctionCallOutput(
				item.CallID,
				result,
			))
		} else {
			// 実行結果を次のプロンプトに含める
			toolOutputs = append(toolOutputs, responses.ResponseInputItemParamOfFunctionCallOutput(
				item.CallID,
				result,
			))
		}

		// ツール呼び出し情報を記録
		toolCalls = append(toolCalls, session.ToolCall{
			Name:      item.Name,
			Arguments: item.Arguments,
			Result:    result,
		})
	}

	// ツールコールがあった場合は、再度APIを呼び出して結果を返す
	if len(toolOutputs) > 0 {
		nextResp, err := c.client.Responses.New(ctx, responses.ResponseNewParams{
			Model:        shared.ChatModelGPT4_1,
			Instructions: openai.String(systemPrompt),
			Input: responses.ResponseNewParamsInputUnion{
				OfInputItemList: toolOutputs,
			},
			Tools:              tools.GetAllToolParams(),
			PreviousResponseID: openai.String(resp.ID),
		})
		if err != nil {
			return "", toolCalls, "", fmt.Errorf("failed to call response API: %w", err)
		}

		// 再帰的に処理（ツール呼び出し情報を引き継ぐ）
		nextText, nextToolCalls, lastRespID, err := c.resolveToolCalls(ctx, nextResp)
		if err != nil {
			return "", toolCalls, "", err
		}
		// ツール呼び出し情報をマージ
		allToolCalls := append(toolCalls, nextToolCalls...)
		return nextText, allToolCalls, lastRespID, nil
	}

	// 最終的な応答を返す
	return resp.OutputText(), toolCalls, resp.ID, nil
}

func (c *OpenAIClient) handleFunctionCall(ctx context.Context, item responses.ResponseFunctionToolCall) (string, error) {
	fmt.Fprintf(c.config.debugOutput, "function called: %s\n", item.Name)
	return tools.CallFunction(ctx, item.Name, item.Arguments)
}
