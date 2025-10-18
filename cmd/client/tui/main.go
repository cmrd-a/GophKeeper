package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cmrd-a/GophKeeper/gen/proto/v1/vault"
)

// MainScreen represents the main vault items screen
type MainScreen struct {
	width       int
	height      int
	cursor      int
	items       []VaultDisplayItem
	vaultItems  *vault.GetVaultItemsResponse
	showHelp    bool
	searchMode  bool
	searchQuery string
}

// VaultDisplayItem represents an item for display in the list
type VaultDisplayItem struct {
	ID       string
	Type     ItemType
	Title    string
	Subtitle string
	Data     any
}

// NewMainScreen creates a new main screen
func NewMainScreen() *MainScreen {
	return &MainScreen{
		items: make([]VaultDisplayItem, 0),
	}
}

// Init initializes the main screen
func (ms *MainScreen) Init() tea.Cmd {
	return nil
}

// Update handles messages for the main screen
func (ms *MainScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ms.width = msg.Width
		ms.height = msg.Height

	case tea.KeyMsg:
		if ms.searchMode {
			return ms.handleSearchInput(msg)
		}

		switch msg.String() {
		case "up", "k":
			if ms.cursor > 0 {
				ms.cursor--
			}

		case "down", "j":
			if ms.cursor < len(ms.items)-1 {
				ms.cursor++
			}

		case "enter":
			if len(ms.items) > 0 {
				item := ms.items[ms.cursor]
				return ms, func() tea.Msg {
					return ViewItemMsg{
						Item: item.Data,
						Type: item.Type,
					}
				}
			}

		case "a":
			// Show add item menu
			return ms, func() tea.Msg {
				return ShowAddMenuMsg{}
			}

		case "d":
			if len(ms.items) > 0 {
				item := ms.items[ms.cursor]
				return ms, ms.deleteItem(item)
			}

		case "/":
			ms.searchMode = true
			ms.searchQuery = ""

		case "h", "?":
			ms.showHelp = !ms.showHelp

		case "r":
			// Refresh items
			return ms, func() tea.Msg {
				return RefreshItemsMsg{}
			}
		}
	}

	return ms, nil
}

// View renders the main screen
func (ms *MainScreen) View() string {
	if ms.width == 0 || ms.height == 0 {
		return "Loading..."
	}

	var content strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(ms.width - 4).
		Render("🗃️  Vault Items")

	content.WriteString(title)
	content.WriteString("\n\n")

	// Search bar
	if ms.searchMode {
		searchBar := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Render("Search: " + ms.searchQuery + "█")
		content.WriteString(searchBar)
		content.WriteString("\n\n")
	}

	// Items list
	if len(ms.items) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			Italic(true).
			Render("No items in vault. Press 'a' to add your first item.")
		content.WriteString(emptyMsg)
	} else {
		ms.renderItemsList(&content)
	}

	// Help section
	if ms.showHelp {
		content.WriteString("\n")
		content.WriteString(ms.renderHelp())
	}

	return content.String()
}

// handleSearchInput handles search mode input
func (ms *MainScreen) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "escape":
		ms.searchMode = false
		ms.searchQuery = ""
		ms.filterItems()

	case "enter":
		ms.searchMode = false
		ms.filterItems()

	case "backspace":
		if len(ms.searchQuery) > 0 {
			ms.searchQuery = ms.searchQuery[:len(ms.searchQuery)-1]
			ms.filterItems()
		}

	default:
		if len(msg.Runes) > 0 {
			ms.searchQuery += string(msg.Runes)
			ms.filterItems()
		}
	}

	return ms, nil
}

// renderItemsList renders the list of vault items
func (ms *MainScreen) renderItemsList(content *strings.Builder) {
	maxHeight := ms.height - 10 // Leave space for header, footer, etc.
	if maxHeight < 1 {
		maxHeight = 1
	}

	start := 0
	end := len(ms.items)

	// Adjust view window if cursor is outside visible area
	if ms.cursor >= maxHeight {
		start = ms.cursor - maxHeight + 1
		end = start + maxHeight
		if end > len(ms.items) {
			end = len(ms.items)
		}
	}

	for i := start; i < end; i++ {
		item := ms.items[i]
		isSelected := i == ms.cursor

		// Item styling
		var style lipgloss.Style
		if isSelected {
			style = lipgloss.NewStyle().
				Background(lipgloss.Color("#7D56F4")).
				Foreground(lipgloss.Color("#FAFAFA")).
				Bold(true).
				Padding(0, 1)
		} else {
			style = lipgloss.NewStyle().
				Padding(0, 1)
		}

		// Type icon
		icon := ms.getTypeIcon(item.Type)

		// Render item
		itemText := fmt.Sprintf("%s %s", icon, item.Title)
		if item.Subtitle != "" {
			itemText += fmt.Sprintf("\n   %s",
				lipgloss.NewStyle().
					Foreground(lipgloss.Color("#6272A4")).
					Render(item.Subtitle))
		}

		content.WriteString(style.Render(itemText))
		content.WriteString("\n")
	}

	// Scroll indicator
	if len(ms.items) > maxHeight {
		scrollInfo := fmt.Sprintf("(%d/%d)", ms.cursor+1, len(ms.items))
		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			Render(scrollInfo))
	}
}

