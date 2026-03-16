package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli-auth/internal/auth"
	"github.com/cli-auth/internal/session"
	"github.com/cli-auth/internal/user"
	"github.com/pquerna/otp"
)

// appState tracks which screen is active.
type appState int

const (
	stateLoading appState = iota
	stateMenu
	stateRegister
	stateLogin
	stateTOTPPrompt
	stateDashboard
	stateEnableTOTP
	stateConfirmTOTP
	stateDisableTOTP
)

// menuItem represents a selectable menu entry.
type menuItem struct {
	label string
	desc  string
}

// Model is the top-level bubbletea model.
type Model struct {
	// Dependencies
	authSvc      *auth.Service
	sessionStore *session.Store

	// State
	state    appState
	err      string
	info     string
	quitting bool

	// Menu
	menuItems []menuItem
	menuIdx   int

	// Text inputs
	inputs    []textinput.Model
	inputIdx  int
	inputMode string // "register", "login", "totp", "enable2fa", "disable2fa"

	// Auth state
	currentUser  *user.User
	sessionToken string
	expiresAt    string

	// TOTP enrollment
	totpKey    *otp.Key
	totpSecret string
	totpQR     string

	// Stored credentials for TOTP flow
	pendingUsername string
	pendingPassword string

	// Dashboard command input
	dashInput textinput.Model
	dashCmd   string

	// Window size
	width  int
	height int
}

// NewModel creates the initial app model.
func NewModel(authSvc *auth.Service, sessionStore *session.Store) Model {
	m := Model{
		authSvc:      authSvc,
		sessionStore: sessionStore,
		state:        stateLoading,
	}
	m.initMenu()
	m.initDashInput()
	return m
}

func (m *Model) initMenu() {
	m.menuItems = []menuItem{
		{label: "register", desc: "Create a new account"},
		{label: "login", desc: "Login with your credentials"},
		{label: "help", desc: "Show available commands"},
		{label: "exit", desc: "Quit the program"},
	}
	m.menuIdx = 0
}

func (m *Model) initDashInput() {
	ti := textinput.New()
	ti.Placeholder = "type a command..."
	ti.Prompt = "❯ "
	ti.CharLimit = 64
	m.dashInput = ti
}

func (m *Model) initRegisterInputs() {
	username := textinput.New()
	username.Placeholder = "username"
	username.Prompt = "  "
	username.CharLimit = 64
	username.Focus()

	password := textinput.New()
	password.Placeholder = "password (min 8 chars)"
	password.Prompt = "  "
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '•'
	password.CharLimit = 128

	confirm := textinput.New()
	confirm.Placeholder = "confirm password"
	confirm.Prompt = "  "
	confirm.EchoMode = textinput.EchoPassword
	confirm.EchoCharacter = '•'
	confirm.CharLimit = 128

	m.inputs = []textinput.Model{username, password, confirm}
	m.inputIdx = 0
	m.inputMode = "register"
}

func (m *Model) initLoginInputs() {
	username := textinput.New()
	username.Placeholder = "username"
	username.Prompt = "  "
	username.CharLimit = 64
	username.Focus()

	password := textinput.New()
	password.Placeholder = "password"
	password.Prompt = "  "
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '•'
	password.CharLimit = 128

	m.inputs = []textinput.Model{username, password}
	m.inputIdx = 0
	m.inputMode = "login"
}

func (m *Model) initTOTPInput() {
	code := textinput.New()
	code.Placeholder = "6-digit code"
	code.Prompt = "  "
	code.CharLimit = 6
	code.Focus()

	m.inputs = []textinput.Model{code}
	m.inputIdx = 0
	m.inputMode = "totp"
}

func (m *Model) initEnable2FAInput() {
	code := textinput.New()
	code.Placeholder = "enter code from authenticator"
	code.Prompt = "  "
	code.CharLimit = 6
	code.Focus()

	m.inputs = []textinput.Model{code}
	m.inputIdx = 0
	m.inputMode = "enable2fa"
}

func (m *Model) initDisable2FAInput() {
	code := textinput.New()
	code.Placeholder = "enter current TOTP code"
	code.Prompt = "  "
	code.CharLimit = 6
	code.Focus()

	m.inputs = []textinput.Model{code}
	m.inputIdx = 0
	m.inputMode = "disable2fa"
}

