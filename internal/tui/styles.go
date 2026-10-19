// Package tui implements Mímir's interactive terminal interface (inspired by
// Charmbracelet Bubble Tea / Lipgloss / Glamour / Huh and smolcoder).
package tui

import (
	"fmt"
	"strings"
)

// ANSI 256 / TrueColor Lipgloss-inspired palette for Mímir's Norse/Book-Spine aesthetic.
const (
	ansiReset   = "\033[0m"
	ansiBold    = "\033[1m"
	ansiDim     = "\033[2m"
	ansiItalic  = "\033[3m"
	ansiGold    = "\033[38;5;214m" // Mímir gold accent
	ansiCyan    = "\033[38;5;81m"  // Tool & code accent
	ansiGreen   = "\033[38;5;114m" // Success / diff +
	ansiRed     = "\033[38;5;203m" // Error / diff -
	ansiPurple  = "\033[38;5;141m" // Memory / Cortex neuron
	ansiGray    = "\033[38;5;244m" // Subtle border / status
	ansiBgBadge = "\033[48;5;236m" // Pill background
)

// RenderBanner renders the compact Charmbracelet/smolcoder-style startup card.
func RenderBanner(version, providerID, model, mode, effort, storeDesc, webURL string, toolCount int) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(ansiGold + ansiBold + "  ╭──────────────────────────────────────────────────────────────────╮\n" + ansiReset)
	title := fmt.Sprintf("Mímir %s", version)
	sub := "the agent that remembers · tiny self-hosted CLI & TUI"
	b.WriteString(fmt.Sprintf("  %s│%s  %s%-16s%s %s%-46s%s %s│%s\n",
		ansiGold, ansiReset, ansiBold, title, ansiReset, ansiDim, sub, ansiReset, ansiGold, ansiReset))
	b.WriteString(ansiGold + "  ├──────────────────────────────────────────────────────────────────┤\n" + ansiReset)

	provLine := fmt.Sprintf("%s · %s", providerID, model)
	if providerID == "" || providerID == "none" {
		provLine = "none (run Ollama on :11434 or set MIMIR_BASE_URL / MIMO_API_KEY)"
	}
	b.WriteString(fmt.Sprintf("  %s│%s  %sModel:%s   %-55s%s│%s\n",
		ansiGold, ansiReset, ansiCyan, ansiReset, truncate(provLine, 55), ansiGold, ansiReset))

	effStr := effort
	if effStr == "" {
		effStr = "default"
	}
	modeLine := fmt.Sprintf("mode=%s  ·  effort=%s  ·  tools=%d", mode, effStr, toolCount)
	b.WriteString(fmt.Sprintf("  %s│%s  %sState:%s   %-55s%s│%s\n",
		ansiGold, ansiReset, ansiGreen, ansiReset, truncate(modeLine, 55), ansiGold, ansiReset))

	b.WriteString(fmt.Sprintf("  %s│%s  %sCortex:%s  %-55s%s│%s\n",
		ansiGold, ansiReset, ansiPurple, ansiReset, truncate(storeDesc, 55), ansiGold, ansiReset))

	if webURL != "" {
		b.WriteString(fmt.Sprintf("  %s│%s  %sWeb UI:%s  %-55s%s│%s\n",
			ansiGold, ansiReset, ansiGold, ansiReset, truncate(webURL, 55), ansiGold, ansiReset))
	}

	b.WriteString(ansiGold + "  ╰──────────────────────────────────────────────────────────────────╯\n" + ansiReset)
	b.WriteString(ansiDim + "  Commands: /help · /model · /mode [ro|edit|bypass] · /effort [off|low|med|high] · /todo · /web · /quit\n\n" + ansiReset)
	return b.String()
}

