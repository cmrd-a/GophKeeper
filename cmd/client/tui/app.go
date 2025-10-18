package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cmrd-a/GophKeeper/gen/proto/v1/vault"
)

// Client interface for the gRPC client
type Client interface {
	Login(ctx context.Context, login, password string) error
	Register(ctx context.Context, login, password string) error
	GetVaultItems(ctx context.Context) (*vault.GetVaultItemsResponse, error)
	SaveLoginPassword(ctx context.Context, login, password string) (string, error)
	SaveTextData(ctx context.Context, text string) (string, error)
	SaveCardData(ctx context.Context, number, holder, expire, cvv string) (string, error)
	SaveBinaryData(ctx context.Context, data []byte) (string, error)
	SaveMeta(ctx context.Context, meta []*vault.Meta) error
	DeleteVaultItem(ctx context.Context, id, itemType string) error
}

// AppState represents the current state of the application
type AppState int

const (
	StateLogin AppState = iota
	StateMain
	StateAddItem
	StateViewItem
)

// ItemType represents the type of vault item
type ItemType int

const (
	TypeLoginPassword ItemType = iota
	TypeTextData
	TypeCardData
	TypeBinaryData
)

// App represents the main TUI application
type App struct {
	client Client
	state  AppState
	width  int
	height int

	// Authentication
	isAuthenticated bool

	// Current screens
	loginScreen    *LoginScreen
	mainScreen     *MainScreen
	addItemScreen  *AddItemScreen
	viewItemScreen *ViewItemScreen

	// Status and messages
	message     string
	messageType MessageType

	// Loading state
	loading bool
}

// MessageType represents the type of message to display
type MessageType int

const (
	MessageInfo MessageType = iota
	MessageError
	MessageSuccess
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			MarginTop(1)
)

// NewApp creates a new TUI application
func NewApp(client Client) *App {
	app := &App{
		client: client,
		state:  StateLogin,
	}

	// Initialize screens
	app.loginScreen = NewLoginScreen()
	app.mainScreen = NewMainScreen()
	app.addItemScreen = NewAddItemScreen()
	app.viewItemScreen = NewViewItemScreen()

	return app
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.loginScreen.Init(),
		tea.EnterAltScreen,
	)
}

// Update handles messages and updates the application state
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

		// Update all screens with new dimensions
		_, cmd := a.loginScreen.Update(msg)
		cmds = append(cmds, cmd)
		_, cmd = a.mainScreen.Update(msg)
		cmds = append(cmds, cmd)
		_, cmd = a.addItemScreen.Update(msg)
		cmds = append(cmds, cmd)
		_, cmd = a.viewItemScreen.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if a.state == StateMain {
				return a, tea.Quit
			}
		case "esc":
			// Go back to previous state
			switch a.state {
			case StateAddItem, StateViewItem:
				a.state = StateMain
				a.clearMessage()
			}
		}

	case LoginSuccessMsg:
		a.isAuthenticated = true
		a.state = StateMain
		a.setMessage("Login successful!", MessageSuccess)
		return a, a.loadVaultItems()

	case LoginErrorMsg:
		a.setMessage(fmt.Sprintf("Login failed: %s", msg.Error), MessageError)

	case RegisterSuccessMsg:
		a.setMessage("Registration successful! Please log in.", MessageSuccess)

	case RegisterErrorMsg:
		a.setMessage(fmt.Sprintf("Registration failed: %s", msg.Error), MessageError)

	case VaultItemsLoadedMsg:
		a.mainScreen.SetVaultItems(msg.Items)
		a.loading = false

	case VaultItemsErrorMsg:
		a.setMessage(fmt.Sprintf("Failed to load items: %s", msg.Error), MessageError)
		a.loading = false

	case AddItemMsg:
		a.state = StateAddItem
		a.addItemScreen.SetItemType(msg.Type)
		a.clearMessage()

	case ItemSavedMsg:
		a.state = StateMain
		a.setMessage("Item saved successfully!", MessageSuccess)
		return a, a.loadVaultItems()

	case ItemSaveErrorMsg:
		a.setMessage(fmt.Sprintf("Failed to save item: %s", msg.Error), MessageError)

	case ViewItemMsg:
		a.state = StateViewItem
		a.viewItemScreen.SetItem(msg.Item, msg.Type)
		a.clearMessage()

	case DeleteItemMsg:
		a.state = StateMain
		a.setMessage("Item deleted successfully!", MessageSuccess)
		return a, a.loadVaultItems()

	case DeleteItemErrorMsg:
		a.setMessage(fmt.Sprintf("Failed to delete item: %s", msg.Error), MessageError)

	case LoadingMsg:
		a.loading = msg.Loading

	case LoginAttemptMsg:
		return a, a.performLogin(msg.Login, msg.Password)

	case RegisterAttemptMsg:
		return a, a.performRegister(msg.Login, msg.Password)

	case ShowAddMenuMsg:
		a.state = StateAddItem
		a.addItemScreen.Reset()
		a.clearMessage()

	case SaveItemAttemptMsg:
		return a, a.performSaveItem(msg)

	case DeleteItemAttemptMsg:
		return a, a.performDeleteItem(msg.ID, msg.Type)

	case RefreshItemsMsg:
		return a, a.loadVaultItems()

	case CopyToClipboardMsg:
		a.setMessage("Copied to clipboard!", MessageSuccess)
	}

	// Update the current screen
	switch a.state {
	case StateLogin:
		_, cmd := a.loginScreen.Update(msg)
		cmds = append(cmds, cmd)

	case StateMain:
		_, cmd := a.mainScreen.Update(msg)
		cmds = append(cmds, cmd)

	case StateAddItem:
		_, cmd := a.addItemScreen.Update(msg)
		cmds = append(cmds, cmd)

	case StateViewItem:
		_, cmd := a.viewItemScreen.Update(msg)
		cmds = append(cmds, cmd)
	}

	return a, tea.Batch(cmds...)
}

