package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	channelCacheTTL = 24 * time.Hour
	contactsTTL     = 1 * time.Hour
	userCacheTTL    = 1 * time.Hour
)

func cacheDir() string {
	if d := os.Getenv("XDG_CACHE_HOME"); d != "" {
		return filepath.Join(d, "director-slack")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "director-slack")
}

func cachePath(name string) string {
	return filepath.Join(cacheDir(), name)
}

func ensureCacheDir() error {
	return os.MkdirAll(cacheDir(), 0o755)
}

func readCache(name string, out any) bool {
	data, err := os.ReadFile(cachePath(name))
	if err != nil {
		return false
	}
	return json.Unmarshal(data, out) == nil
}

func writeCache(name string, in any) error {
	if err := ensureCacheDir(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cachePath(name), data, 0o644)
}

// --- channel cache (name → ID) ---

type channelCache struct {
	mu      sync.Mutex
	loaded  bool
	Updated int64             `json:"updated_at"`
	Map     map[string]string `json:"channels"`
}

var chCache channelCache

func (c *channelCache) load() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.loaded {
		return
	}
	c.Map = map[string]string{}
	var buf channelCache
	if readCache("channels.json", &buf) {
		if time.Since(time.Unix(buf.Updated, 0)) < channelCacheTTL {
			c.Map = buf.Map
		}
	}
	c.loaded = true
}

func channelCacheGet(name string) string {
	chCache.load()
	chCache.mu.Lock()
	defer chCache.mu.Unlock()
	return chCache.Map[strings.ToLower(name)]
}

func channelCacheSet(name, id string) {
	chCache.load()
	chCache.mu.Lock()
	defer chCache.mu.Unlock()
	chCache.Map[strings.ToLower(name)] = id
	chCache.Updated = time.Now().Unix()
	_ = writeCache("channels.json", chCache)
}

func channelCacheClear() error {
	chCache.mu.Lock()
	defer chCache.mu.Unlock()
	chCache.Map = map[string]string{}
	chCache.Updated = 0
	chCache.loaded = true
	return os.Remove(cachePath("channels.json"))
}

// --- recent contacts cache (name/handle → user_id) ---

type contactEntry struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	LastSeen int64  `json:"last_seen"`
}

type contactsCache struct {
	mu       sync.Mutex
	loaded   bool
	Updated  int64                   `json:"updated_at"`
	ByName   map[string]contactEntry `json:"by_name"`
	ByHandle map[string]contactEntry `json:"by_handle"`
}

var contacts contactsCache

func (c *contactsCache) load() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.loaded {
		return
	}
	c.ByName = map[string]contactEntry{}
	c.ByHandle = map[string]contactEntry{}
	var buf contactsCache
	if readCache("recent_contacts.json", &buf) {
		if time.Since(time.Unix(buf.Updated, 0)) < contactsTTL {
			if buf.ByName != nil {
				c.ByName = buf.ByName
			}
			if buf.ByHandle != nil {
				c.ByHandle = buf.ByHandle
			}
		}
	}
	c.loaded = true
}

func recentContactsLookup(name string) string {
	contacts.load()
	contacts.mu.Lock()
	defer contacts.mu.Unlock()
	if e, ok := contacts.ByName[strings.ToLower(name)]; ok {
		return e.UserID
	}
	// match by first-name only (single match)
	first := strings.ToLower(strings.Fields(name)[0])
	var matches []contactEntry
	for k, v := range contacts.ByName {
		if k == first || strings.HasPrefix(k, first+" ") {
			matches = append(matches, v)
		}
	}
	if len(matches) == 1 {
		return matches[0].UserID
	}
	return ""
}

func recentContactsLookupHandle(handle string) string {
	contacts.load()
	contacts.mu.Lock()
	defer contacts.mu.Unlock()
	if e, ok := contacts.ByHandle[strings.ToLower(handle)]; ok {
		return e.UserID
	}
	return ""
}

