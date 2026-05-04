package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	slackAPI       = "https://slack.com/api"
	httpTimeout    = 30 * time.Second
	maxRetry429    = 3
)

type slackResp struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
	// Echoed back on errors
	ResponseMetadata struct {
		Messages []string `json:"messages"`
	} `json:"response_metadata"`
}

type apiClient struct {
	token string
	http  *http.Client
}

func newClient(token string) *apiClient {
	return &apiClient{token: token, http: &http.Client{Timeout: httpTimeout}}
}

// call invokes a Slack API method with form-encoded params; out receives parsed JSON.
func (c *apiClient) call(ctx context.Context, method string, params url.Values, out any) error {
	rateWait(method)
	for attempt := 0; attempt <= maxRetry429; attempt++ {
		body := strings.NewReader(params.Encode())
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, slackAPI+"/"+method, body)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == 429 {
			wait := parseRetryAfter(resp.Header.Get("Retry-After"), 1)
			time.Sleep(wait)
			continue
		}
		if out != nil {
			if err := json.Unmarshal(raw, out); err != nil {
				return fmt.Errorf("decode %s: %w", method, err)
			}
		}
		var base slackResp
		_ = json.Unmarshal(raw, &base)
		if !base.OK {
			return fmt.Errorf("%s: %s", method, base.Error)
		}
		return nil
	}
	return fmt.Errorf("%s: rate-limited after %d retries", method, maxRetry429)
}

func parseRetryAfter(h string, fallback int) time.Duration {
	if n, err := strconv.Atoi(strings.TrimSpace(h)); err == nil && n > 0 {
		return time.Duration(n) * time.Second
	}
	return time.Duration(fallback) * time.Second
}

// --- channels ---

type channel struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	IsArchived bool   `json:"is_archived"`
	IsMember   bool   `json:"is_member"`
	IsChannel  bool   `json:"is_channel"`
	IsGroup    bool   `json:"is_group"`
	IsIM       bool   `json:"is_im"`
	IsMpIM     bool   `json:"is_mpim"`
}

// findChannelByName resolves a #channel name to a channel ID. Uses the on-disk
// channels cache first; falls back to users.conversations.
func (c *apiClient) findChannelByName(ctx context.Context, name string) (string, error) {
	name = strings.TrimPrefix(strings.ToLower(name), "#")
	if id := channelCacheGet(name); id != "" {
		return id, nil
	}
	cursor := ""
	for {
		params := url.Values{}
		params.Set("types", "public_channel")
		params.Set("exclude_archived", "true")
		params.Set("limit", "200")
		if cursor != "" {
			params.Set("cursor", cursor)
		}
		var r struct {
			slackResp
			Channels         []channel `json:"channels"`
			ResponseMetadata struct {
				NextCursor string `json:"next_cursor"`
			} `json:"response_metadata"`
		}
		if err := c.call(ctx, "users.conversations", params, &r); err != nil {
			return "", err
		}
		for _, ch := range r.Channels {
			channelCacheSet(ch.Name, ch.ID)
			if ch.Name == name {
				return ch.ID, nil
			}
		}
		if r.ResponseMetadata.NextCursor == "" {
			break
		}
		cursor = r.ResponseMetadata.NextCursor
	}
	return "", fmt.Errorf("channel #%s not found (must be a public channel you've joined)", name)
}

// --- DMs ---

func (c *apiClient) openDM(ctx context.Context, userID string) (string, error) {
	params := url.Values{}
	params.Set("users", userID)
	var r struct {
		slackResp
		Channel struct {
			ID string `json:"id"`
		} `json:"channel"`
	}
	if err := c.call(ctx, "conversations.open", params, &r); err != nil {
		return "", err
	}
	return r.Channel.ID, nil
}

func (c *apiClient) openGroupDM(ctx context.Context, userIDs []string) (string, error) {
	params := url.Values{}
	params.Set("users", strings.Join(userIDs, ","))
	var r struct {
		slackResp
		Channel struct {
			ID string `json:"id"`
		} `json:"channel"`
	}
	if err := c.call(ctx, "conversations.open", params, &r); err != nil {
		return "", err
	}
	return r.Channel.ID, nil
}

// --- users ---

type slackUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	RealName string `json:"real_name"`
	Profile  struct {
		DisplayName string `json:"display_name"`
		RealName    string `json:"real_name"`
		Email       string `json:"email"`
	} `json:"profile"`
	Deleted bool `json:"deleted"`
	IsBot   bool `json:"is_bot"`
}

func (u slackUser) bestName() string {
	if u.Profile.DisplayName != "" {
		return u.Profile.DisplayName
	}
	if u.Profile.RealName != "" {
		return u.Profile.RealName
	}
	if u.RealName != "" {
		return u.RealName
	}
	return u.Name
}

func (c *apiClient) lookupByEmail(ctx context.Context, email string) (*slackUser, error) {
	params := url.Values{}
	params.Set("email", email)
	var r struct {
		slackResp
		User slackUser `json:"user"`
	}
	if err := c.call(ctx, "users.lookupByEmail", params, &r); err != nil {
		return nil, err
	}
	return &r.User, nil
}

func (c *apiClient) usersInfo(ctx context.Context, id string) (*slackUser, error) {
	params := url.Values{}
	params.Set("user", id)
	var r struct {
		slackResp
		User slackUser `json:"user"`
	}
	if err := c.call(ctx, "users.info", params, &r); err != nil {
		return nil, err
	}
	return &r.User, nil
}

