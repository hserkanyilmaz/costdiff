package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/hserkanyilmaz/costdiff/internal/diff"
)

// RenderJSON outputs the diff result as JSON to stdout
func RenderJSON(result *diff.Result) error {
	return RenderJSONTo(os.Stdout, result)
}

// RenderJSONTo outputs the diff result as JSON to the specified writer
func RenderJSONTo(w io.Writer, result *diff.Result) error {
	output := result.ToJSON()
	return writeJSON(w, output)
}

// RenderTopJSON outputs the top result as JSON to stdout
func RenderTopJSON(result *diff.TopResult) error {
	return RenderTopJSONTo(os.Stdout, result)
}

// RenderTopJSONTo outputs the top result as JSON to the specified writer
func RenderTopJSONTo(w io.Writer, result *diff.TopResult) error {
	output := result.ToJSON()
	return writeJSON(w, output)
}

// RenderWatchJSON outputs the watch result as JSON to stdout
func RenderWatchJSON(result *diff.WatchResult) error {
	return RenderWatchJSONTo(os.Stdout, result)
}

// RenderWatchJSONTo outputs the watch result as JSON to the specified writer
func RenderWatchJSONTo(w io.Writer, result *diff.WatchResult) error {
	output := result.ToJSON()
	return writeJSON(w, output)
}

func writeJSON(w io.Writer, v interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(v); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

