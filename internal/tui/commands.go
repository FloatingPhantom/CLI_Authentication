package tui

import (
	"context"
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli-auth/internal/auth"
	"github.com/cli-auth/internal/user"
	"github.com/pquerna/otp"
)

// --- Messages ---

// registerMsg is sent after a successful registration.
type registerMsg struct{ user *user.User }

// loginMsg is sent after a successful login.
type loginMsg struct{ result *auth.LoginResult }

type dashActionMsg struct{ action string }

type sessionExpiredMsg struct{}

// totpRequiredMsg is sent when login requires TOTP.
type totpRequiredMsg struct {
	userID   int
	username string
}

// logoutMsg is sent after logout.
type logoutMsg struct{}

// sessionRestoredMsg is sent when a session is restored from file.
type sessionRestoredMsg struct {
	user    *user.User
	token   string
	expires string
}

// enableTOTPMsg carries the TOTP enrollment data.
type enableTOTPMsg struct {
	key    *otp.Key
	secret string
}

// confirmTOTPMsg is sent after 2FA is confirmed.
type confirmTOTPMsg struct{}

// disableTOTPMsg is sent after 2FA is disabled.
type disableTOTPMsg struct{}

// userRefreshedMsg is sent after refreshing user details.
type userRefreshedMsg struct{ user *user.User }

// noSessionMsg is sent when no valid session exists.
type noSessionMsg struct{}

// errMsg is sent on any error.
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

// --- Commands ---

func registerCmd(svc *auth.Service, username, password string) tea.Cmd {
	return func() tea.Msg {
		u, err := svc.Register(context.Background(), username, password)
		if err != nil {
			return errMsg{err}
		}
		return registerMsg{user: u}
	}
}

func loginCmd(svc *auth.Service, username, password, totpCode string) tea.Cmd {
	return func() tea.Msg {
		result, err := svc.Login(context.Background(), username, password, totpCode)
		if err != nil {
			if e, ok := err.(*auth.ErrTOTPRequired); ok {
				return totpRequiredMsg{userID: e.UserID, username: e.Username}
			}
			return errMsg{err}
		}
		return loginMsg{result: result}
	}
}

func totpLoginCmd(svc *auth.Service, username, password, code string) tea.Cmd {
	return func() tea.Msg {
		result, err := svc.Login(context.Background(), username, password, code)
		if err != nil {
			return errMsg{err}
		}
		return loginMsg{result: result}
	}
}

func logoutCmd(svc *auth.Service, token string) tea.Cmd {
	return func() tea.Msg {
		if err := svc.Logout(context.Background(), token); err != nil {
			return errMsg{err}
		}
		return logoutMsg{}
	}
}

func restoreSessionCmd(svc *auth.Service, loadToken func() (string, error), timeout string) tea.Cmd {
	return func() tea.Msg {
		token, err := loadToken()
		if err != nil {
			return noSessionMsg{} // No session file — go to menu
		}
		u, sess, err := svc.ValidateSession(context.Background(), token)
		if err != nil {
			return noSessionMsg{} // Expired session — go to menu
		}
		return sessionRestoredMsg{
			user:    u,
			token:   token,
			expires: sess.ExpiresAt.In(IST).Format("2006-01-02 15:04:05 IST"),
		}
	}
}

func enableTOTPCmd(svc *auth.Service, username string) tea.Cmd {
	return func() tea.Msg {
		enrollment, err := svc.EnableTOTP(username)
		if err != nil {
			return errMsg{err}
		}
		// We need to extract the *otp.Key from the enrollment
		key, ok := enrollment.Key.(*otp.Key)
		if !ok {
			return errMsg{fmt.Errorf("unexpected key type")}
		}
		return enableTOTPMsg{key: key, secret: enrollment.Secret}
	}
}

// renderQRString renders the QR code into a string.
func renderQRString(key *otp.Key) string {
	pr, pw := io.Pipe()
	go func() {
		auth.RenderQR(pw, key)
		pw.Close()
	}()
	data, _ := io.ReadAll(pr)
	return string(data)
}

func confirmTOTPCmd(svc *auth.Service, userID int, secret, code string) tea.Cmd {
	return func() tea.Msg {
		if err := svc.ConfirmTOTP(context.Background(), userID, secret, code); err != nil {
			return errMsg{err}
		}
		return confirmTOTPMsg{}
	}
}

func disableTOTPCmd(svc *auth.Service, userID int, secret, code string) tea.Cmd {
	return func() tea.Msg {
		if err := svc.DisableTOTP(context.Background(), userID, secret, code); err != nil {
			return errMsg{err}
		}
		return disableTOTPMsg{}
	}
}

func refreshUserCmd(svc *auth.Service, token string) tea.Cmd {
	return func() tea.Msg {
		u, _, err := svc.ValidateSession(context.Background(), token)
		if err != nil {
			return sessionExpiredMsg{}
		}
		return userRefreshedMsg{user: u}
	}
}

// validateSessionCmd checks if the session is still valid before running a dashboard action.
func validateSessionCmd(svc *auth.Service, token string, action string) tea.Cmd {
	return func() tea.Msg {
		_, _, err := svc.ValidateSession(context.Background(), token)
		if err != nil {
			return sessionExpiredMsg{}
		}
		return dashActionMsg{action: action}
	}
}
