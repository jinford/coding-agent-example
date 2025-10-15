package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

//go:generate go tool go-jsonschema -p tools -o write_file_params_gen.go write_file_params.json

//go:embed write_file_params.json
var writeFileParamsJSONSchema string

var getWriteFileParamsOnce = sync.OnceValue(func() openai.FunctionParameters {
	var params openai.FunctionParameters
	_ = json.Unmarshal([]byte(writeFileParamsJSONSchema), &params)
	return params
})

const ToolNameWriteFile = "write_file"

func GetWriteFileToolParam() responses.ToolUnionParam {
	return responses.ToolUnionParam{
		OfFunction: &responses.FunctionToolParam{
			Name:        ToolNameWriteFile,
			Description: openai.String("指定されたパスに新しいファイルを作成し、内容を書き込む"),
			Parameters:  getWriteFileParamsOnce(),
			Strict:      openai.Bool(true),
		},
	}
}

type WriteFileOut struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func WriteFile(_ context.Context, args WriteFileParamsJson) (*WriteFileOut, error) {
	// ディレクトリが存在しない場合は作成
	dir := filepath.Dir(args.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("ディレクトリ %q の作成に失敗しました: %w", dir, err)
	}

	// ファイルを作成して内容を書き込む
	if err := os.WriteFile(args.Path, []byte(args.Content), 0644); err != nil {
		return nil, fmt.Errorf("ファイル %q の書き込みに失敗しました: %w", args.Path, err)
	}

	return &WriteFileOut{
		Success: true,
		Message: fmt.Sprintf("ファイル %q を正常に作成しました", args.Path),
	}, nil
}
