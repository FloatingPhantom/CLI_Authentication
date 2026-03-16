package tui

import (
	"fmt"
	"strings"
)

func (m Model) viewMenu() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(headerStyle.Render(" 🔐 CLI Auth ") + "\n\n")

	for i, item := range m.menuItems {
		cursor := "  "
		style := menuItemStyle
		if i == m.menuIdx {
			cursor = "▸ "
			style = menuSelectedStyle
		}
		num := fmt.Sprintf("%d. ", i+1)
		b.WriteString(style.Render(cursor+num+item.label) + mutedStyle.Render("  "+item.desc) + "\n")
	}

	b.WriteString("\n")
	if m.info != "" {
		b.WriteString("  " + m.info + "\n\n")
	}
	if m.err != "" {
		b.WriteString("  " + m.err + "\n\n")
	}

	b.WriteString(helpStyle.Render("  ↑/↓ navigate • enter select • 1-4 quick select • q quit"))
	b.WriteString("\n")

	return b.String()
}

func (m Model) viewRegister() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(headerStyle.Render(" 📝 Register ") + "\n\n")

	labels := []string{"Username", "Password", "Confirm Password"}
	for i, label := range labels {
		active := " "
		if i == m.inputIdx {
			active = "▸"
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", active, subtitleStyle.Render(label)))
		b.WriteString("    " + m.inputs[i].View() + "\n\n")
	}

	if m.err != "" {
		b.WriteString("  " + m.err + "\n\n")
	}

	b.WriteString(helpStyle.Render("  tab next field • enter submit • esc back"))
	b.WriteString("\n")

	return b.String()
}

func (m Model) viewLogin() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(headerStyle.Render(" 🔑 Login ") + "\n\n")

	labels := []string{"Username", "Password"}
	for i, label := range labels {
		active := " "
		if i == m.inputIdx {
			active = "▸"
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", active, subtitleStyle.Render(label)))
		b.WriteString("    " + m.inputs[i].View() + "\n\n")
	}

	if m.err != "" {
		b.WriteString("  " + m.err + "\n\n")
	}

	b.WriteString(helpStyle.Render("  tab next field • enter submit • esc back"))
	b.WriteString("\n")

	return b.String()
}

func (m Model) viewTOTPPrompt() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(headerStyle.Render(" 🔒 Two-Factor Authentication ") + "\n\n")
	b.WriteString("  " + mutedStyle.Render("Enter the 6-digit code from your authenticator app") + "\n\n")

	b.WriteString("  ▸ " + subtitleStyle.Render("TOTP Code") + "\n")
	b.WriteString("    " + m.inputs[0].View() + "\n\n")

	if m.err != "" {
		b.WriteString("  " + m.err + "\n\n")
	}

	b.WriteString(helpStyle.Render("  enter submit • esc cancel"))
	b.WriteString("\n")

	return b.String()
}

func (m Model) viewDashboard() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(headerStyle.Render(fmt.Sprintf(" 🏠 Dashboard — %s ", m.currentUser.Username)) + "\n\n")

	// Show user info card
	b.WriteString(renderUserDetails(m.currentUser, m.expiresAt))
	b.WriteString("\n")

	if m.info != "" {
		b.WriteString("  " + m.info + "\n\n")
	}
	if m.err != "" {
		b.WriteString("  " + m.err + "\n\n")
	}

	// Command prompt
	b.WriteString("  " + m.dashInput.View() + "\n")
	b.WriteString(helpStyle.Render("  tab complete • up/down history • commands: whoami • enable-2fa • disable-2fa • logout • help • exit"))
	b.WriteString("\n")

	return b.String()
}

func (m Model) viewEnableTOTP() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(headerStyle.Render(" 🔐 Enable Two-Factor Authentication ") + "\n\n")

	if m.totpQR != "" {
		b.WriteString("  " + subtitleStyle.Render("Scan this QR code with your authenticator app:") + "\n\n")
		// Indent QR code
		for _, line := range strings.Split(m.totpQR, "\n") {
			b.WriteString("  " + line + "\n")
		}
		b.WriteString("\n")
		b.WriteString("  " + mutedStyle.Render("Or manually enter this secret:") + "\n")
		b.WriteString("  " + successStyle.Render(m.totpSecret) + "\n\n")
	}

	b.WriteString("  ▸ " + subtitleStyle.Render("Verification Code") + "\n")
	b.WriteString("    " + m.inputs[0].View() + "\n\n")

	if m.err != "" {
		b.WriteString("  " + m.err + "\n\n")
	}

	b.WriteString(helpStyle.Render("  enter verify • esc cancel"))
	b.WriteString("\n")

	return b.String()
}

func (m Model) viewDisableTOTP() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(headerStyle.Render(" 🔓 Disable Two-Factor Authentication ") + "\n\n")
	b.WriteString("  " + mutedStyle.Render("Enter your current TOTP code to confirm disabling 2FA") + "\n\n")

	b.WriteString("  ▸ " + subtitleStyle.Render("TOTP Code") + "\n")
	b.WriteString("    " + m.inputs[0].View() + "\n\n")

	if m.err != "" {
		b.WriteString("  " + m.err + "\n\n")
	}

	b.WriteString(helpStyle.Render("  enter confirm • esc cancel"))
	b.WriteString("\n")

	return b.String()
}
