package tui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"log"

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

	// Form state tracking
	formInitialized bool
	lastError       string

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
	ais.lastError = ""
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

		// Handle form navigation and actions
		if ais.form != nil && ais.formInitialized {
			// Handle escape to return to menu
			if msg.String() == "esc" {
				ais.Reset()
				return ais, nil
			}

			// Handle quit
			if msg.String() == "ctrl+c" {
				return ais, tea.Quit
			}

			// Clear any previous errors when user interacts
			if msg.String() != "" && ais.lastError != "" {
				ais.lastError = ""
			}

			// Handle save - try multiple conditions to catch form completion
			if msg.String() == "ctrl+s" {
				// Manual save with Ctrl+S
				log.Printf("Manual save triggered")
				if ais.isFormValid() {
					return ais, ais.saveItem()
				} else {
					ais.lastError = "Please fill all required fields"
				}
			} else if ais.form.State == huh.StateCompleted && (msg.String() == "enter" || msg.String() == " ") {
				log.Printf("Form completed save triggered")
				return ais, ais.saveItem()
			} else if msg.String() == "enter" && ais.isFormValid() {
				// Try to save on enter if form looks valid
				log.Printf("Enter key save attempt")
				return ais, ais.saveItem()
			}

			if ais.form.State != huh.StateCompleted {
				log.Printf("Form state: %v, key: %s", ais.form.State, msg.String())
			}

			// Pass all key messages to the form for navigation and input
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
		ais.lastError = ""
		ais.buildForm()
		if ais.form != nil && ais.formInitialized {
			// Initialize and focus the form
			cmd := ais.form.Init()
			return ais, cmd
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

	var formView string
	var help string

	if ais.form != nil {
		formView = ais.form.View()

		switch ais.form.State {
		case huh.StateCompleted:
			help = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#50FA7B")).
				Bold(true).
				MarginTop(2).
				Render("✓ Form Complete • Enter: Save Item • Esc: Back to menu")
		case huh.StateAborted:
			help = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF5F87")).
				MarginTop(2).
				Render("Form cancelled • Esc: Back to menu")
		default:
			stateDebug := fmt.Sprintf("State: %v", ais.form.State)
			saveHint := ""
			if ais.isFormValid() {
				saveHint = " • Enter: Save"
			}
			help = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6272A4")).
				MarginTop(2).
				Render("Tab/Shift+Tab: Navigate • Ctrl+S: Save" + saveHint + " • " + stateDebug + " • Esc: Cancel")
		}
	} else {
		formView = "Loading form..."
		help = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			MarginTop(2).
			Render("Form failed to initialize • Esc: Back to menu")
	}

	// Show any error messages
	var errorView string
	if ais.lastError != "" {
		errorView = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Background(lipgloss.Color("#2D1B00")).
			Padding(0, 1).
			MarginTop(1).
			Render("Error: " + ais.lastError)
	}

	var contentParts []string
	contentParts = append(contentParts, titleView, formView)

	if errorView != "" {
		contentParts = append(contentParts, errorView)
	}

	contentParts = append(contentParts, help)

	content := lipgloss.JoinVertical(lipgloss.Left, contentParts...)

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
		// Create form with proper configuration
		formWidth := 60
		formHeight := 20
		if ais.width > 0 && ais.width-4 < formWidth {
			formWidth = ais.width - 4
		}
		if ais.height > 0 && ais.height-10 < formHeight {
			formHeight = ais.height - 10
		}

		// Ensure minimum dimensions
		if formWidth < 40 {
			formWidth = 40
		}
		if formHeight < 10 {
			formHeight = 10
		}

		group := huh.NewGroup(fields...)
		ais.form = huh.NewForm(group).
			WithWidth(formWidth).
			WithHeight(formHeight).
			WithTheme(huh.ThemeBase()).
			WithShowErrors(true).
			WithShowHelp(true)

		ais.formInitialized = true
		ais.lastError = ""
	} else {
		ais.form = nil
		ais.formInitialized = false
		ais.lastError = "No form fields available for this item type"
	}
}

// saveItem creates a command to save the item
func (ais *AddItemScreen) saveItem() tea.Cmd {
	log.Printf("saveItem called for type: %v", ais.itemType)
	switch ais.itemType {
	case TypeLoginPassword:
		return func() tea.Msg {
			log.Printf("Saving login/password: login=%s", ais.loginUser)
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
			textPreview := ais.textContent
			if len(textPreview) > 50 {
				textPreview = textPreview[:50] + "..."
			}
			log.Printf("Saving text data: %s", textPreview)
			return SaveItemAttemptMsg{
				Type: TypeTextData,
				Data: map[string]any{
					"text": ais.textContent,
				},
			}
		}

	case TypeCardData:
		return func() tea.Msg {
			log.Printf("Saving card data: holder=%s", ais.cardHolder)
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
			log.Printf("Saving binary data: %d bytes", len(ais.binaryData))
			return SaveItemAttemptMsg{
				Type: TypeBinaryData,
				Data: map[string]any{
					"data": ais.binaryData,
				},
			}
		}
	}

	log.Printf("Unknown item type: %v", ais.itemType)
	return nil
}

// isFormValid checks if the current form has all required fields filled
func (ais *AddItemScreen) isFormValid() bool {
	switch ais.itemType {
	case TypeLoginPassword:
		return strings.TrimSpace(ais.loginUser) != "" && strings.TrimSpace(ais.loginPassword) != ""
	case TypeTextData:
		return strings.TrimSpace(ais.textContent) != ""
	case TypeCardData:
		return strings.TrimSpace(ais.cardNumber) != "" &&
			strings.TrimSpace(ais.cardHolder) != "" &&
			strings.TrimSpace(ais.cardExpiry) != "" &&
			strings.TrimSpace(ais.cardCVV) != ""
	case TypeBinaryData:
		return len(ais.binaryData) > 0
	}
	return false
}

// Reset resets the form fields
func (ais *AddItemScreen) Reset() {
	ais.showMenu = true
	ais.menuCursor = 0
	ais.form = nil
	ais.formInitialized = false
	ais.lastError = ""

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
