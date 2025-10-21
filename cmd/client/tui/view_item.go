package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cmrd-a/GophKeeper/gen/proto/v1/vault"
)

// ViewItemScreen represents the view item screen
type ViewItemScreen struct {
	width    int
	height   int
	item     any
	itemType ItemType
	showRaw  bool
}

// NewViewItemScreen creates a new view item screen
func NewViewItemScreen() *ViewItemScreen {
	return &ViewItemScreen{}
}

// Init initializes the view item screen
func (vis *ViewItemScreen) Init() tea.Cmd {
	return nil
}

// SetItem sets the item to view
func (vis *ViewItemScreen) SetItem(item any, itemType ItemType) {
	vis.item = item
	vis.itemType = itemType
	vis.showRaw = false
}

// Update handles messages for the view item screen
func (vis *ViewItemScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		vis.width = msg.Width
		vis.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			vis.showRaw = !vis.showRaw

		case "d":
			if vis.item != nil {
				return vis, vis.deleteCurrentItem()
			}

		case "c":
			// Copy to clipboard functionality could be added here
			return vis, func() tea.Msg {
				return CopyToClipboardMsg{Data: vis.getItemText()}
			}
		}
	}

	return vis, nil
}

// View renders the view item screen
func (vis *ViewItemScreen) View() string {
	if vis.width == 0 || vis.height == 0 {
		return "Loading..."
	}

	if vis.item == nil {
		return "No item selected"
	}

	var content strings.Builder

	// Title with item type
	title := vis.getItemTitle()
	titleView := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(vis.width - 4).
		Render(title)

	content.WriteString(titleView)
	content.WriteString("\n\n")

	// Item content
	if vis.showRaw {
		content.WriteString(vis.renderRawView())
	} else {
		content.WriteString(vis.renderFormattedView())
	}

	// Footer with actions
	content.WriteString("\n\n")
	content.WriteString(vis.renderActions())

	return content.String()
}

// getItemTitle returns the title for the current item
func (vis *ViewItemScreen) getItemTitle() string {
	icon := vis.getTypeIcon()

	switch vis.itemType {
	case TypeLoginPassword:
		if login, ok := vis.item.(*vault.LoginPassword); ok {
			return fmt.Sprintf("%s Login: %s", icon, login.Login)
		}
		return fmt.Sprintf("%s Login & Password", icon)

	case TypeTextData:
		return fmt.Sprintf("%s Text Note", icon)

	case TypeCardData:
		if card, ok := vis.item.(*vault.CardData); ok {
			maskedNumber := vis.maskCardNumber(card.Number)
			return fmt.Sprintf("%s Card: %s", icon, maskedNumber)
		}
		return fmt.Sprintf("%s Credit Card", icon)

	case TypeBinaryData:
		if binary, ok := vis.item.(*vault.BinaryData); ok {
			size := fmt.Sprintf("(%d bytes)", len(binary.Data))
			return fmt.Sprintf("%s Binary File %s", icon, size)
		}
		return fmt.Sprintf("%s Binary File", icon)

	default:
		return fmt.Sprintf("%s Vault Item", icon)
	}
}

// getTypeIcon returns the icon for the item type
func (vis *ViewItemScreen) getTypeIcon() string {
	switch vis.itemType {
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

// renderFormattedView renders the item in a user-friendly format
func (vis *ViewItemScreen) renderFormattedView() string {
	var content strings.Builder

	switch vis.itemType {
	case TypeLoginPassword:
		if login, ok := vis.item.(*vault.LoginPassword); ok {
			content.WriteString(vis.renderLoginPassword(login))
		}

	case TypeTextData:
		if text, ok := vis.item.(*vault.TextData); ok {
			content.WriteString(vis.renderTextData(text))
		}

	case TypeCardData:
		if card, ok := vis.item.(*vault.CardData); ok {
			content.WriteString(vis.renderCardData(card))
		}

	case TypeBinaryData:
		if binary, ok := vis.item.(*vault.BinaryData); ok {
			content.WriteString(vis.renderBinaryData(binary))
		}
	}

	return content.String()
}

// renderLoginPassword renders login/password data
func (vis *ViewItemScreen) renderLoginPassword(login *vault.LoginPassword) string {
	var content strings.Builder

	// Metadata
	content.WriteString(vis.renderMetadata(login.Base))
	content.WriteString("\n")

	// Login details
	fieldStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#8BE9FD"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2"))

	content.WriteString(fieldStyle.Render("Username: "))
	content.WriteString(valueStyle.Render(login.Login))
	content.WriteString("\n\n")

	content.WriteString(fieldStyle.Render("Password: "))
	// Show password masked by default, could add toggle
	masked := strings.Repeat("•", len(login.Password))
	content.WriteString(valueStyle.Render(masked))
	content.WriteString(" ")
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272A4")).
		Italic(true).
		Render("(press 'r' to show raw)"))

	return content.String()
}

// renderTextData renders text data
func (vis *ViewItemScreen) renderTextData(text *vault.TextData) string {
	var content strings.Builder

	// Metadata
	content.WriteString(vis.renderMetadata(text.Base))
	content.WriteString("\n")

	// Text content
	contentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6272A4")).
		Padding(1, 2).
		Width(vis.width - 8)

	content.WriteString(contentBox.Render(text.Text))

	return content.String()
}

// renderCardData renders card data
func (vis *ViewItemScreen) renderCardData(card *vault.CardData) string {
	var content strings.Builder

	// Metadata
	content.WriteString(vis.renderMetadata(card.Base))
	content.WriteString("\n")

	// Card details
	fieldStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#8BE9FD"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2"))

	content.WriteString(fieldStyle.Render("Card Number: "))
	content.WriteString(valueStyle.Render(vis.maskCardNumber(card.Number)))
	content.WriteString("\n\n")

	content.WriteString(fieldStyle.Render("Cardholder: "))
	content.WriteString(valueStyle.Render(card.Holder))
	content.WriteString("\n\n")

	content.WriteString(fieldStyle.Render("Expiry: "))
	content.WriteString(valueStyle.Render(card.Expire))
	content.WriteString("\n\n")

	content.WriteString(fieldStyle.Render("CVV: "))
	content.WriteString(valueStyle.Render("•••"))
	content.WriteString(" ")
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272A4")).
		Italic(true).
		Render("(press 'r' to show raw)"))

	return content.String()
}

