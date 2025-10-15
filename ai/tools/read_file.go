package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

//go:generate go tool go-jsonschema -p tools -o read_file_params_gen.go read_file_params.json

//go:embed read_file_params.json
var readFileParamsJSONSchema string

var getReadFileParamsOnce = sync.OnceValue(func() openai.FunctionParameters {
	var params openai.FunctionParameters
	_ = json.Unmarshal([]byte(readFileParamsJSONSchema), &params)
	return params
})

const ToolNameReadFile = "read_file"

func GetReadFileToolParam() responses.ToolUnionParam {
	return responses.ToolUnionParam{
		OfFunction: &responses.FunctionToolParam{
			Name:        ToolNameReadFile,
			Description: openai.String("指定されたファイルの内容を全て読み込む"),
			Parameters:  getReadFileParamsOnce(),
			Strict:      openai.Bool(true),
		},
	}
}

type ReadFileOut struct {
	Content string `json:"content"`
}

func ReadFile(_ context.Context, args ReadFileParamsJson) (*ReadFileOut, error) {
	file, err := os.Open(args.Path)
	if err != nil {
		return nil, fmt.Errorf("ファイル %q のオープンに失敗しました: %w", args.Path, err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("ファイル %q の読み込みに失敗しました: %w", args.Path, err)
	}

	return &ReadFileOut{
		Content: string(content),
	}, nil
}
