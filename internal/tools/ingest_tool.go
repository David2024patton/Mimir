package tools

import (
	"context"
	"fmt"

	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/ingest"
)

// IngestRepoTool exposes Mímir's built-in GitIngest & Yggdrasil Mind Map builder
// to the agent so it can ingest any local directory or GitHub URL directly into Cortex memory.
type IngestRepoTool struct {
	WorkDir string
	Store   cortex.Store
}

func (t *IngestRepoTool) Name() string { return "ingest_repo" }

func (t *IngestRepoTool) Description() string {
	return "Ingests a local directory or Git repository URL (like gitingest.com), builds a Yggdrasil Mind Map of files/symbols/tests, and stores it in Cortex memory."
}

func (t *IngestRepoTool) Parameters() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"source": map[string]any{
				"type":        "string",
				"description": "Local folder path or remote Git repository URL (defaults to current working directory).",
			},
			"branch": map[string]any{
				"type":        "string",
				"description": "Optional Git branch to clone when source is a remote URL.",
			},
		},
	}
}

func (t *IngestRepoTool) Run(ctx context.Context, args map[string]any) (string, error) {
	src, _ := args["source"].(string)
	if src == "" {
		src = t.WorkDir
		if src == "" {
			src = "."
		}
	}
	branch, _ := args["branch"].(string)

	res, err := ingest.Run(ctx, ingest.Options{
		Source: src,
		Branch: branch,
	}, t.Store)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s\nSaved %d Yggdrasil Mind Map neurons to Cortex.\n\n%s",
		res.Summary, res.NeuronsSaved, ingest.FormatMindMapCompact(res.MindMap, 30)), nil
}
