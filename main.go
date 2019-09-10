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
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const defaultIndexPath = "/etc/skel/public_html/index.html"
const description = `an intentional digital community for creating and sharing
works of art, peer education, and technological anachronism. we are
non-commercial, donation supported, and committed to rejecting false
technological progress in favor of empathy and sustainable computing.`

type newsEntry struct {
	Title   string `json:"title"`   // Title of entry
	Pubdate string `json:"pubdate"` // Human readable date
	Content string `json:"content"` // HTML of entry
}

type user struct {
	Username  string `json:"username"` // Username of user
	PageTitle string `json:"title"`    // Title of user's HTML page, if they have one
	Mtime     int64  `json:"mtime"`    // Timestamp representing the last time a user's index.html was modified
	// Town additions
	DefaultPage bool `json:"default"` // Whether or not user has updated their default index.html
}

type tildeData struct {
	Name        string  `json:"name"`        // Name of the server
	URL         string  `json:"url"`         // URL of the server's homepage
	SignupURL   string  `json:"signup_url"`  // URL for server's signup page
	WantUsers   bool    `json:"want_users"`  // Whether or not new users are being accepted
	AdminEmail  string  `json:"admin_email"` // Email for server admin
	Description string  `json:"description"` // Description of server
	UserCount   int     `json:"user_count"`  // Total number of users on server sorted by last activity time
	Users       []*user `json:"users"`
	// Town Additions
	LiveUserCount   int         `json:"live_user_count"`   // Users who have changed their index.html
	ActiveUserCount int         `json:"active_user_count"` // Users with an active session
	GeneratedAt     string      `json:"generated_at"`      // When this was generated in '%Y-%m-%d %H:%M:%S' format
	GeneratedAtSec  int64       `json:"generated_at_sec"`  // When this was generated in seconds since epoch
	Uptime          string      `json:"uptime"`            // output of `uptime -p`
	News            []newsEntry // Collection of town news entries
}

func homesDir() string {
	hDir := os.Getenv("HOMES_DIR")
	if hDir == "" {
		hDir = "/home"
	}

	return hDir
}

func getNews() (entries []newsEntry, err error) {
	inMeta := true
	inContent := false
	current := newsEntry{}
	blankLineRe := regexp.MustCompile(`^ *\n$`)

	newsPath := os.Getenv("NEWS_PATH")
	if newsPath == "" {
		newsPath = "/town/news.posts"
	}

	newsFile, err := os.Open(newsPath)
	if err != nil {
		return entries, fmt.Errorf("unable to read news file: %s", err)
	}
	defer newsFile.Close()

	scanner := bufio.NewScanner(newsFile)

	for scanner.Scan() {
		newsLine := scanner.Text()
		if strings.HasPrefix(newsLine, "#") || newsLine == "" || blankLineRe.FindStringIndex(newsLine) != nil {
			continue
		} else if strings.HasPrefix(newsLine, "--") {
			entries = append(entries, current)
			current = newsEntry{}
			inMeta = true
			inContent = false
		} else if inMeta {
			kv := strings.SplitN(newsLine, ":", 2)
			if kv[0] == "pubdate" {
				current.Pubdate = strings.TrimSpace(kv[1])
			} else if kv[0] == "title" {
				current.Title = strings.TrimSpace(kv[1])
			} else {
				log.Printf("Ignoring unknown metadata in news entry: %v\n", newsLine)
			}
			if current.Pubdate != "" && current.Title != "" {
				inMeta = false
				inContent = true
			}
		} else if inContent {
			current.Content += fmt.Sprintf("\n%v", strings.TrimSpace(newsLine))
		}
	}
	return entries, nil
}

func indexPathFor(username string) (string, error) {
	potentialPaths := []string{"index.html", "index.htm"}
	indexPath := ""
	errs := []error{}
	for _, p := range potentialPaths {
		fullPath := path.Join(homesDir(), username, "public_html", p)
		_, staterr := os.Stat(fullPath)
		if staterr != nil {
			errs = append(errs, staterr)
		} else {
			indexPath = fullPath
			break
		}
	}

	if indexPath == "" {
		return "", fmt.Errorf("Failed to locate index file for %v; tried %v; encountered errors: %v", username, potentialPaths, errs)
	}

	return indexPath, nil
}

func pageTitleFor(username string) string {
	pageTitleRe := regexp.MustCompile(`<title[^>]*>(.*)</title>`)
	indexPath, err := indexPathFor(username)
	if err != nil {
		log.Print(err)
		return ""
	}
	content, err := ioutil.ReadFile(indexPath)
	if err != nil {
		log.Printf("failed to read %q: %v\n", indexPath, err)
		return ""
	}
	titleMatch := pageTitleRe.FindStringSubmatch(string(content))
	if len(titleMatch) < 2 {
		return ""
	}
	return titleMatch[1]
}

