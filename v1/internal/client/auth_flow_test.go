package client

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
)

func TestDetectAuthChallenge(t *testing.T) {
	tests := []struct {
		message string
		want    AuthChallengeKind
	}{
		{"需要短信验证", AuthChallengeSMS},
		{"请完成令牌验证", AuthChallengeOTP},
		{"登录态不可用", AuthChallengeNone},
	}

	for _, tt := range tests {
		if got := DetectAuthChallenge(tt.message); got != tt.want {
			t.Fatalf("DetectAuthChallenge(%q) = %q, want %q", tt.message, got, tt.want)
		}
	}
}

func TestParseAuthAPIResponse(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"success":true}`)),
	}
	if err := parseAuthAPIResponse(resp, "短信验证"); err != nil {
		t.Fatalf("parseAuthAPIResponse success: %v", err)
	}

	resp = &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"success":false,"message":"验证码错误"}`)),
	}
	if err := parseAuthAPIResponse(resp, "短信验证"); err == nil || !strings.Contains(err.Error(), "验证码错误") {
		t.Fatalf("parseAuthAPIResponse error = %v, want message", err)
	}
}

type authRoundTripper struct {
	sendCalls   int
	tokenCalls  int
	messageCalls int
}

func (rt *authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	body := `{"success":true}`
	switch {
	case strings.Contains(req.URL.String(), string(SEND_MESSAGE)):
		rt.sendCalls++
	case strings.Contains(req.URL.String(), string(LOGIN_BY_TOKEN)):
		rt.tokenCalls++
	case strings.Contains(req.URL.String(), string(LOGIN_BY_MESSAGE)):
		rt.messageCalls++
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}

func TestAuthSubmitHelpers(t *testing.T) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New: %v", err)
	}
	rt := &authRoundTripper{}
	c := &Client{
		httpClient: &http.Client{Jar: jar, Transport: rt},
		deviceUUID: "test-uuid",
	}
	jar.SetCookies(&url.URL{Scheme: "https", Host: "treehole.pku.edu.cn"}, []*http.Cookie{{Name: "XSRF-TOKEN", Value: "xsrf-token"}})

	if err := c.SendSMSCode(); err != nil {
		t.Fatalf("SendSMSCode: %v", err)
	}
	if err := c.SubmitOTPCode("123456"); err != nil {
		t.Fatalf("SubmitOTPCode: %v", err)
	}
	if err := c.SubmitSMSCode("654321"); err != nil {
		t.Fatalf("SubmitSMSCode: %v", err)
	}
	if rt.sendCalls != 1 || rt.tokenCalls != 1 || rt.messageCalls != 1 {
		t.Fatalf("unexpected auth helper call counts: %+v", rt)
	}
}
