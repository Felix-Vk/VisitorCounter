package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
)

type Stats struct {
	TotalVisits    int             `json:"total_visits"`
	UniqueVisitors map[string]bool `json:"unique_visitors"`
	UserVisits     map[string]int  `json:"user_visits"`
}

var (
	mu    sync.Mutex
	stats Stats
	file  = "stats.json"
)

// Load stats from file (or create new)
func loadStats() {
	f, err := os.Open(file)
	if err != nil {
		stats = Stats{
			TotalVisits:    0,
			UniqueVisitors: make(map[string]bool),
			UserVisits:     make(map[string]int),
		}
		return
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&stats)
	if err != nil {
		log.Println("Error decoding stats, starting fresh:", err)
		stats = Stats{
			TotalVisits:    0,
			UniqueVisitors: make(map[string]bool),
			UserVisits:     make(map[string]int),
		}
	}

	// Ensure maps are not nil
	if stats.UniqueVisitors == nil {
		stats.UniqueVisitors = make(map[string]bool)
	}
	if stats.UserVisits == nil {
		stats.UserVisits = make(map[string]int)
	}
}

// Save stats to file
func saveStats() {
	f, err := os.Create(file)
	if err != nil {
		log.Println("Error saving stats:", err)
		return
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(&stats)
	if err != nil {
		log.Println("Error encoding stats:", err)
	}
}

// Get client IP
func getIP(r *http.Request) string {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// Homepage handler (personal + global stats)
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/favicon.ico" {
		return
	}

	ip := getIP(r)

	mu.Lock()
	stats.TotalVisits++
	stats.UniqueVisitors[ip] = true
	stats.UserVisits[ip]++
	saveStats()

	personalVisits := stats.UserVisits[ip]
	total := stats.TotalVisits
	unique := len(stats.UniqueVisitors)
	mu.Unlock()

	fmt.Fprintf(w, `<h1>Welcome!</h1>
	<p>You have visited <b>%d</b> times.</p>
	<p>Total visits across all users: <b>%d</b></p>
	<p>Unique visitors: <b>%d</b></p>
	<p>Check <a href="/stats">/stats</a> for live stats</p>`,
		personalVisits, total, unique)
}

// Stats handler (global + per-visitor breakdown)
func statsHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	fmt.Fprintf(w, `<h1>Stats</h1>
	<p>Total visits: %d</p>
	<p>Unique visitors: %d</p>
	<h2>Per-Visitor Breakdown:</h2><ul>`,
		stats.TotalVisits, len(stats.UniqueVisitors))

	for ip, count := range stats.UserVisits {
		fmt.Fprintf(w, "<li>%s → %d visits</li>", ip, count)
	}
	fmt.Fprintf(w, "</ul>")
}

// Favicon handler (ignore requests)
func faviconHandler(w http.ResponseWriter, r *http.Request) {
	// Do nothing to prevent double counting
}

func main() {
	loadStats()
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/stats", statsHandler)
	http.HandleFunc("/favicon.ico", faviconHandler)

	log.Println("Visitor counter running on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Server failed:", err)
	}
}
