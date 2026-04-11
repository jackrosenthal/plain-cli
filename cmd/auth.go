package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackrosenthal/plain-cli/internal/client"
	"github.com/jackrosenthal/plain-cli/internal/config"
	"github.com/jackrosenthal/plain-cli/internal/output"
)

type AuthCmd struct {
	Login  AuthLoginCmd  `cmd:"" help:"Authenticate with a Plain device."`
	Status AuthStatusCmd `cmd:"" help:"Check whether the current token is valid."`
}

type AuthLoginCmd struct {
	Password bool `help:"Prompt for a password before authenticating."`
}

type AuthStatusCmd struct{}

type authMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type authStatusResult struct {
	Status string `json:"status"`
	Valid  bool   `json:"valid"`
}

func (c *AuthLoginCmd) Run(cli *CLI, printer output.Printer) error {
	cfg, err := resolvedConfig(cli)
	if err != nil {
		return err
	}

	if cfg.Host == "" {
		cfg.Host, err = promptText("Plain URL:")
		if err != nil {
			return err
		}
		if err := requireHost(cfg.Host); err != nil {
			return err
		}
	}

	password := ""
	if c.Password {
		password, err = promptPassword()
		if err != nil {
			return err
		}
	} else {
		_, password, err = client.InitLogin(context.Background(), cfg.Host, cfg.ClientID, nil)
		if err != nil {
			return err
		}
	}

	token, err := client.Login(context.Background(), cfg.Host, cfg.ClientID, password, func() {
		_, _ = fmt.Fprintln(os.Stderr, "Waiting for device confirmation...")
	})
	if err != nil {
		return err
	}

	cfg.Token = token
	if err := cfg.Save(); err != nil {
		return err
	}

	return printer.Print(authMessage{
		Status:  "ok",
		Message: "Authentication succeeded.",
	})
}

func (c *AuthStatusCmd) Run(cli *CLI, printer output.Printer) error {
	cfg, err := resolvedConfig(cli)
	if err != nil {
		return err
	}
	if err := requireHost(cfg.Host); err != nil {
		return err
	}

	var sessionKey []byte
	if cfg.Token != "" {
		sessionKey, err = client.DeriveSessionKey(cfg.Token)
		if err != nil {
			return fmt.Errorf("derive session key: %w", err)
		}
	}

	valid, err := client.CheckToken(context.Background(), cfg.Host, cfg.ClientID, sessionKey)
	if err != nil {
		return err
	}

	status := "expired"
	if valid {
		status = "valid"
	}

	return printer.Print(authStatusResult{
		Status: status,
		Valid:  valid,
	})
}

func resolvedConfig(cli *CLI) (config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return config.Config{}, err
	}

	if cli == nil {
		return cfg, nil
	}

	if cli.Host != "" {
		cfg.Host = cli.Host
	}
	if cli.Token != "" {
		cfg.Token = cli.Token
	}
	if cli.ClientID != "" {
		cfg.ClientID = cli.ClientID
	}

	return cfg, nil
}

func requireHost(host string) error {
	if strings.TrimSpace(host) == "" {
		return errors.New("plain host is required; set --host, PLAIN_HOST, or config.host")
	}

	return nil
}

type passwordPromptModel struct {
	input    textinput.Model
	value    string
	canceled bool
}

func newPromptModel(prompt, placeholder string, echoMode textinput.EchoMode) passwordPromptModel {
	input := textinput.New()
	input.Placeholder = placeholder
	input.Prompt = prompt + " "
	input.EchoMode = echoMode
	if echoMode == textinput.EchoPassword {
		input.EchoCharacter = '*'
	}
	input.Focus()

	return passwordPromptModel{input: input}
}

func newPasswordPromptModel() passwordPromptModel {
	return newPromptModel("Password:", "Password", textinput.EchoPassword)
}

func newTextPromptModel(prompt string) passwordPromptModel {
	return newPromptModel(prompt, "", textinput.EchoNormal)
}

func (m passwordPromptModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m passwordPromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.canceled = true
			return m, tea.Quit
		case tea.KeyEnter:
			m.value = m.input.Value()
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m passwordPromptModel) View() string {
	return m.input.View() + "\n"
}

func promptPassword() (string, error) {
	program := tea.NewProgram(newPasswordPromptModel())
	model, err := program.Run()
	if err != nil {
		return "", err
	}

	result, ok := model.(passwordPromptModel)
	if !ok {
		return "", errors.New("unexpected password prompt result")
	}
	if result.canceled {
		return "", errors.New("password prompt canceled")
	}

	return result.value, nil
}

func promptText(prompt string) (string, error) {
	program := tea.NewProgram(newTextPromptModel(prompt))
	model, err := program.Run()
	if err != nil {
		return "", err
	}

	result, ok := model.(passwordPromptModel)
	if !ok {
		return "", errors.New("unexpected text prompt result")
	}
	if result.canceled {
		return "", errors.New("text prompt canceled")
	}

	return strings.TrimSpace(result.value), nil
}
