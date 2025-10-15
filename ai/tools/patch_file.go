package tools

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

//go:generate go tool go-jsonschema -p tools -o patch_file_params_gen.go patch_file_params.json

//go:embed patch_file_params.json
var patchFileParamsJSONSchema string

var getPatchFileParamsOnce = sync.OnceValue(func() openai.FunctionParameters {
	var params openai.FunctionParameters
	_ = json.Unmarshal([]byte(patchFileParamsJSONSchema), &params)
	return params
})

const ToolNamePatchFile = "patch_file"

func GetPatchFileToolParam() responses.ToolUnionParam {
	return responses.ToolUnionParam{
		OfFunction: &responses.FunctionToolParam{
			Name:        ToolNamePatchFile,
			Description: openai.String("Unified Diff形式のパッチを使用してファイルを編集する"),
			Parameters:  getPatchFileParamsOnce(),
			Strict:      openai.Bool(true),
		},
	}
}

type PatchFileOut struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func PatchFile(_ context.Context, args PatchFileParamsJson) (*PatchFileOut, error) {
	// ファイルを読み込む
	originalContent, err := os.ReadFile(args.Path)
	if err != nil {
		return nil, fmt.Errorf("ファイル %q の読み込みに失敗しました: %w", args.Path, err)
	}

	// パッチをパース
	files, _, err := gitdiff.Parse(strings.NewReader(args.Patch))
	if err != nil {
		return nil, fmt.Errorf("パッチのパースに失敗しました: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("パッチが空です")
	}

	// 最初のファイルの差分を適用（単一ファイル編集を想定）
	file := files[0]

	// パッチを適用
	var output bytes.Buffer
	src := bytes.NewReader(originalContent)
	if err := gitdiff.Apply(&output, src, file); err != nil {
		return nil, fmt.Errorf("パッチの適用に失敗しました: %w", err)
	}

	// 結果をファイルに書き戻す
	if err := os.WriteFile(args.Path, output.Bytes(), 0644); err != nil {
		return nil, fmt.Errorf("ファイル %q の書き込みに失敗しました: %w", args.Path, err)
	}

	return &PatchFileOut{
		Success: true,
		Message: fmt.Sprintf("ファイル %q にパッチを適用しました", args.Path),
	}, nil
}
