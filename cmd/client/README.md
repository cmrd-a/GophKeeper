# GophKeeper TUI Client

A beautiful terminal user interface (TUI) client for GophKeeper built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

### 🔐 Authentication
- User registration and login
- Secure token-based authentication
- Easy mode switching between login and registration

### 🗃️ Vault Management
- **Login & Password**: Store username/password credentials
- **Text Notes**: Store secure text documents and notes
- **Credit Cards**: Store credit card information securely
- **Binary Files**: Store files and binary data

### 🎨 User Interface
- Beautiful, responsive TUI with intuitive navigation
- Real-time search and filtering
- Multiple view modes (formatted and raw)
- Contextual help and keyboard shortcuts
- Status messages and loading indicators

### ⌨️ Keyboard Navigation
- **Arrow keys** or **j/k**: Navigate lists
- **Enter**: Select/view items
- **Tab/Shift+Tab**: Navigate form fields
- **Esc**: Go back/cancel
- **q**: Quit application
- **a**: Add new item
- **d**: Delete item
- **r**: Refresh items or toggle raw view
- **/** : Search items
- **h/?**: Toggle help
- **Ctrl+R**: Switch between login/register modes

## Installation

### Build from source
```bash
go build -o gophkeeper-client ./cmd/client
```

### Run
```bash
./gophkeeper-client
```

## Configuration

The client connects to `localhost:8082` by default. You can override this by setting the `GOPHKEEPER_SERVER` environment variable:

```bash
export GOPHKEEPER_SERVER=your-server:8082
./gophkeeper-client
```

## Usage

### First Time Setup
1. Launch the client
2. Press `Ctrl+R` to switch to registration mode
3. Enter your desired username and password
4. After successful registration, switch back to login mode
5. Log in with your credentials

### Managing Vault Items

#### Adding Items
1. Press `a` from the main screen
2. Select the type of item you want to add
3. Fill in the required information
4. Press `Enter` to save

#### Viewing Items
1. Navigate to an item using arrow keys or `j/k`
2. Press `Enter` to view details
3. Press `r` to toggle between formatted and raw view
4. Press `c` to copy item data to clipboard

#### Deleting Items
1. Navigate to the item you want to delete
2. Press `d` to delete the item
3. Confirm the deletion

#### Searching
1. Press `/` to enter search mode
2. Type your search query
3. Press `Enter` to apply the filter
4. Press `Esc` to clear the search

## Security Features

- Passwords are masked in the interface by default
- Credit card numbers are masked for display
- Raw view available for when full details are needed
- All communication with the server uses TLS encryption
- Secure token-based authentication

## Troubleshooting

### Connection Issues
- Ensure the GophKeeper server is running
- Check that the server address is correct
- Verify TLS certificates are properly configured

### Authentication Issues
- Make sure you're using the correct username/password
- Try registering a new account if login fails
- Check server logs for authentication errors

### Display Issues
- Resize your terminal window if the interface appears cramped
- Ensure your terminal supports colors and Unicode characters
- Try a different terminal emulator if rendering issues persist

## Development

The TUI client is built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Huh](https://github.com/charmbracelet/huh) - Forms and inputs
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- gRPC - Server communication

### Project Structure
```
cmd/client/
├── main.go           # Main client and gRPC client wrapper
├── tui/
│   ├── app.go        # Main application and state management
│   ├── login.go      # Login/registration screen
│   ├── main.go       # Vault items list screen
│   ├── add_item.go   # Add item forms screen
│   └── view_item.go  # Item viewing screen
└── README.md
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

This project is licensed under the same license as the main GophKeeper project.