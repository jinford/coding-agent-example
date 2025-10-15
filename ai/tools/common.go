package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3/responses"
)

func GetAllToolParams() []responses.ToolUnionParam {
	return []responses.ToolUnionParam{
		GetReadFileToolParam(),
		GetListFileToolParam(),
		GetGrepFileToolParam(),
		GetWriteFileToolParam(),
		GetPatchFileToolParam(),
	}
}

func CallFunction(ctx context.Context, name string, argsJSONStr string) (string, error) {
	switch name {
	case ToolNameReadFile:
		var args ReadFileParamsJson
		if err := json.Unmarshal([]byte(argsJSONStr), &args); err != nil {
			return "", fmt.Errorf("failed to unmarshal arguments for read_file: %w", err)
		}

		result, err := ReadFile(ctx, args)
		if err != nil {
			return "", err
		}

		outJSON, err := json.Marshal(result)
		if err != nil {
			return "", fmt.Errorf("failed to marshal output for read_file: %w", err)
		}

		return string(outJSON), nil
	case ToolNameListFile:
		var args ListFileParamsJson
		if err := json.Unmarshal([]byte(argsJSONStr), &args); err != nil {
			return "", fmt.Errorf("failed to unmarshal arguments for list_file: %w", err)
		}

		result, err := ListFile(ctx, args)
		if err != nil {
			return "", err
		}

		outJSON, err := json.Marshal(result)
		if err != nil {
			return "", fmt.Errorf("failed to marshal output for list_file: %w", err)
		}

		return string(outJSON), nil
	case ToolNameGrepFile:
		var args GrepFileParamsJson
		if err := json.Unmarshal([]byte(argsJSONStr), &args); err != nil {
			return "", fmt.Errorf("failed to unmarshal arguments for grep_file: %w", err)
		}

		result, err := GrepFile(ctx, args)
		if err != nil {
			return "", err
		}

		outJSON, err := json.Marshal(result)
		if err != nil {
			return "", fmt.Errorf("failed to marshal output for grep_file: %w", err)
		}

		return string(outJSON), nil
	case ToolNameWriteFile:
		var args WriteFileParamsJson
		if err := json.Unmarshal([]byte(argsJSONStr), &args); err != nil {
			return "", fmt.Errorf("failed to unmarshal arguments for write_file: %w", err)
		}

		result, err := WriteFile(ctx, args)
		if err != nil {
			return "", err
		}

		outJSON, err := json.Marshal(result)
		if err != nil {
			return "", fmt.Errorf("failed to marshal output for write_file: %w", err)
		}

		return string(outJSON), nil
	case ToolNamePatchFile:
		var args PatchFileParamsJson
		if err := json.Unmarshal([]byte(argsJSONStr), &args); err != nil {
			return "", fmt.Errorf("failed to unmarshal arguments for patch_file: %w", err)
		}

		result, err := PatchFile(ctx, args)
		if err != nil {
			return "", err
		}

		outJSON, err := json.Marshal(result)
		if err != nil {
			return "", fmt.Errorf("failed to marshal output for patch_file: %w", err)
		}

		return string(outJSON), nil
	default:
		return "", fmt.Errorf("unknown function call: %s", name)
	}
}
