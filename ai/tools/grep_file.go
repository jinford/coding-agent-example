package tools

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

//go:generate go tool go-jsonschema -p tools -o grep_file_params_gen.go grep_file_params.json

//go:embed grep_file_params.json
var grepFileParamsJSONSchema string

var getGrepFileParamsOnce = sync.OnceValue(func() openai.FunctionParameters {
	var params openai.FunctionParameters
	_ = json.Unmarshal([]byte(grepFileParamsJSONSchema), &params)
	return params
})

const ToolNameGrepFile = "grep_file"

func GetGrepFileToolParam() responses.ToolUnionParam {
	return responses.ToolUnionParam{
		OfFunction: &responses.FunctionToolParam{
			Name:        ToolNameGrepFile,
			Description: openai.String("指定されたディレクトリ配下を再帰的に検索し、キーワードを含むファイルを見つける"),
			Parameters:  getGrepFileParamsOnce(),
			Strict:      openai.Bool(true),
		},
	}
}

type GrepMatch struct {
	FilePath   string `json:"file_path"`
	LineNumber int    `json:"line_number"`
	Line       string `json:"line"`
}

type GrepFileOut struct {
	Matches []GrepMatch `json:"matches"`
}

func GrepFile(_ context.Context, args GrepFileParamsJson) (*GrepFileOut, error) {
	var matches []GrepMatch
	keyword := args.Keyword

	// case_sensitiveがfalseの場合は小文字に変換
	caseSensitive := args.CaseSensitive
	if !caseSensitive {
		keyword = strings.ToLower(keyword)
	}

	err := filepath.Walk(args.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// ディレクトリはスキップ
		if info.IsDir() {
			return nil
		}

		// ファイルを開いて行ごとに検索
		file, err := os.Open(path)
		if err != nil {
			// 読み込めないファイルはスキップ
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNumber := 0
		for scanner.Scan() {
			lineNumber++
			line := scanner.Text()
			searchLine := line

			if !caseSensitive {
				searchLine = strings.ToLower(searchLine)
			}

			if strings.Contains(searchLine, keyword) {
				matches = append(matches, GrepMatch{
					FilePath:   path,
					LineNumber: lineNumber,
					Line:       line,
				})
			}
		}

		if err := scanner.Err(); err != nil {
			// スキャンエラーは無視して続行
			return nil
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("ディレクトリ %q の検索に失敗しました: %w", args.Path, err)
	}

	return &GrepFileOut{
		Matches: matches,
	}, nil
}
