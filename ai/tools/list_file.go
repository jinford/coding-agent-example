package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

//go:generate go tool go-jsonschema -p tools -o list_file_params_gen.go list_file_params.json

//go:embed list_file_params.json
var listFileParamsJSONSchema string

var getListFileParamsOnce = sync.OnceValue(func() openai.FunctionParameters {
	var params openai.FunctionParameters
	_ = json.Unmarshal([]byte(listFileParamsJSONSchema), &params)
	return params
})

const ToolNameListFile = "list_file"

func GetListFileToolParam() responses.ToolUnionParam {
	return responses.ToolUnionParam{
		OfFunction: &responses.FunctionToolParam{
			Name:        ToolNameListFile,
			Description: openai.String("指定されたディレクトリ内のファイルとディレクトリの一覧を取得する"),
			Parameters:  getListFileParamsOnce(),
			Strict:      openai.Bool(true),
		},
	}
}

type FileEntry struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
}

type ListFileOut struct {
	Entries []FileEntry `json:"entries"`
}

func ListFile(_ context.Context, args ListFileParamsJson) (*ListFileOut, error) {
	entries, err := os.ReadDir(args.Path)
	if err != nil {
		return nil, fmt.Errorf("ディレクトリ %q の読み込みに失敗しました: %w", args.Path, err)
	}

	result := make([]FileEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, FileEntry{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
		})
	}

	return &ListFileOut{
		Entries: result,
	}, nil
}
