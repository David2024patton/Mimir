package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/David2024patton/Mimir/internal/tools"
)

// PromptQuestions renders a Charmbracelet Huh-inspired interactive terminal form
// for Mímir's QuestionTool. Users can pick an option by number/letter or type a
// custom write-in response directly.
func PromptQuestions(ctx context.Context, in io.Reader, out io.Writer, questions []tools.Question) (string, error) {
	reader := bufio.NewReader(in)
	var answers []string

	for idx, q := range questions {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		header := q.Header
		if header == "" {
			header = fmt.Sprintf("Question %d", idx+1)
		}
		fmt.Fprintf(out, "\n%s╭─ %s ─%s\n", ansiGold+ansiBold, header, ansiReset)
		fmt.Fprintf(out, "%s│%s %s%s%s\n", ansiGold, ansiReset, ansiBold, q.Question, ansiReset)
		if q.Multiple {
			fmt.Fprintf(out, "%s│%s %s(Select one or more numbers separated by commas, or type a custom answer)%s\n",
				ansiGold, ansiReset, ansiDim, ansiReset)
		} else {
			fmt.Fprintf(out, "%s│%s %s(Enter option number 1-%d, or type a custom write-in answer)%s\n",
				ansiGold, ansiReset, ansiDim, len(q.Options), ansiReset)
		}

		for i, opt := range q.Options {
			desc := ""
			if opt.Description != "" {
				desc = " - " + ansiDim + opt.Description + ansiReset
			}
			fmt.Fprintf(out, "%s│%s   %s[%d]%s %s%s%s%s\n",
				ansiGold, ansiReset, ansiCyan+ansiBold, i+1, ansiReset, ansiBold, opt.Label, ansiReset, desc)
		}
		fmt.Fprintf(out, "%s╰─▸%s ", ansiGold+ansiBold, ansiReset)

		line, err := reader.ReadString('\n')
		if err != nil && len(line) == 0 {
			return "", err
		}
		choice := strings.TrimSpace(line)
		resolved := resolveChoice(choice, q.Options, q.Multiple)
		answers = append(answers, fmt.Sprintf("%s: %s", header, resolved))
	}

	return strings.Join(answers, "\n"), nil
}

func resolveChoice(input string, options []tools.Option, multiple bool) string {
	if input == "" {
		if len(options) > 0 {
			return options[0].Label
		}
		return "(skipped)"
	}
	if multiple && strings.Contains(input, ",") {
		parts := strings.Split(input, ",")
		var picked []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if n, err := strconv.Atoi(p); err == nil && n >= 1 && n <= len(options) {
				picked = append(picked, options[n-1].Label)
			} else if p != "" {
				picked = append(picked, p)
			}
		}
		return strings.Join(picked, ", ")
	}
	if n, err := strconv.Atoi(input); err == nil && n >= 1 && n <= len(options) {
		return options[n-1].Label
	}
	return input
}