// renderHelp renders the help section
func (ms *MainScreen) renderHelp() string {
	help := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6272A4")).
		Padding(1, 2).
		Render(`Help:
↑/↓, j/k: Navigate items
Enter: View/Edit item
a: Add new item
d: Delete selected item
/: Search items
r: Refresh items
h/?: Toggle help
q: Quit`)

	return help
}

// getTypeIcon returns an icon for the item type
func (ms *MainScreen) getTypeIcon(itemType ItemType) string {
	switch itemType {
	case TypeLoginPassword:
		return "🔑"
	case TypeTextData:
		return "📝"
	case TypeCardData:
		return "💳"
	case TypeBinaryData:
		return "📁"
	default:
		return "📄"
	}
}

// SetVaultItems sets the vault items and converts them for display
func (ms *MainScreen) SetVaultItems(items *vault.GetVaultItemsResponse) {
	ms.vaultItems = items
	ms.convertItemsForDisplay()
	ms.cursor = 0 // Reset cursor position
}

// convertItemsForDisplay converts vault items to display format
func (ms *MainScreen) convertItemsForDisplay() {
	ms.items = make([]VaultDisplayItem, 0)

	if ms.vaultItems == nil {
		return
	}

	// Add login/password items
	for _, item := range ms.vaultItems.LoginPasswords {
		displayItem := VaultDisplayItem{
			ID:       item.Base.Id,
			Type:     TypeLoginPassword,
			Title:    fmt.Sprintf("Login: %s", item.Login),
			Subtitle: "Password entry",
			Data:     item,
		}
		ms.items = append(ms.items, displayItem)
	}

	// Add text data items
	for _, item := range ms.vaultItems.TextData {
		text := item.Text
		if len(text) > 50 {
			text = text[:47] + "..."
		}
		displayItem := VaultDisplayItem{
			ID:       item.Base.Id,
			Type:     TypeTextData,
			Title:    "Text Note",
			Subtitle: text,
			Data:     item,
		}
		ms.items = append(ms.items, displayItem)
	}

	// Add card data items
	for _, item := range ms.vaultItems.CardData {
		maskedNumber := ms.maskCardNumber(item.Number)
		displayItem := VaultDisplayItem{
			ID:       item.Base.Id,
			Type:     TypeCardData,
			Title:    fmt.Sprintf("Card: %s", maskedNumber),
			Subtitle: item.Holder,
			Data:     item,
		}
		ms.items = append(ms.items, displayItem)
	}

	// Add binary data items
	for _, item := range ms.vaultItems.BinaryData {
		size := fmt.Sprintf("(%d bytes)", len(item.Data))
		displayItem := VaultDisplayItem{
			ID:       item.Base.Id,
			Type:     TypeBinaryData,
			Title:    "Binary File",
			Subtitle: size,
			Data:     item,
		}
		ms.items = append(ms.items, displayItem)
	}
}

// maskCardNumber masks a credit card number for display
func (ms *MainScreen) maskCardNumber(number string) string {
	if len(number) < 4 {
		return number
	}

	masked := strings.Repeat("*", len(number)-4)
	return masked + number[len(number)-4:]
}

// filterItems filters items based on search query
func (ms *MainScreen) filterItems() {
	if ms.searchQuery == "" {
		ms.convertItemsForDisplay()
		return
	}

	query := strings.ToLower(ms.searchQuery)
	filtered := make([]VaultDisplayItem, 0)

	for _, item := range ms.items {
		if strings.Contains(strings.ToLower(item.Title), query) ||
			strings.Contains(strings.ToLower(item.Subtitle), query) {
			filtered = append(filtered, item)
		}
	}

	ms.items = filtered
	if ms.cursor >= len(ms.items) {
		ms.cursor = len(ms.items) - 1
	}
	if ms.cursor < 0 {
		ms.cursor = 0
	}
}

// deleteItem creates a command to delete an item
func (ms *MainScreen) deleteItem(item VaultDisplayItem) tea.Cmd {
	return func() tea.Msg {
		return DeleteItemAttemptMsg{
			ID:   item.ID,
			Type: ms.getItemTypeString(item.Type),
		}
	}
}

// getItemTypeString converts ItemType to string for API
func (ms *MainScreen) getItemTypeString(itemType ItemType) string {
	switch itemType {
	case TypeLoginPassword:
		return "login_password"
	case TypeTextData:
		return "text_data"
	case TypeCardData:
		return "card_data"
	case TypeBinaryData:
		return "binary_data"
	default:
		return "unknown"
	}
}

// Messages
type ShowAddMenuMsg struct{}
type RefreshItemsMsg struct{}
type DeleteItemAttemptMsg struct {
	ID   string
	Type string
}
