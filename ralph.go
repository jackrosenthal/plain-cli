//go:build ignore

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/lipgloss"
)

var cli struct {
	Agent string `arg:"" enum:"claude,codex" help:"Agent to use (claude or codex)."`
}

var (
	stepCounterStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("12"))

	stepTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	doneStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10"))

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("9"))
)

func scanSteps(path string) (completed, total int, nextStep string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "- [x] ") {
			completed++
			total++
		} else if strings.HasPrefix(line, "- [ ] ") {
			if nextStep == "" {
				nextStep = strings.TrimPrefix(line, "- [ ] ")
			}
			total++
		}
	}
	return completed, total, nextStep, scanner.Err()
}

func main() {
	kong.Parse(&cli)

	prompt, err := os.ReadFile("prompt.md")
	if err != nil {
		fmt.Fprintln(os.Stderr, errorStyle.Render("error reading prompt.md: "+err.Error()))
		os.Exit(1)
	}

	home := os.Getenv("HOME")
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, errorStyle.Render("error getting working directory: "+err.Error()))
		os.Exit(1)
	}

	for {
		completed, total, step, err := scanSteps("todo.md")
		if err != nil {
			fmt.Fprintln(os.Stderr, errorStyle.Render("error reading todo.md: "+err.Error()))
			os.Exit(1)
		}
		if step == "" {
			fmt.Println(doneStyle.Render("✓ All steps completed."))
			return
		}

		current := completed + 1
		divider := dividerStyle.Render(strings.Repeat("─", 60))
		fmt.Println(divider)
		fmt.Println(stepCounterStyle.Render(fmt.Sprintf("Step %d / %d", current, total)))
		fmt.Println(stepTitleStyle.Render(step))
		fmt.Println(divider)
		fmt.Println()

		var agentCmd []string
		switch cli.Agent {
		case "claude":
			agentCmd = []string{"claude", "-p", "--dangerously-skip-permissions", string(prompt)}
		case "codex":
			agentCmd = []string{"codex", "exec", "--dangerously-bypass-approvals-and-sandbox", string(prompt)}
		}

		podmanArgs := []string{
			"run", "--rm",
			"-v", cwd + ":/workspace",
			"-v", home + "/.claude:/root/.claude",
			"-v", home + "/.codex:/root/.codex",
			"plain-cli-agent",
		}
		podmanArgs = append(podmanArgs, agentCmd...)

		cmd := exec.Command("podman", podmanArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, errorStyle.Render("\nagent exited non-zero: "+err.Error()))
			os.Exit(1)
		}

		completedAfter, totalAfter, nextStep, err := scanSteps("todo.md")
		if err != nil {
			fmt.Fprintln(os.Stderr, errorStyle.Render("error reading todo.md after run: "+err.Error()))
			os.Exit(1)
		}
		if totalAfter != total {
			fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("unexpected: todo.md step count changed (%d → %d)", total, totalAfter)))
			os.Exit(1)
		}
		if nextStep == step {
			fmt.Fprintln(os.Stderr, errorStyle.Render("expected step was not marked complete: "+step))
			os.Exit(1)
		}
		if completedAfter != completed+1 {
			fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("expected exactly 1 step to be completed, got %d", completedAfter-completed)))
			os.Exit(1)
		}
	}
}
