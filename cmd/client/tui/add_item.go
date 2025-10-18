package tui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// AddItemScreen represents the add item screen
type AddItemScreen struct {
	width      int
	height     int
	itemType   ItemType
	form       *huh.Form
	showMenu   bool
	menuCursor int

	// Form fields for different item types
	// Login/Password
	loginUser     string
	loginPassword string

	// Text data
	textContent string

	// Card data
	cardNumber string
	cardHolder string
	cardExpiry string
	cardCVV    string

	// Binary data
	binaryFilePath string
	binaryData     []byte
}

// NewAddItemScreen creates a new add item screen
func NewAddItemScreen() *AddItemScreen {
	return &AddItemScreen{
		showMenu:   true,
		menuCursor: 0,
	}
}

// Init initializes the add item screen
func (ais *AddItemScreen) Init() tea.Cmd {
	return nil
}

// SetItemType sets the item type and builds the appropriate form
func (ais *AddItemScreen) SetItemType(itemType ItemType) {
	ais.itemType = itemType
	ais.showMenu = false
	ais.buildForm()
}

// Update handles messages for the add item screen
func (ais *AddItemScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ais.width = msg.Width
		ais.height = msg.Height

	case tea.KeyMsg:
		if ais.showMenu {
			return ais.handleMenuInput(msg)
		}

		switch msg.String() {
		case "esc":
			if ais.form != nil {
				ais.showMenu = true
				ais.form = nil
				return ais, nil
			}

		case "enter":
			if ais.form != nil && ais.form.State == huh.StateCompleted {
				return ais, ais.saveItem()
			}
		}

		// Update form if it exists
		if ais.form != nil {
			form, cmd := ais.form.Update(msg)
			if f, ok := form.(*huh.Form); ok {
				ais.form = f
			}
			return ais, cmd
		}
	}

	return ais, nil
}

// View renders the add item screen
func (ais *AddItemScreen) View() string {
	if ais.width == 0 || ais.height == 0 {
		return "Loading..."
	}

	if ais.showMenu {
		return ais.renderMenu()
	}

	return ais.renderForm()
}

// handleMenuInput handles input when the menu is shown
func (ais *AddItemScreen) handleMenuInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if ais.menuCursor > 0 {
			ais.menuCursor--
		}

	case "down", "j":
		if ais.menuCursor < 3 { // 4 item types (0-3)
			ais.menuCursor++
		}

	case "enter":
		ais.itemType = ItemType(ais.menuCursor)
		ais.showMenu = false
		ais.buildForm()
		if ais.form != nil {
			return ais, ais.form.Init()
		}
	}

	return ais, nil
}

// renderMenu renders the item type selection menu
func (ais *AddItemScreen) renderMenu() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		MarginBottom(2).
		Render("➕ Add New Item")

	menuItems := []struct {
		icon string
		name string
		desc string
	}{
		{"🔑", "Login & Password", "Store username and password credentials"},
		{"📝", "Text Note", "Store secure text notes and documents"},
		{"💳", "Credit Card", "Store credit card information"},
		{"📁", "Binary File", "Store files and binary data"},
	}

	var menu strings.Builder
	menu.WriteString(title + "\n\n")
	menu.WriteString("Select item type to add:\n\n")

	for i, item := range menuItems {
		var style lipgloss.Style
		if i == ais.menuCursor {
			style = lipgloss.NewStyle().
				Background(lipgloss.Color("#7D56F4")).
				Foreground(lipgloss.Color("#FAFAFA")).
				Bold(true).
				Padding(0, 1)
		} else {
			style = lipgloss.NewStyle().Padding(0, 1)
		}

		itemText := fmt.Sprintf("%s %s", item.icon, item.name)
		menu.WriteString(style.Render(itemText) + "\n")

		if i == ais.menuCursor {
			menu.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6272A4")).
				Italic(true).
				MarginLeft(4).
				Render(item.desc) + "\n")
		}
		menu.WriteString("\n")
	}

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272A4")).
		MarginTop(2).
		Render("↑/↓: Navigate • Enter: Select • Esc: Cancel")

	content := menu.String() + help

	// Center content
	if ais.width > 0 {
		contentWidth := lipgloss.Width(content)
		if contentWidth < ais.width {
			padding := (ais.width - contentWidth) / 2
			content = lipgloss.NewStyle().
				PaddingLeft(padding).
				Render(content)
		}
	}

	return content
}

// renderForm renders the form for the selected item type
func (ais *AddItemScreen) renderForm() string {
	if ais.form == nil {
		return "Loading form..."
	}

	var title string
	switch ais.itemType {
	case TypeLoginPassword:
		title = "🔑 Add Login & Password"
	case TypeTextData:
		title = "📝 Add Text Note"
	case TypeCardData:
		title = "💳 Add Credit Card"
	case TypeBinaryData:
		title = "📁 Add Binary File"
	}

	titleView := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		MarginBottom(2).
		Render(title)

	formView := ais.form.View()

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272A4")).
		MarginTop(2).
		Render("Tab/Shift+Tab: Navigate • Enter: Save • Esc: Back to menu")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		titleView,
		formView,
		help,
	)

	// Center content
	if ais.width > 0 {
		contentWidth := lipgloss.Width(content)
		if contentWidth < ais.width {
			padding := (ais.width - contentWidth) / 2
			content = lipgloss.NewStyle().
				PaddingLeft(padding).
				Render(content)
		}
	}

	return content
}