// renderBinaryData renders binary data information
func (vis *ViewItemScreen) renderBinaryData(binary *vault.BinaryData) string {
	var content strings.Builder

	// Metadata
	content.WriteString(vis.renderMetadata(binary.Base))
	content.WriteString("\n")

	// Binary data info
	fieldStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#8BE9FD"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2"))

	content.WriteString(fieldStyle.Render("Size: "))
	content.WriteString(valueStyle.Render(vis.formatBytes(len(binary.Data))))
	content.WriteString("\n\n")

	content.WriteString(fieldStyle.Render("Type: "))
	content.WriteString(valueStyle.Render("Binary Data"))
	content.WriteString("\n\n")

	// Preview first few bytes
	preview := vis.getBinaryPreview(binary.Data)
	if preview != "" {
		content.WriteString(fieldStyle.Render("Preview: "))
		content.WriteString("\n")
		previewBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6272A4")).
			Padding(1, 2).
			Width(vis.width - 8)
		content.WriteString(previewBox.Render(preview))
	}

	return content.String()
}

// renderRawView renders the raw item data
func (vis *ViewItemScreen) renderRawView() string {
	var raw string

	switch vis.itemType {
	case TypeLoginPassword:
		if login, ok := vis.item.(*vault.LoginPassword); ok {
			raw = fmt.Sprintf("ID: %s\nLogin: %s\nPassword: %s\nCreated: %s\nUpdated: %s",
				login.Base.Id,
				login.Login,
				login.Password,
				vis.formatTimestamp(login.Base.CreatedAt),
				vis.formatTimestamp(login.Base.UpdatedAt))
		}

	case TypeTextData:
		if text, ok := vis.item.(*vault.TextData); ok {
			raw = fmt.Sprintf("ID: %s\nText: %s\nCreated: %s\nUpdated: %s",
				text.Base.Id,
				text.Text,
				vis.formatTimestamp(text.Base.CreatedAt),
				vis.formatTimestamp(text.Base.UpdatedAt))
		}

	case TypeCardData:
		if card, ok := vis.item.(*vault.CardData); ok {
			raw = fmt.Sprintf("ID: %s\nNumber: %s\nHolder: %s\nExpiry: %s\nCVV: %s\nCreated: %s\nUpdated: %s",
				card.Base.Id,
				card.Number,
				card.Holder,
				card.Expire,
				card.Cvv,
				vis.formatTimestamp(card.Base.CreatedAt),
				vis.formatTimestamp(card.Base.UpdatedAt))
		}

	case TypeBinaryData:
		if binary, ok := vis.item.(*vault.BinaryData); ok {
			raw = fmt.Sprintf("ID: %s\nSize: %d bytes\nCreated: %s\nUpdated: %s\n\nHex dump:\n%s",
				binary.Base.Id,
				len(binary.Data),
				vis.formatTimestamp(binary.Base.CreatedAt),
				vis.formatTimestamp(binary.Base.UpdatedAt),
				vis.getHexDump(binary.Data))
		}
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF5F87")).
		Padding(1, 2).
		Width(vis.width - 8).
		Render(raw)
}

// renderMetadata renders common metadata for items
func (vis *ViewItemScreen) renderMetadata(base *vault.VaultItem) string {
	metaStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272A4")).
		Italic(true)

	var content strings.Builder
	content.WriteString(metaStyle.Render("ID: " + base.Id))
	content.WriteString("\n")
	content.WriteString(metaStyle.Render("Created: " + vis.formatTimestamp(base.CreatedAt)))
	content.WriteString("\n")
	content.WriteString(metaStyle.Render("Updated: " + vis.formatTimestamp(base.UpdatedAt)))

	return content.String()
}

// renderActions renders available actions
func (vis *ViewItemScreen) renderActions() string {
	actions := []string{
		"r: Toggle raw view",
		"c: Copy to clipboard",
		"d: Delete item",
		"Esc: Back",
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272A4")).
		Render(strings.Join(actions, " • "))
}

// Helper functions

// maskCardNumber masks a credit card number
func (vis *ViewItemScreen) maskCardNumber(number string) string {
	if len(number) < 4 {
		return number
	}
	return strings.Repeat("*", len(number)-4) + number[len(number)-4:]
}

// formatBytes formats byte size in human readable format
func (vis *ViewItemScreen) formatBytes(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatTimestamp formats a timestamp for display
func (vis *ViewItemScreen) formatTimestamp(ts any) string {
	// This would need to be implemented based on the actual timestamp format
	// from the protobuf message
	return "N/A" // Placeholder
}

// getBinaryPreview generates a preview of binary data
func (vis *ViewItemScreen) getBinaryPreview(data []byte) string {
	if len(data) == 0 {
		return "Empty file"
	}

	// Check if it's likely text
	printable := 0
	for _, b := range data[:min(100, len(data))] {
		if b >= 32 && b <= 126 || b == '\n' || b == '\r' || b == '\t' {
			printable++
		}
	}

	if printable > len(data[:min(100, len(data))])*3/4 {
		// Likely text, show first 200 chars
		preview := string(data[:min(200, len(data))])
		if len(data) > 200 {
			preview += "..."
		}
		return preview
	}

	// Binary data, show hex dump of first 64 bytes
	return vis.getHexDump(data[:min(64, len(data))])
}

// getHexDump generates a hex dump of binary data
func (vis *ViewItemScreen) getHexDump(data []byte) string {
	if len(data) == 0 {
		return "No data"
	}

	var dump strings.Builder
	for i := 0; i < len(data); i += 16 {
		// Address
		dump.WriteString(fmt.Sprintf("%04x: ", i))

		// Hex bytes
		for j := 0; j < 16; j++ {
			if i+j < len(data) {
				dump.WriteString(fmt.Sprintf("%02x ", data[i+j]))
			} else {
				dump.WriteString("   ")
			}
		}

		// ASCII representation
		dump.WriteString(" |")
		for j := 0; j < 16 && i+j < len(data); j++ {
			b := data[i+j]
			if b >= 32 && b <= 126 {
				dump.WriteByte(b)
			} else {
				dump.WriteByte('.')
			}
		}
		dump.WriteString("|\n")
	}

	return dump.String()
}

// getItemText returns the textual representation of the item for copying
func (vis *ViewItemScreen) getItemText() string {
	switch vis.itemType {
	case TypeLoginPassword:
		if login, ok := vis.item.(*vault.LoginPassword); ok {
			return fmt.Sprintf("Username: %s\nPassword: %s", login.Login, login.Password)
		}
	case TypeTextData:
		if text, ok := vis.item.(*vault.TextData); ok {
			return text.Text
		}
	case TypeCardData:
		if card, ok := vis.item.(*vault.CardData); ok {
			return fmt.Sprintf("Number: %s\nHolder: %s\nExpiry: %s\nCVV: %s",
				card.Number, card.Holder, card.Expire, card.Cvv)
		}
	}
	return ""
}

// deleteCurrentItem creates a delete command for the current item
func (vis *ViewItemScreen) deleteCurrentItem() tea.Cmd {
	var id string
	var itemType string

	switch vis.itemType {
	case TypeLoginPassword:
		if login, ok := vis.item.(*vault.LoginPassword); ok {
			id = login.Base.Id
			itemType = "login_password"
		}
	case TypeTextData:
		if text, ok := vis.item.(*vault.TextData); ok {
			id = text.Base.Id
			itemType = "text_data"
		}
	case TypeCardData:
		if card, ok := vis.item.(*vault.CardData); ok {
			id = card.Base.Id
			itemType = "card_data"
		}
	case TypeBinaryData:
		if binary, ok := vis.item.(*vault.BinaryData); ok {
			id = binary.Base.Id
			itemType = "binary_data"
		}
	}

	if id != "" {
		return func() tea.Msg {
			return DeleteItemAttemptMsg{
				ID:   id,
				Type: itemType,
			}
		}
	}

	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Messages
type CopyToClipboardMsg struct {
	Data string
}