// Init starts the app by trying to restore a session.
func (m Model) Init() tea.Cmd {
	return restoreSessionCmd(
		m.authSvc,
		m.sessionStore.LoadToken,
		m.sessionStore.Timeout().String(),
	)
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Global quit
		if msg.Type == tea.KeyCtrlC {
			m.quitting = true
			return m, tea.Quit
		}

	case sessionExpiredMsg:
		m.currentUser = nil
		m.sessionToken = ""
		m.expiresAt = ""
		m.state = stateMenu
		m.err = errorStyle.Render("✗ Session expired. Please login again.")
		m.info = ""
		return m, nil

	case dashActionMsg:
		return m.handleDashboardCmd(msg.action)

	case noSessionMsg:
		m.state = stateMenu
		return m, nil

	case sessionRestoredMsg:
		m.currentUser = msg.user
		m.sessionToken = msg.token
		m.expiresAt = msg.expires
		m.state = stateDashboard
		m.dashInput.Focus()
		m.info = "Session restored"
		return m, nil

	case registerMsg:
		m.err = ""
		m.info = successStyle.Render("✓ Account created successfully! You can now login.")
		m.state = stateMenu
		return m, nil

	case loginMsg:
		m.err = ""
		m.currentUser = msg.result.User
		m.sessionToken = msg.result.Token
		m.expiresAt = msg.result.ExpiresAt.In(IST).Format("2006-01-02 15:04:05 IST")
		m.state = stateDashboard
		m.dashInput.Focus()
		m.dashInput.SetValue("")
		m.info = ""
		m.pendingUsername = ""
		m.pendingPassword = ""
		return m, nil

	case totpRequiredMsg:
		m.state = stateTOTPPrompt
		m.initTOTPInput()
		m.err = ""
		m.info = ""
		return m, nil

	case logoutMsg:
		m.currentUser = nil
		m.sessionToken = ""
		m.expiresAt = ""
		m.state = stateMenu
		m.err = ""
		m.info = successStyle.Render("✓ Logged out successfully")
		return m, nil

	case enableTOTPMsg:
		m.totpKey = msg.key
		m.totpSecret = msg.secret
		m.totpQR = renderQRString(msg.key)
		m.state = stateConfirmTOTP
		m.initEnable2FAInput()
		m.err = ""
		return m, nil

	case confirmTOTPMsg:
		m.state = stateDashboard
		m.dashInput.Focus()
		m.dashInput.SetValue("")
		m.info = successStyle.Render("✓ Two-factor authentication enabled!")
		m.err = ""
		return m, refreshUserCmd(m.authSvc, m.sessionToken)

	case disableTOTPMsg:
		m.state = stateDashboard
		m.dashInput.Focus()
		m.dashInput.SetValue("")
		m.info = successStyle.Render("✓ Two-factor authentication disabled!")
		m.err = ""
		return m, refreshUserCmd(m.authSvc, m.sessionToken)

	case userRefreshedMsg:
		m.currentUser = msg.user
		if m.state == stateDashboard {
			m.info = renderUserDetails(m.currentUser, m.expiresAt)
		}
		return m, nil

	case errMsg:
		m.err = errorStyle.Render("✗ " + msg.Error())
		return m, nil
	}

	// State-specific update
	switch m.state {
	case stateLoading:
		// If we get here without a session restored, go to menu
		m.state = stateMenu
		return m, nil

	case stateMenu:
		return m.updateMenu(msg)

	case stateRegister, stateLogin, stateTOTPPrompt, stateConfirmTOTP, stateDisableTOTP:
		return m.updateInputs(msg)

	case stateDashboard:
		return m.updateDashboard(msg)
	}

	return m, nil
}

