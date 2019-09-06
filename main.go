// townstats returns information about tilde.town in the tilde data protcol format
// It was originally a Python script written by Michael F. Lamb <https://datagrok.org>
// License: GPLv3+

// TDP is defined at http://protocol.club/~datagrok/beta-wiki/tdp.html
// It is a JSON structure of the form:

// Usage: stats > /var/www/html/tilde.json

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

// TODO read ENV var for additional system users but hard code the ones we know about

const default_html_filename = "/etc/skel/public_html/index.html"
const description = `an intentional digital community for creating and sharing
works of art, peer education, and technological anachronism. we are
non-commercial, donation supported, and committed to rejecting false
technological progress in favor of empathy and sustainable computing.`

type NewsEntry struct {
	Title   string `json:"title"`   // Title of entry
	Pubdate string `json:"pubdate"` // Human readable date
	Content string `json:"content"` // HTML of entry
}

type User struct {
	Username  string `json:"username"` // Username of user
	PageTitle string `json:"title"`    // Title of user's HTML page, if they have one
	Mtime     int    `json:"mtime"`    // Timestamp representing the last time a user's index.html was modified
	// Town additions
	DisplayName string `json:"display_name"` // Display Name of user
	DefaultPage bool   `json:"default"`      // Whether or not user has updated their default index.html
	Favicon     string `json:"favicon"`      // URL to a small image representing the user
}

type TildeData struct {
	Name        string `json:"name"`        // Name of the server
	URL         string `json:"url"`         // URL of the server's homepage
	SignupURL   string `json:"signup_url"`  // URL for server's signup page
	WantUsers   bool   `json:"want_users"`  // Whether or not new users are being accepted
	AdminEmail  string `json:"admin_email"` // Email for server admin
	Description string `json:"description"` // Description of server
	UserCount   int    `json:"user_count"`  // Total number of users on server sorted by last activity time
	Users       []User `json:"users"`
	// Town Additions
	LiveUserCount   int         `json:"live_user_count"`   // Users who have changed their index.html
	ActiveUserCount int         `json:"active_user_count"` // Users with an active session
	GeneratedAt     string      `json:"generated_at"`      // When this was generated in '%Y-%m-%d %H:%M:%S' format
	GeneratedAtSec  int64       `json:"generated_at_sec"`  // When this was generated in seconds since epoch
	Uptime          string      `json:"uptime"`            // output of `uptime -p`
	News            []NewsEntry // Collection of town news entries
}

func news() []NewsEntry {
	// TODO
	return []NewsEntry{}
}

func userCount() int {
	// TODO
	return 0
}

func users() []User {
	// TODO
	return []User{}
}

func liveUserCount() int {
	// TODO
	return 0
}

func activeUserCount() int {
	out, err := exec.Command("who").Output()
	scanner := bufio.NewScanner(bytes.NewReader(out))
	if err != nil {
		log.Fatalf("could not run who %s", err)
	}

	activeUsers := map[string]bool{}

	for scanner.Scan() {
		whoLine := scanner.Text()
		username := strings.Split(whoLine, " ")[0]
		activeUsers[username] = true
	}

	return len(activeUsers)
}

func uptime() string {
	out, err := exec.Command("uptime").Output()
	if err != nil {
		log.Fatalf("could not run uptime %s", err)
	}
	return strings.TrimSpace(string(out))
}

func tdp() TildeData {
	return TildeData{
		Name:            "tilde.town",
		URL:             "https://tilde.town",
		SignupURL:       "https://cgi.tilde.town/users/signup",
		WantUsers:       true,
		AdminEmail:      "root@tilde.town",
		Description:     description,
		UserCount:       userCount(),
		Users:           users(),
		LiveUserCount:   liveUserCount(),
		ActiveUserCount: activeUserCount(),
		Uptime:          uptime(),
		News:            news(),
		GeneratedAt:     time.Now().UTC().Format("2006-01-02 15:04:05"),
		GeneratedAtSec:  time.Now().Unix(),
	}
}

func main() {
	data, err := json.Marshal(tdp())
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %s", err)
	}
	fmt.Printf("%s\n", data)
}