// buildForm builds the form based on the selected item type
func (ais *AddItemScreen) buildForm() {
	var fields []huh.Field

	switch ais.itemType {
	case TypeLoginPassword:
		fields = []huh.Field{
			huh.NewInput().
				Title("Username/Email").
				Value(&ais.loginUser).
				Placeholder("Enter username or email").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("username cannot be empty")
					}
					return nil
				}),
			huh.NewInput().
				Title("Password").
				Value(&ais.loginPassword).
				Placeholder("Enter password").
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("password cannot be empty")
					}
					return nil
				}),
		}

	case TypeTextData:
		fields = []huh.Field{
			huh.NewText().
				Title("Text Content").
				Value(&ais.textContent).
				Placeholder("Enter your text content here...").
				Lines(5).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("text content cannot be empty")
					}
					return nil
				}),
		}

	case TypeCardData:
		fields = []huh.Field{
			huh.NewInput().
				Title("Card Number").
				Value(&ais.cardNumber).
				Placeholder("1234 5678 9012 3456").
				Validate(func(s string) error {
					// Remove spaces and validate card number
					cleaned := strings.ReplaceAll(s, " ", "")
					if len(cleaned) < 13 || len(cleaned) > 19 {
						return fmt.Errorf("card number must be 13-19 digits")
					}
					if _, err := strconv.ParseUint(cleaned, 10, 64); err != nil {
						return fmt.Errorf("card number must contain only digits")
					}
					return nil
				}),
			huh.NewInput().
				Title("Cardholder Name").
				Value(&ais.cardHolder).
				Placeholder("John Doe").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("cardholder name cannot be empty")
					}
					return nil
				}),
			huh.NewInput().
				Title("Expiry Date").
				Value(&ais.cardExpiry).
				Placeholder("MM/YY").
				Validate(func(s string) error {
					if !strings.Contains(s, "/") || len(s) != 5 {
						return fmt.Errorf("expiry date must be in MM/YY format")
					}
					parts := strings.Split(s, "/")
					if len(parts) != 2 {
						return fmt.Errorf("expiry date must be in MM/YY format")
					}
					month, err := strconv.Atoi(parts[0])
					if err != nil || month < 1 || month > 12 {
						return fmt.Errorf("invalid month")
					}
					year, err := strconv.Atoi(parts[1])
					if err != nil || year < 0 || year > 99 {
						return fmt.Errorf("invalid year")
					}
					return nil
				}),
			huh.NewInput().
				Title("CVV").
				Value(&ais.cardCVV).
				Placeholder("123").
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if len(s) < 3 || len(s) > 4 {
						return fmt.Errorf("CVV must be 3-4 digits")
					}
					if _, err := strconv.ParseUint(s, 10, 64); err != nil {
						return fmt.Errorf("CVV must contain only digits")
					}
					return nil
				}),
		}

	case TypeBinaryData:
		fields = []huh.Field{
			huh.NewInput().
				Title("File Path").
				Value(&ais.binaryFilePath).
				Placeholder("/path/to/file.txt").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("file path cannot be empty")
					}
					// Try to read the file
					data, err := os.ReadFile(s)
					if err != nil {
						return fmt.Errorf("cannot read file: %v", err)
					}
					ais.binaryData = data
					return nil
				}),
		}
	}

	if len(fields) > 0 {
		ais.form = huh.NewForm(
			huh.NewGroup(fields...),
		).WithWidth(60).WithHeight(15)
	}
}

// saveItem creates a command to save the item
func (ais *AddItemScreen) saveItem() tea.Cmd {
	switch ais.itemType {
	case TypeLoginPassword:
		return func() tea.Msg {
			return SaveItemAttemptMsg{
				Type: TypeLoginPassword,
				Data: map[string]any{
					"login":    ais.loginUser,
					"password": ais.loginPassword,
				},
			}
		}

	case TypeTextData:
		return func() tea.Msg {
			return SaveItemAttemptMsg{
				Type: TypeTextData,
				Data: map[string]any{
					"text": ais.textContent,
				},
			}
		}

	case TypeCardData:
		return func() tea.Msg {
			return SaveItemAttemptMsg{
				Type: TypeCardData,
				Data: map[string]any{
					"number": ais.cardNumber,
					"holder": ais.cardHolder,
					"expire": ais.cardExpiry,
					"cvv":    ais.cardCVV,
				},
			}
		}

	case TypeBinaryData:
		return func() tea.Msg {
			return SaveItemAttemptMsg{
				Type: TypeBinaryData,
				Data: map[string]any{
					"data": ais.binaryData,
				},
			}
		}
	}

	return nil
}

// Reset resets the form fields
func (ais *AddItemScreen) Reset() {
	ais.showMenu = true
	ais.menuCursor = 0
	ais.form = nil

	// Clear all form fields
	ais.loginUser = ""
	ais.loginPassword = ""
	ais.textContent = ""
	ais.cardNumber = ""
	ais.cardHolder = ""
	ais.cardExpiry = ""
	ais.cardCVV = ""
	ais.binaryFilePath = ""
	ais.binaryData = nil
}

// Messages
type SaveItemAttemptMsg struct {
	Type ItemType
	Data map[string]any
}