func (m Model) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if m.menuIdx > 0 {
				m.menuIdx--
			}
		case tea.KeyDown:
			if m.menuIdx < len(m.menuItems)-1 {
				m.menuIdx++
			}
		case tea.KeyEnter:
			return m.selectMenuItem()
		}

		// Number keys for quick select
		switch msg.String() {
		case "1":
			m.menuIdx = 0
			return m.selectMenuItem()
		case "2":
			m.menuIdx = 1
			return m.selectMenuItem()
		case "3":
			m.menuIdx = 2
			return m.selectMenuItem()
		case "4":
			m.menuIdx = 3
			return m.selectMenuItem()
		case "q":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) selectMenuItem() (tea.Model, tea.Cmd) {
	m.err = ""
	switch m.menuItems[m.menuIdx].label {
	case "register":
		m.state = stateRegister
		m.initRegisterInputs()
		m.info = ""
	case "login":
		m.state = stateLogin
		m.initLoginInputs()
		m.info = ""
	case "help":
		m.info = renderPreAuthHelp()
	case "exit":
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) updateInputs(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.state == stateTOTPPrompt || m.state == stateConfirmTOTP || m.state == stateDisableTOTP {
				m.state = stateDashboard
				m.dashInput.Focus()
				m.dashInput.SetValue("")
				m.err = ""
				m.info = ""
				if m.state == stateTOTPPrompt {
					// Go back to menu if cancelling TOTP during login
					m.state = stateMenu
				}
			} else {
				m.state = stateMenu
				m.err = ""
				m.info = ""
			}
			return m, nil

		case tea.KeyTab, tea.KeyDown:
			m.inputIdx++
			if m.inputIdx >= len(m.inputs) {
				m.inputIdx = 0
			}
			return m.focusInput(), nil

		case tea.KeyShiftTab, tea.KeyUp:
			m.inputIdx--
			if m.inputIdx < 0 {
				m.inputIdx = len(m.inputs) - 1
			}
			return m.focusInput(), nil

		case tea.KeyEnter:
			// If not on last input, move to next
			if m.inputIdx < len(m.inputs)-1 {
				m.inputIdx++
				return m.focusInput(), nil
			}
			// Submit
			return m.submitInputs()
		}
	}

	// Update the focused input
	if m.inputIdx < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.inputIdx], cmd = m.inputs[m.inputIdx].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) focusInput() Model {
	for i := range m.inputs {
		if i == m.inputIdx {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
	return m
}

func (m Model) submitInputs() (tea.Model, tea.Cmd) {
	switch m.inputMode {
	case "register":
		username := strings.TrimSpace(m.inputs[0].Value())
		password := m.inputs[1].Value()
		confirm := m.inputs[2].Value()

		if password != confirm {
			m.err = errorStyle.Render("✗ Passwords do not match")
			return m, nil
		}

		return m, registerCmd(m.authSvc, username, password)

	case "login":
		m.pendingUsername = strings.TrimSpace(m.inputs[0].Value())
		m.pendingPassword = m.inputs[1].Value()

		return m, loginCmd(m.authSvc, m.pendingUsername, m.pendingPassword, "")

	case "totp":
		code := strings.TrimSpace(m.inputs[0].Value())
		return m, totpLoginCmd(m.authSvc, m.pendingUsername, m.pendingPassword, code)

	case "enable2fa":
		code := strings.TrimSpace(m.inputs[0].Value())
		return m, confirmTOTPCmd(m.authSvc, m.currentUser.ID, m.totpSecret, code)

	case "disable2fa":
		code := strings.TrimSpace(m.inputs[0].Value())
		if m.currentUser.TOTPSecret == nil {
			m.err = errorStyle.Render("✗ 2FA is not enabled")
			return m, nil
		}
		return m, disableTOTPCmd(m.authSvc, m.currentUser.ID, *m.currentUser.TOTPSecret, code)
	}

	return m, nil
}

func (m Model) updateDashboard(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			cmd := strings.TrimSpace(m.dashInput.Value())
			m.dashInput.SetValue("")
			return m.handleDashboardCmd(cmd)
		}
	}

	var cmd tea.Cmd
	m.dashInput, cmd = m.dashInput.Update(msg)
	return m, cmd
}

func (m Model) handleDashboardCmd(cmd string) (tea.Model, tea.Cmd) {
	m.err = ""
	m.info = ""

	cmd = strings.ToLower(strings.TrimSpace(cmd))
	if cmd == "" {
		return m, nil
	}

	if cmd == "exit" || cmd == "quit" {
		m.quitting = true
		return m, tea.Quit
	}
	if cmd == "help" {
		m.info = renderPostAuthHelp()
		return m, nil
	}

	validCommands := map[string]bool{
		"whoami":      true,
		"enable-2fa":  true,
		"disable-2fa": true,
		"logout":      true,
	}

	if !validCommands[cmd] {
		m.err = errorStyle.Render("✗ Unknown command: " + cmd + ". Type 'help' for available commands.")
		return m, nil
	}

	return m, validateSessionCmd(m.authSvc, m.sessionToken, cmd)
}

