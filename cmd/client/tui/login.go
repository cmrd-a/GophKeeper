package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// LoginScreen represents the login/registration screen
type LoginScreen struct {
	form          *huh.Form
	isRegistering bool
	login         string
	password      string
	confirmPass   string
	width         int
	height        int
}

// NewLoginScreen creates a new login screen
func NewLoginScreen() *LoginScreen {
	ls := &LoginScreen{}
	ls.buildForm()
	return ls
}

// Init initializes the login screen
func (ls *LoginScreen) Init() tea.Cmd {
	return ls.form.Init()
}

// Update handles messages for the login screen
func (ls *LoginScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ls.width = msg.Width
		ls.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+r":
			// Toggle between login and register mode
			ls.isRegistering = !ls.isRegistering
			ls.buildForm()
			return ls, ls.form.Init()
		case "enter":
			if ls.form.State == huh.StateCompleted {
				if ls.isRegistering {
					// Validate password confirmation
					if ls.password != ls.confirmPass {
						return ls, func() tea.Msg {
							return RegisterErrorMsg{Error: "Passwords do not match"}
						}
					}
					return ls, ls.performRegister()
				}
				return ls, ls.performLogin()
			}
		}
	}

	// Update form
	form, cmd := ls.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		ls.form = f
	}

	return ls, cmd
}

// View renders the login screen
func (ls *LoginScreen) View() string {
	var title string
	if ls.isRegistering {
		title = "Register New Account"
	} else {
		title = "Login"
	}

	titleView := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		MarginBottom(2).
		Render(title)

	formView := ls.form.View()

	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272A4")).
		MarginTop(2).
		Render("Ctrl+R: Switch to " + ls.getToggleText() + " • Enter: Submit")

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleView,
		formView,
		instructions,
	)

	// Center the content
	if ls.width > 0 && ls.height > 0 {
		contentWidth := lipgloss.Width(content)
		contentHeight := lipgloss.Height(content)

		horizontalPadding := (ls.width - contentWidth) / 2
		verticalPadding := (ls.height - contentHeight) / 2

		if horizontalPadding > 0 {
			content = lipgloss.NewStyle().
				PaddingLeft(horizontalPadding).
				Render(content)
		}

		if verticalPadding > 0 {
			content = strings.Repeat("\n", verticalPadding) + content
		}
	}

	return content
}

// buildForm constructs the form based on current mode
func (ls *LoginScreen) buildForm() {
	var inputs []huh.Field

	// Login field
	inputs = append(inputs, huh.NewInput().
		Title("Username").
		Value(&ls.login).
		Placeholder("Enter your username").
		Validate(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return fmt.Errorf("username cannot be empty")
			}
			if len(s) < 3 {
				return fmt.Errorf("username must be at least 3 characters")
			}
			return nil
		}))

	// Password field
	inputs = append(inputs, huh.NewInput().
		Title("Password").
		Value(&ls.password).
		Placeholder("Enter your password").
		EchoMode(huh.EchoModePassword).
		Validate(func(s string) error {
			if len(s) < 6 {
				return fmt.Errorf("password must be at least 6 characters")
			}
			return nil
		}))

	// Confirm password field (only for registration)
	if ls.isRegistering {
		inputs = append(inputs, huh.NewInput().
			Title("Confirm Password").
			Value(&ls.confirmPass).
			Placeholder("Confirm your password").
			EchoMode(huh.EchoModePassword).
			Validate(func(s string) error {
				if s != ls.password {
					return fmt.Errorf("passwords do not match")
				}
				return nil
			}))
	}

	ls.form = huh.NewForm(
		huh.NewGroup(inputs...),
	).WithWidth(50).WithHeight(10)
}

// getToggleText returns the text for mode toggle
func (ls *LoginScreen) getToggleText() string {
	if ls.isRegistering {
		return "Login"
	}
	return "Register"
}

// performLogin performs the login operation
func (ls *LoginScreen) performLogin() tea.Cmd {
	login := ls.login
	password := ls.password

	return func() tea.Msg {
		// Don't create context here - let the parent App handle it
		// This prevents premature cancellation
		return LoginAttemptMsg{
			Login:    login,
			Password: password,
		}
	}
}

// performRegister performs the registration operation
func (ls *LoginScreen) performRegister() tea.Cmd {
	login := ls.login
	password := ls.password

	return func() tea.Msg {
		// Don't create context here - let the parent App handle it
		// This prevents premature cancellation
		return RegisterAttemptMsg{
			Login:    login,
			Password: password,
		}
	}
}

// Messages for login operations
type LoginAttemptMsg struct {
	Login    string
	Password string
}

type RegisterAttemptMsg struct {
	Login    string
	Password string
}