// RenderStatusBar renders the 1-line status bar shown before each prompt.
func RenderStatusBar(providerID, model, mode, effort string, estTokens, totalTokens int, local, smol bool) string {
	modeBadge := ansiGreen + "[" + mode + "]" + ansiReset
	switch mode {
	case "ro":
		modeBadge = ansiCyan + "[ro (read-only)]" + ansiReset
	case "bypass":
		modeBadge = ansiGold + "[bypass]" + ansiReset
	}
	smolBadge := ""
	if smol {
		smolBadge = " " + ansiGold + "⚡smol" + ansiReset
	}
	eff := effort
	if eff == "" {
		eff = "auto"
	}
	ctxPct := (estTokens * 100) / 32768
	if ctxPct > 100 {
		ctxPct = 100
	}
	costLabel := "cloud"
	if local {
		costLabel = "$0 local"
	}
	return fmt.Sprintf("%s●%s %s%s/%s%s%s %s·%s effort:%s %s·%s ctx:%d%% (%d tok, %s) %s·%s %s",
		ansiGold, ansiReset,
		ansiBold, providerID, model, ansiReset, smolBadge,
		ansiGray, ansiReset, eff,
		ansiGray, ansiReset, ctxPct, totalTokens, costLabel,
		ansiGray, ansiReset, modeBadge)
}

// RenderToolPill formats a compact collapsible-style tool execution pill with diff highlighting.
func RenderToolPill(name, args, result, errStr string) string {
	var b strings.Builder
	icon := ansiCyan + "▸" + ansiReset
	status := ansiGreen + "✓" + ansiReset
	if errStr != "" {
		status = ansiRed + "✗ " + truncate(errStr, 70) + ansiReset
	}
	summary := summarizeToolArgs(name, args)
	b.WriteString(fmt.Sprintf("  %s %s%s%s %s%s%s %s\n",
		icon, ansiBgBadge+ansiBold, " "+name+" ", ansiReset, ansiDim, summary, ansiReset, status))

	// Show a compact 4-line preview for shell/read/edit outputs
	if strings.TrimSpace(result) != "" && errStr == "" {
		lines := strings.Split(strings.TrimSpace(result), "\n")
		maxLines := 4
		if len(lines) < maxLines {
			maxLines = len(lines)
		}
		for i := 0; i < maxLines; i++ {
			line := lines[i]
			color := ansiDim
			if strings.HasPrefix(line, "+") {
				color = ansiGreen
			} else if strings.HasPrefix(line, "-") {
				color = ansiRed
			}
			b.WriteString(fmt.Sprintf("    %s│ %s%s\n", ansiGray, color+truncate(line, 88), ansiReset))
		}
		if len(lines) > maxLines {
			b.WriteString(fmt.Sprintf("    %s╰─ ... (%d more lines)%s\n", ansiDim, len(lines)-maxLines, ansiReset))
		}
	}
	return b.String()
}

// FormatMarkdown applies lightweight Glamour-style terminal highlighting to Markdown blocks
// and converts [choice]...[/choice] blocks into numbered terminal option cards.
func FormatMarkdown(text string) string {
	lines := strings.Split(text, "\n")
	var out strings.Builder
	inCode := false
	inChoice := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[choice]" {
			inChoice = true
			out.WriteString(ansiGold + "  ╭─ Options ─────────────────────────────────────────╮\n" + ansiReset)
			continue
		}
		if trimmed == "[/choice]" {
			inChoice = false
			out.WriteString(ansiGold + "  ╰───────────────────────────────────────────────────╯\n" + ansiReset)
			continue
		}
		if inChoice {
			out.WriteString(fmt.Sprintf("  %s│%s %s%s%s\n", ansiGold, ansiReset, ansiCyan+ansiBold, trimmed, ansiReset))
			continue
		}
		if strings.HasPrefix(trimmed, "```") {
			inCode = !inCode
			out.WriteString(ansiGray + "  " + line + ansiReset + "\n")
			continue
		}
		if inCode {
			if strings.HasPrefix(line, "+") {
				out.WriteString("  " + ansiGreen + line + ansiReset + "\n")
			} else if strings.HasPrefix(line, "-") {
				out.WriteString("  " + ansiRed + line + ansiReset + "\n")
			} else {
				out.WriteString("  " + ansiCyan + line + ansiReset + "\n")
			}
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			out.WriteString(ansiGold + ansiBold + line + ansiReset + "\n")
			continue
		}
		out.WriteString(line + "\n")
	}
	return strings.TrimRight(out.String(), "\n")
}

func summarizeToolArgs(name, rawArgs string) string {
	rawArgs = strings.TrimSpace(rawArgs)
	if len(rawArgs) > 72 {
		return rawArgs[:69] + "..."
	}
	return rawArgs
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
