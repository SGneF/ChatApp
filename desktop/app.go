package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// App struct
type App struct {
	ctx        context.Context
	backendURL string
	client     *http.Client
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateProfileRequest struct {
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Signature string `json:"signature"`
}

type ApplyFriendRequest struct {
	ToUserID uint64 `json:"to_user_id"`
	Remark   string `json:"remark"`
}

type HandleFriendRequest struct {
	RequestID uint64 `json:"request_id"`
}

type FriendRequestResponse struct {
	ID           uint64 `json:"id"`
	FromUserID   uint64 `json:"from_user_id"`
	FromUsername string `json:"from_username"`
	FromNickname string `json:"from_nickname"`
	FromAvatar   string `json:"from_avatar"`
	Remark       string `json:"remark"`
	Status       string `json:"status"`
	CreateTime   string `json:"create_time"`
}

type FriendResponse struct {
	ID        uint64 `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Signature string `json:"signature"`
	Remark    string `json:"remark"`
}

type SearchUserResponse struct {
	ID        uint64 `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Signature string `json:"signature"`
}
type UserResponse struct {
	ID         uint64 `json:"id"`
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	Signature  string `json:"signature"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type BackendStatus struct {
	OK      bool   `json:"ok"`
	URL     string `json:"url"`
	Message string `json:"message"`
}

type apiEnvelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	backendURL := os.Getenv("LIGHTCHAT_API_BASE_URL")
	if backendURL == "" {
		backendURL = "http://127.0.0.1:8080"
	}

	normalized, err := normalizeBackendURL(backendURL)
	if err != nil {
		normalized = "http://127.0.0.1:8080"
	}

	return &App{
		backendURL: normalized,
		client:     &http.Client{Timeout: 8 * time.Second},
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s", name)
}

func (a *App) BackendURL() string {
	return a.backendURL
}

func (a *App) SetBackendURL(value string) (string, error) {
	normalized, err := normalizeBackendURL(value)
	if err != nil {
		return "", err
	}
	a.backendURL = normalized
	return a.backendURL, nil
}

func (a *App) PingBackend() BackendStatus {
	var envelope apiEnvelope
	if err := a.call(http.MethodGet, "/ping", "", nil, &envelope); err != nil {
		return BackendStatus{OK: false, URL: a.backendURL, Message: err.Error()}
	}

	message := envelope.Msg
	if message == "" {
		message = "pong"
	}
	return BackendStatus{OK: true, URL: a.backendURL, Message: message}
}

func (a *App) Register(req RegisterRequest) (*UserResponse, error) {
	var user UserResponse
	if err := a.call(http.MethodPost, "/api/user/register", "", req, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (a *App) Login(req LoginRequest) (*LoginResponse, error) {
	var resp LoginResponse
	if err := a.call(http.MethodPost, "/api/user/login", "", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (a *App) GetUserInfo(token string) (*UserResponse, error) {
	var user UserResponse
	if err := a.call(http.MethodGet, "/api/user/info", token, nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (a *App) UpdateProfile(token string, req UpdateProfileRequest) (*UserResponse, error) {
	var user UserResponse
	if err := a.call(http.MethodPost, "/api/user/profile", token, req, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (a *App) SearchUsers(token string, keyword string) ([]SearchUserResponse, error) {
	var list []SearchUserResponse
	endpoint := "/api/user/search"
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		endpoint += "?keyword=" + url.QueryEscape(keyword)
	}
	if err := a.call(http.MethodGet, endpoint, token, nil, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (a *App) ApplyFriend(token string, req ApplyFriendRequest) error {
	return a.call(http.MethodPost, "/api/friend/apply", token, req, nil)
}

func (a *App) ListFriendRequests(token string) ([]FriendRequestResponse, error) {
	var list []FriendRequestResponse
	if err := a.call(http.MethodGet, "/api/friend/requests", token, nil, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (a *App) AcceptFriendRequest(token string, req HandleFriendRequest) error {
	return a.call(http.MethodPost, "/api/friend/accept", token, req, nil)
}

func (a *App) RejectFriendRequest(token string, req HandleFriendRequest) error {
	return a.call(http.MethodPost, "/api/friend/reject", token, req, nil)
}

func (a *App) ListFriends(token string) ([]FriendResponse, error) {
	var list []FriendResponse
	if err := a.call(http.MethodGet, "/api/friend/list", token, nil, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (a *App) DeleteFriend(token string, friendID uint64) error {
	return a.call(http.MethodDelete, fmt.Sprintf("/api/friend/%d", friendID), token, nil, nil)
}
func (a *App) call(method string, path string, token string, payload interface{}, out interface{}) error {
	var body io.Reader
	if payload != nil {
		buf, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(buf)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, a.backendURL+path, body)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求后端服务失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var envelope apiEnvelope
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return fmt.Errorf("接口响应格式错误: %s", strings.TrimSpace(string(respBody)))
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		if envelope.Msg != "" {
			return errors.New(envelope.Msg)
		}
		return fmt.Errorf("接口请求失败: HTTP %d", resp.StatusCode)
	}
	if envelope.Code != 1 {
		if envelope.Msg != "" {
			return errors.New(envelope.Msg)
		}
		return errors.New("接口返回失败")
	}

	if out == nil {
		return nil
	}
	if target, ok := out.(*apiEnvelope); ok {
		*target = envelope
		return nil
	}
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return nil
	}
	return json.Unmarshal(envelope.Data, out)
}

func normalizeBackendURL(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", errors.New("后端地址不能为空")
	}
	if !strings.Contains(trimmed, "://") {
		trimmed = "http://" + trimmed
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("后端地址必须使用 http 或 https")
	}
	if parsed.Host == "" {
		return "", errors.New("后端地址缺少主机")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return strings.TrimRight(parsed.String(), "/"), nil
}