// --- messages ---

type slackMessage struct {
	Type        string  `json:"type"`
	User        string  `json:"user"`
	Username    string  `json:"username"`
	Text        string  `json:"text"`
	Ts          string  `json:"ts"`
	ThreadTs    string  `json:"thread_ts"`
	ReplyCount  int     `json:"reply_count"`
	Reactions   []reactSummary `json:"reactions"`
	BotID       string  `json:"bot_id"`
	Files       []slackFile    `json:"files"`
}

type reactSummary struct {
	Name  string   `json:"name"`
	Count int      `json:"count"`
	Users []string `json:"users"`
}

type slackFile struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Title string `json:"title"`
	URL   string `json:"url_private"`
}

func (c *apiClient) postMessage(ctx context.Context, channelID, text, threadTs string, broadcast bool) (string, error) {
	params := url.Values{}
	params.Set("channel", channelID)
	params.Set("text", text)
	if threadTs != "" {
		params.Set("thread_ts", threadTs)
		if broadcast {
			params.Set("reply_broadcast", "true")
		}
	}
	var r struct {
		slackResp
		Ts string `json:"ts"`
	}
	if err := c.call(ctx, "chat.postMessage", params, &r); err != nil {
		return "", err
	}
	return r.Ts, nil
}

func (c *apiClient) addReaction(ctx context.Context, channelID, ts, name string) error {
	params := url.Values{}
	params.Set("channel", channelID)
	params.Set("timestamp", ts)
	params.Set("name", strings.Trim(name, ":"))
	return c.call(ctx, "reactions.add", params, nil)
}

func (c *apiClient) history(ctx context.Context, channelID string, limit int) ([]slackMessage, error) {
	params := url.Values{}
	params.Set("channel", channelID)
	params.Set("limit", strconv.Itoa(limit))
	var r struct {
		slackResp
		Messages []slackMessage `json:"messages"`
	}
	if err := c.call(ctx, "conversations.history", params, &r); err != nil {
		return nil, err
	}
	return r.Messages, nil
}

func (c *apiClient) replies(ctx context.Context, channelID, threadTs string, limit int) ([]slackMessage, error) {
	params := url.Values{}
	params.Set("channel", channelID)
	params.Set("ts", threadTs)
	params.Set("limit", strconv.Itoa(limit))
	var r struct {
		slackResp
		Messages []slackMessage `json:"messages"`
	}
	if err := c.call(ctx, "conversations.replies", params, &r); err != nil {
		return nil, err
	}
	return r.Messages, nil
}

// --- search ---

type searchHit struct {
	Type      string  `json:"type"`
	Channel   struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"channel"`
	User      string  `json:"user"`
	Username  string  `json:"username"`
	Text      string  `json:"text"`
	Ts        string  `json:"ts"`
	Permalink string  `json:"permalink"`
}

type searchResult struct {
	Total    int         `json:"total"`
	Matches  []searchHit `json:"matches"`
}

func (c *apiClient) searchMessages(ctx context.Context, query string, count int, sort string) (*searchResult, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("count", strconv.Itoa(count))
	if sort != "" {
		params.Set("sort", sort)
	}
	var r struct {
		slackResp
		Messages searchResult `json:"messages"`
	}
	if err := c.call(ctx, "search.messages", params, &r); err != nil {
		return nil, err
	}
	return &r.Messages, nil
}

// --- file upload (v2 endpoint) ---

func (c *apiClient) uploadFile(ctx context.Context, channelID, path, message, title string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	filename := filepath.Base(path)
	if title == "" {
		title = filename
	}

	// 1. Get upload URL
	params := url.Values{}
	params.Set("filename", filename)
	params.Set("length", strconv.FormatInt(info.Size(), 10))
	var step1 struct {
		slackResp
		UploadURL string `json:"upload_url"`
		FileID    string `json:"file_id"`
	}
	if err := c.call(ctx, "files.getUploadURLExternal", params, &step1); err != nil {
		return err
	}

	// 2. POST file bytes to upload_url
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, step1.UploadURL, f)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("upload returned HTTP %d", resp.StatusCode)
	}

	// 3. Complete upload, share to channel
	files := []map[string]string{{"id": step1.FileID, "title": title}}
	filesJSON, _ := json.Marshal(files)
	params2 := url.Values{}
	params2.Set("files", string(filesJSON))
	params2.Set("channel_id", channelID)
	if message != "" {
		params2.Set("initial_comment", message)
	}
	return c.call(ctx, "files.completeUploadExternal", params2, nil)
}

// multipartBody is used by upload paths that need form-encoded multipart (currently unused
// since v2 upload uses raw PUT, but kept for future legacy fallback).
func multipartBody(fields map[string]string, fileField, filePath string) (io.Reader, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			return nil, "", err
		}
	}
	if filePath != "" {
		f, err := os.Open(filePath)
		if err != nil {
			return nil, "", err
		}
		defer f.Close()
		fw, err := w.CreateFormFile(fileField, filepath.Base(filePath))
		if err != nil {
			return nil, "", err
		}
		if _, err := io.Copy(fw, f); err != nil {
			return nil, "", err
		}
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return &buf, w.FormDataContentType(), nil
}

var _ = errors.New // keep errors imported