func systemUsers() map[string]bool {
	systemUsers := map[string]bool{
		"ubuntu":  true,
		"ttadmin": true,
		"root":    true,
	}
	envSystemUsers := os.Getenv("SYSTEM_USERS")
	if envSystemUsers != "" {
		for _, username := range strings.Split(envSystemUsers, ",") {
			systemUsers[username] = true
		}
	}

	return systemUsers
}

func mtimeFor(username string) int64 {
	path := path.Join(homesDir(), username, "public_html")
	var maxMtime int64 = 0
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if maxMtime < info.ModTime().Unix() {
			maxMtime = info.ModTime().Unix()
		}
		return nil
	})
	if err != nil {
		log.Printf("error walking %q: %v\n", path, err)
	}

	return maxMtime
}

func detectDefaultPageFor(username string, defaultHTML []byte) bool {
	indexPath, err := indexPathFor(username)
	if err != nil {
		log.Print(err)
		return false
	}
	indexFile, err := os.Open(indexPath)
	if err != nil {
		return false
	}
	defer indexFile.Close()

	indexHTML, err := ioutil.ReadAll(indexFile)
	if err != nil {
		return false
	}
	return bytes.Equal(indexHTML, defaultHTML)
}

func getDefaultHTML() ([]byte, error) {
	indexPath := os.Getenv("DEFAULT_INDEX_PATH")
	if indexPath == "" {
		indexPath = defaultIndexPath
	}

	defaultIndexFile, err := os.Open(indexPath)
	if err != nil {
		return []byte{}, fmt.Errorf("could not open default index: %s", err)
	}
	defer defaultIndexFile.Close()

	defaultIndexHTML, err := ioutil.ReadAll(defaultIndexFile)
	if err != nil {
		return []byte{}, fmt.Errorf("could not read default index: %s", err)
	}

	return defaultIndexHTML, nil
}

type usersByMtime []*user

func getUsers() (users []*user, err error) {
	// TODO sort by mtime
	// For the purposes of this program, we discover users via:
	// - presence in /home/
	// - absence in systemUsers list (sourced from source code and potentially augmented by an environment variable)
	// We formally used passwd parsing. This is definitely more "correct" and I'm
	// not opposed to going back to that; going back to parsing /home is mainly to
	// get this new version going.
	defaultIndexHTML, err := getDefaultHTML()
	if err != nil {
		return users, err
	}

	out, err := exec.Command("ls", homesDir()).Output()
	if err != nil {
		return users, fmt.Errorf("could not run ls: %s", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))

	systemUsers := systemUsers()

	for scanner.Scan() {
		username := scanner.Text()
		if systemUsers[username] {
			continue
		}
		user := user{
			Username:    username,
			PageTitle:   pageTitleFor(username),
			Mtime:       mtimeFor(username),
			DefaultPage: detectDefaultPageFor(username, defaultIndexHTML),
		}
		users = append(users, &user)
	}

	return users, nil
}

func liveUserCount(users []*user) int {
	count := 0
	for _, u := range users {
		if !u.DefaultPage {
			count++
		}
	}
	return count
}

func activeUserCount() (int, error) {
	out, err := exec.Command("who").Output()
	if err != nil {
		return 0, fmt.Errorf("could not run who: %s", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))

	activeUsers := map[string]bool{}

	for scanner.Scan() {
		whoLine := scanner.Text()
		username := strings.Split(whoLine, " ")[0]
		activeUsers[username] = true
	}

	return len(activeUsers), nil
}

func getUptime() (string, error) {
	out, err := exec.Command("uptime").Output()
	if err != nil {
		return "", fmt.Errorf("could not run uptime: %s", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func tdp() (tildeData, error) {
	users, err := getUsers()
	if err != nil {
		return tildeData{}, fmt.Errorf("could not get user list: %s", err)
	}
	activeUsers, err := activeUserCount()
	if err != nil {
		return tildeData{}, fmt.Errorf("could not count non-default users: %s", err)
	}
	news, err := getNews()
	if err != nil {
		return tildeData{}, fmt.Errorf("could not get news: %s", err)
	}

	uptime, err := getUptime()
	if err != nil {
		return tildeData{}, fmt.Errorf("could not determine uptime: %s", err)
	}

	return tildeData{
		Name:            "tilde.town",
		URL:             "https://tilde.town",
		SignupURL:       "https://cgi.tilde.town/users/signup",
		WantUsers:       true,
		AdminEmail:      "root@tilde.town",
		Description:     description,
		UserCount:       len(users),
		Users:           users,
		LiveUserCount:   liveUserCount(users),
		ActiveUserCount: activeUsers,
		Uptime:          uptime,
		News:            news,
		GeneratedAt:     time.Now().UTC().Format("2006-01-02 15:04:05"),
		GeneratedAtSec:  time.Now().Unix(),
	}, nil
}

func main() {
	systemData, err := tdp()
	if err != nil {
		log.Fatal(err)
	}
	data, err := json.Marshal(systemData)
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %s", err)
	}
	fmt.Printf("%s\n", data)
}