func (m Model) executeDashboardAction(cmd string) (tea.Model, tea.Cmd) {
	m.err = ""
	m.info = ""
	switch cmd {
	case "whoami":
		return m, refreshUserCmd(m.authSvc, m.sessionToken)

	case "enable-2fa":
		if m.currentUser.TOTPEnabled {
			m.info = warningStyle.Render("⚠ Two-factor authentication is already enabled")
			return m, nil
		}
		return m, enableTOTPCmd(m.authSvc, m.currentUser.Username)

	case "disable-2fa":
		if !m.currentUser.TOTPEnabled {
			m.info = warningStyle.Render("⚠ Two-factor authentication is not enabled")
			return m, nil
		}
		m.state = stateDisableTOTP
		m.initDisable2FAInput()
		return m, nil

	case "logout":
		return m, logoutCmd(m.authSvc, m.sessionToken)
	}

	return m, nil
}

// View renders the current state.
func (m Model) View() string {
	if m.quitting {
		return mutedStyle.Render("\n  Goodbye! 👋\n\n")
	}

	switch m.state {
	case stateLoading:
		return "\n  Loading...\n"
	case stateMenu:
		return m.viewMenu()
	case stateRegister:
		return m.viewRegister()
	case stateLogin:
		return m.viewLogin()
	case stateTOTPPrompt:
		return m.viewTOTPPrompt()
	case stateDashboard:
		return m.viewDashboard()
	case stateEnableTOTP, stateConfirmTOTP:
		return m.viewEnableTOTP()
	case stateDisableTOTP:
		return m.viewDisableTOTP()
	}

	return ""
}

func renderPreAuthHelp() string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render("Available Commands") + "\n\n")
	b.WriteString("  " + promptStyle.Render("register") + "  — Create a new account\n")
	b.WriteString("  " + promptStyle.Render("login") + "     — Login with your credentials\n")
	b.WriteString("  " + promptStyle.Render("help") + "      — Show this help message\n")
	b.WriteString("  " + promptStyle.Render("exit") + "      — Quit the program\n")
	return b.String()
}

var IST = time.FixedZone("IST", 5*60*60+30*60)

func renderPostAuthHelp() string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render("Available Commands") + "\n\n")
	b.WriteString("  " + promptStyle.Render("whoami") + "       — Show your account details\n")
	b.WriteString("  " + promptStyle.Render("enable-2fa") + "   — Enable TOTP-based 2FA\n")
	b.WriteString("  " + promptStyle.Render("disable-2fa") + "  — Disable 2FA\n")
	b.WriteString("  " + promptStyle.Render("logout") + "       — End your session\n")
	b.WriteString("  " + promptStyle.Render("help") + "         — Show this help message\n")
	b.WriteString("  " + promptStyle.Render("exit") + "         — Quit the program\n")
	return b.String()
}

func renderUserDetails(u *user.User, expiresAt string) string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render("User Details") + "\n\n")
	b.WriteString(labelStyle.Render("  Username:") + "  " + valueStyle.Render(u.Username) + "\n")
	b.WriteString(labelStyle.Render("  Registered:") + "  " + valueStyle.Render(u.CreatedAt.In(IST).Format("2006-01-02 15:04:05 IST")) + "\n")

	mfaStatus := successStyle.Render("Enabled ✓")
	if !u.TOTPEnabled {
		mfaStatus = warningStyle.Render("Disabled")
	}
	b.WriteString(labelStyle.Render("  MFA Status:") + "  " + mfaStatus + "\n")
	b.WriteString(labelStyle.Render("  Session Expires:") + "  " + valueStyle.Render(expiresAt) + "\n")

	if u.LastLoginAt != nil {
		b.WriteString(labelStyle.Render("  Last Login:") + "  " + valueStyle.Render(u.LastLoginAt.In(IST).Format("2006-01-02 15:04:05 IST")) + "\n")
	} else {
		b.WriteString(labelStyle.Render("  Last Login:") + "  " + mutedStyle.Render("First login") + "\n")
	}

	return b.String()
}
