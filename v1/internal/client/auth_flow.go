package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"treehole/internal/config"

	"github.com/pquerna/otp/totp"
)

type AuthChallengeKind string

const (
	AuthChallengeNone AuthChallengeKind = ""
	AuthChallengeSMS  AuthChallengeKind = "sms"
	AuthChallengeOTP  AuthChallengeKind = "otp"
)

type AuthBootstrapResult struct {
	Status          SessionStatus
	Challenge       AuthChallengeKind
	ChallengeReason string
	LoginAttempted  bool
}

func DetectAuthChallenge(message string) AuthChallengeKind {
	switch {
	case strings.Contains(message, "短信验证"):
		return AuthChallengeSMS
	case strings.Contains(message, "令牌验证"):
		return AuthChallengeOTP
	default:
		return AuthChallengeNone
	}
}

func (c *Client) BootstrapSession(cfg *config.Config) AuthBootstrapResult {
	result := c.finalizeAuthStatus()
	if result.Status.CanReadOnline {
		return result
	}
	if cfg == nil || !cfg.HasPasswordLogin() {
		return result
	}

	result.LoginAttempted = true
	oauthResult, err := c.OAuthLogin(cfg.Username, cfg.Password)
	if err != nil {
		result.Status.FailureKind = ClassifySessionError(err)
		result.Status.Message = err.Error()
		result.Challenge = AuthChallengeNone
		result.ChallengeReason = ""
		return result
	}

	token, ok := oauthResult["token"].(string)
	if !ok || token == "" {
		result.Status.FailureKind = SessionFailureLogin
		result.Status.Message = "OAuth 登录未返回 token"
		result.Challenge = AuthChallengeNone
		result.ChallengeReason = ""
		return result
	}

	if err := c.SSOLogin(token); err != nil {
		result.Status.FailureKind = ClassifySessionError(err)
		result.Status.Message = err.Error()
		result.Challenge = AuthChallengeNone
		result.ChallengeReason = ""
		return result
	}

	result = c.finalizeAuthStatus()
	result.LoginAttempted = true
	if result.Status.CanReadOnline || result.Challenge != AuthChallengeOTP || cfg == nil || !cfg.HasTOTPSecret() {
		return result
	}

	code, err := totp.GenerateCode(cfg.SecretKey, time.Now())
	if err != nil {
		result.Status.FailureKind = SessionFailureLogin
		result.Status.Message = err.Error()
		result.Challenge = AuthChallengeOTP
		result.ChallengeReason = result.Status.Message
		return result
	}

	submit := c.ContinueAuthChallenge(AuthChallengeOTP, code)
	submit.LoginAttempted = true
	return submit
}

func (c *Client) ContinueAuthChallenge(kind AuthChallengeKind, code string) AuthBootstrapResult {
	var err error
	switch kind {
	case AuthChallengeSMS:
		err = c.SubmitSMSCode(code)
	case AuthChallengeOTP:
		err = c.SubmitOTPCode(code)
	default:
		err = fmt.Errorf("unsupported auth challenge: %s", kind)
	}
	if err != nil {
		return AuthBootstrapResult{
			Status: SessionStatus{
				HasSession:  c.GetAuthorization() != "",
				FailureKind: ClassifySessionError(err),
				Message:     err.Error(),
			},
			Challenge:       kind,
			ChallengeReason: err.Error(),
		}
	}
	return c.finalizeAuthStatus()
}

func (c *Client) SendSMSCode() error {
	resp, err := c.SendMessage()
	if err != nil {
		return err
	}
	return parseAuthAPIResponse(resp, "发送短信验证码")
}

func (c *Client) SubmitSMSCode(code string) error {
	resp, err := c.LoginByMessage(code)
	if err != nil {
		return err
	}
	return parseAuthAPIResponse(resp, "短信验证")
}

func (c *Client) SubmitOTPCode(code string) error {
	resp, err := c.LoginByToken(code)
	if err != nil {
		return err
	}
	return parseAuthAPIResponse(resp, "令牌验证")
}

func (c *Client) finalizeAuthStatus() AuthBootstrapResult {
	status := c.ProbeSession()
	challenge := DetectAuthChallenge(status.Message)
	if status.CanReadOnline {
		_ = c.SaveCookies()
	}
	return AuthBootstrapResult{
		Status:          status,
		Challenge:       challenge,
		ChallengeReason: status.Message,
	}
}

func parseAuthAPIResponse(resp *http.Response, action string) error {
	if resp == nil {
		return fmt.Errorf("%s失败: 空响应", action)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s失败: HTTP %d", action, resp.StatusCode)
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("%s失败: %w", action, err)
	}

	if success, ok := payload["success"].(bool); ok && success {
		return nil
	}
	if code, ok := payload["code"].(float64); ok && int(code) == 20000 {
		return nil
	}
	if message, ok := payload["message"].(string); ok && message != "" {
		return fmt.Errorf("%s失败: %s", action, message)
	}
	return fmt.Errorf("%s失败", action)
}