// View renders the application
func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	var content string

	// Header
	header := titleStyle.Render("🔐 GophKeeper Client")

	// Main content based on state
	switch a.state {
	case StateLogin:
		content = a.loginScreen.View()

	case StateMain:
		content = a.mainScreen.View()

	case StateAddItem:
		content = a.addItemScreen.View()

	case StateViewItem:
		content = a.viewItemScreen.View()
	}

	// Footer with status message
	var footer strings.Builder
	if a.message != "" {
		var style lipgloss.Style
		switch a.messageType {
		case MessageError:
			style = errorStyle
		case MessageSuccess:
			style = successStyle
		case MessageInfo:
			style = infoStyle
		}
		footer.WriteString(style.Render(a.message))
		footer.WriteString("\n")
	}

	// Loading indicator
	if a.loading {
		footer.WriteString(infoStyle.Render("Loading..."))
		footer.WriteString("\n")
	}

	// Help text
	help := a.getHelpText()
	if help != "" {
		footer.WriteString(helpStyle.Render(help))
	}

	// Combine all parts
	body := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer.String(),
	)

	// Center the content if needed
	if a.width > 0 && a.height > 0 {
		bodyHeight := lipgloss.Height(body)
		if bodyHeight < a.height {
			padding := (a.height - bodyHeight) / 2
			if padding > 0 {
				body = strings.Repeat("\n", padding) + body
			}
		}
	}

	return body
}

// setMessage sets a status message
func (a *App) setMessage(message string, messageType MessageType) {
	a.message = message
	a.messageType = messageType

	// Clear message after 3 seconds
	go func() {
		time.Sleep(3 * time.Second)
		a.clearMessage()
	}()
}

// clearMessage clears the current status message
func (a *App) clearMessage() {
	a.message = ""
}

// getHelpText returns context-appropriate help text
func (a *App) getHelpText() string {
	switch a.state {
	case StateLogin:
		return "Tab/Shift+Tab: Navigate • Enter: Submit • Ctrl+C: Quit"
	case StateMain:
		return "↑/↓: Navigate • Enter: View item • a: Add item • d: Delete item • q: Quit"
	case StateAddItem:
		return "Tab/Shift+Tab: Navigate • Enter: Save • Esc: Cancel"
	case StateViewItem:
		return "Esc: Back • d: Delete item"
	default:
		return "Esc: Back • Ctrl+C: Quit"
	}
}

// loadVaultItems loads vault items from the server
func (a *App) loadVaultItems() tea.Cmd {
	return func() tea.Msg {
		a.loading = true

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		items, err := a.client.GetVaultItems(ctx)
		if err != nil {
			return VaultItemsErrorMsg{Error: err.Error()}
		}

		return VaultItemsLoadedMsg{Items: items}
	}
}

// performLogin handles login attempts
func (a *App) performLogin(login, password string) tea.Cmd {
	return func() tea.Msg {
		// Create fresh context for each login attempt to avoid cancellation
		ctx := context.Background()
		err := a.client.Login(ctx, login, password)
		if err != nil {
			return LoginErrorMsg{Error: err.Error()}
		}
		return LoginSuccessMsg{}
	}
}

// performRegister handles registration attempts
func (a *App) performRegister(login, password string) tea.Cmd {
	return func() tea.Msg {
		// Create fresh context for each registration attempt to avoid cancellation
		ctx := context.Background()
		err := a.client.Register(ctx, login, password)
		if err != nil {
			return RegisterErrorMsg{Error: err.Error()}
		}
		return RegisterSuccessMsg{}
	}
}

// performSaveItem handles saving vault items
func (a *App) performSaveItem(msg SaveItemAttemptMsg) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		switch msg.Type {
		case TypeLoginPassword:
			data := msg.Data
			login := data["login"].(string)
			password := data["password"].(string)
			_, err = a.client.SaveLoginPassword(ctx, login, password)

		case TypeTextData:
			data := msg.Data
			text := data["text"].(string)
			_, err = a.client.SaveTextData(ctx, text)

		case TypeCardData:
			data := msg.Data
			number := data["number"].(string)
			holder := data["holder"].(string)
			expire := data["expire"].(string)
			cvv := data["cvv"].(string)
			_, err = a.client.SaveCardData(ctx, number, holder, expire, cvv)

		case TypeBinaryData:
			data := msg.Data
			binaryData := data["data"].([]byte)
			_, err = a.client.SaveBinaryData(ctx, binaryData)
		}

		if err != nil {
			return ItemSaveErrorMsg{Error: err.Error()}
		}
		return ItemSavedMsg{}
	}
}

// performDeleteItem handles deleting vault items
func (a *App) performDeleteItem(id, itemType string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := a.client.DeleteVaultItem(ctx, id, itemType)
		if err != nil {
			return DeleteItemErrorMsg{Error: err.Error()}
		}
		return DeleteItemMsg{}
	}
}

// Messages for communication between components

type LoginSuccessMsg struct{}
type LoginErrorMsg struct{ Error string }
type RegisterSuccessMsg struct{}
type RegisterErrorMsg struct{ Error string }

type VaultItemsLoadedMsg struct{ Items *vault.GetVaultItemsResponse }
type VaultItemsErrorMsg struct{ Error string }

type AddItemMsg struct{ Type ItemType }

type ItemSavedMsg struct{}
type ItemSaveErrorMsg struct{ Error string }

type ViewItemMsg struct {
	Item any
	Type ItemType
}

type DeleteItemMsg struct{}
type DeleteItemErrorMsg struct{ Error string }

type LoadingMsg struct{ Loading bool }