func recentContactsAdd(name, userID, username string) {
	contacts.load()
	contacts.mu.Lock()
	defer contacts.mu.Unlock()
	now := time.Now().Unix()
	contacts.ByName[strings.ToLower(name)] = contactEntry{UserID: userID, Username: username, LastSeen: now}
	if username != "" {
		contacts.ByHandle[strings.ToLower(username)] = contactEntry{UserID: userID, Username: username, LastSeen: now}
	}
	contacts.Updated = now
	_ = writeCache("recent_contacts.json", contacts)
}

func recentContactsClear() error {
	contacts.mu.Lock()
	defer contacts.mu.Unlock()
	contacts.ByName = map[string]contactEntry{}
	contacts.ByHandle = map[string]contactEntry{}
	contacts.Updated = 0
	contacts.loaded = true
	return os.Remove(cachePath("recent_contacts.json"))
}

// --- user display-name cache (user_id → display name) ---

type userCacheT struct {
	mu      sync.Mutex
	loaded  bool
	Updated int64             `json:"updated_at"`
	ByID    map[string]string `json:"by_id"`
}

var userNames userCacheT

func (c *userCacheT) load() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.loaded {
		return
	}
	c.ByID = map[string]string{}
	var buf userCacheT
	if readCache("users.json", &buf) {
		if time.Since(time.Unix(buf.Updated, 0)) < userCacheTTL && buf.ByID != nil {
			c.ByID = buf.ByID
		}
	}
	c.loaded = true
}

func userNameGet(id string) string {
	userNames.load()
	userNames.mu.Lock()
	defer userNames.mu.Unlock()
	return userNames.ByID[id]
}

func userNameSet(id, name string) {
	userNames.load()
	userNames.mu.Lock()
	defer userNames.mu.Unlock()
	userNames.ByID[id] = name
	userNames.Updated = time.Now().Unix()
	_ = writeCache("users.json", userNames)
}

// --- rate limit state ---

type rateState struct {
	mu      sync.Mutex
	loaded  bool
	Updated int64                 `json:"updated_at"`
	Calls   map[string][]int64    `json:"calls"` // method → unix-second timestamps in last 60s
}

var rate rateState

const rateWindow = 60 * time.Second

func (r *rateState) load() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.loaded {
		return
	}
	r.Calls = map[string][]int64{}
	var buf rateState
	if readCache("rate_state.json", &buf) && buf.Calls != nil {
		r.Calls = buf.Calls
	}
	r.loaded = true
}

// rateWait records this call and sleeps if we'd exceed our local cap (50/min).
func rateWait(method string) {
	rate.load()
	rate.mu.Lock()
	defer rate.mu.Unlock()
	now := time.Now().Unix()
	cutoff := now - int64(rateWindow.Seconds())
	pruned := rate.Calls[method][:0]
	for _, t := range rate.Calls[method] {
		if t >= cutoff {
			pruned = append(pruned, t)
		}
	}
	rate.Calls[method] = pruned
	cap := 50 // generic tier-3 cap; safe lower bound for most write methods
	if strings.HasPrefix(method, "search.") {
		cap = 20
	}
	if len(pruned) >= cap {
		oldest := pruned[0]
		sleep := time.Until(time.Unix(oldest, 0).Add(rateWindow + 100*time.Millisecond))
		if sleep > 0 {
			rate.mu.Unlock()
			time.Sleep(sleep)
			rate.mu.Lock()
		}
	}
	rate.Calls[method] = append(rate.Calls[method], now)
	rate.Updated = now
	_ = writeCache("rate_state.json", rate)
}

func rateStatus() map[string]int {
	rate.load()
	rate.mu.Lock()
	defer rate.mu.Unlock()
	now := time.Now().Unix()
	cutoff := now - int64(rateWindow.Seconds())
	out := map[string]int{}
	for method, calls := range rate.Calls {
		n := 0
		for _, t := range calls {
			if t >= cutoff {
				n++
			}
		}
		if n > 0 {
			out[method] = n
		}
	}
	return out
}

func rateClear() error {
	rate.mu.Lock()
	defer rate.mu.Unlock()
	rate.Calls = map[string][]int64{}
	rate.Updated = 0
	rate.loaded = true
	return os.Remove(cachePath("rate_state.json"))
}
