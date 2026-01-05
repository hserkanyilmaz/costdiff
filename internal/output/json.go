package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hsy/costdiff/internal/diff"
)

// RenderJSON outputs the diff result as JSON
func RenderJSON(result *diff.Result) error {
	output := result.ToJSON()
	return writeJSON(output)
}

// RenderTopJSON outputs the top result as JSON
func RenderTopJSON(result *diff.TopResult) error {
	output := result.ToJSON()
	return writeJSON(output)
}

// RenderWatchJSON outputs the watch result as JSON
func RenderWatchJSON(result *diff.WatchResult) error {
	output := result.ToJSON()
	return writeJSON(output)
}

func writeJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(v); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

