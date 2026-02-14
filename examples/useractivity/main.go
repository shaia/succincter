// Example: User activity tracking and leaderboard queries
//
// Gaming and social platforms need efficient queries on user activity:
// - "How many users were active before user #50000?"
// - "Who is the 100th most recently active user?"
//
// Run with: go run ./examples/useractivity
package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/shaia/succincter"
)

type User struct {
	ID             int
	Username       string
	LastActive     time.Time
	IsOnline       bool
	IsPremium      bool
	Score          int
	AccountCreated time.Time
}

func main() {
	fmt.Println("=== User Activity Tracking Example ===")

	// Simulate a user database (100K users)
	numUsers := 100_000
	fmt.Printf("\nGenerating %d users...\n", numUsers)
	users := generateUsers(numUsers)

	// Build indices for different user segments
	fmt.Println("Building user indices...")
	start := time.Now()

	onlineIndex := succincter.NewSuccincter(users, func(u User) bool {
		return u.IsOnline
	})
	premiumIndex := succincter.NewSuccincter(users, func(u User) bool {
		return u.IsPremium
	})
	highScorerIndex := succincter.NewSuccincter(users, func(u User) bool {
		return u.Score >= 1000
	})
	recentlyActiveIndex := succincter.NewSuccincter(users, func(u User) bool {
		return time.Since(u.LastActive) < 24*time.Hour
	})

	fmt.Printf("Indices built in %v\n\n", time.Since(start))

	// User segment statistics
	fmt.Println("--- User Segments ---")
	online := onlineIndex.Rank(numUsers)
	premium := premiumIndex.Rank(numUsers)
	highScorers := highScorerIndex.Rank(numUsers)
	recentlyActive := recentlyActiveIndex.Rank(numUsers)

	fmt.Printf("Online users:        %6d (%.1f%%)\n", online, pct(online, numUsers))
	fmt.Printf("Premium users:       %6d (%.1f%%)\n", premium, pct(premium, numUsers))
	fmt.Printf("High scorers (1000+): %6d (%.1f%%)\n", highScorers, pct(highScorers, numUsers))
	fmt.Printf("Active (24h):        %6d (%.1f%%)\n", recentlyActive, pct(recentlyActive, numUsers))

	// Find specific users by rank
	fmt.Println("\n--- Leaderboard Queries ---")
	pos1 := onlineIndex.Select(1)
	fmt.Printf("1st online user:    %s (ID: %d)\n", users[pos1].Username, users[pos1].ID)
	pos100 := onlineIndex.Select(100)
	fmt.Printf("100th online user:  %s (ID: %d)\n", users[pos100].Username, users[pos100].ID)
	pos1000 := onlineIndex.Select(1000)
	fmt.Printf("1000th online user: %s (ID: %d)\n", users[pos1000].Username, users[pos1000].ID)

	// Premium user pagination
	fmt.Println("\n--- Premium User Pagination (Page 5, 10 per page) ---")
	page, pageSize := 5, 10
	startRank := (page-1)*pageSize + 1
	for i := 0; i < pageSize; i++ {
		pos := premiumIndex.Select(startRank + i)
		if pos == -1 {
			break
		}
		u := users[pos]
		fmt.Printf("  %2d. %-15s Score: %4d  Online: %v\n",
			startRank+i, u.Username, u.Score, u.IsOnline)
	}

	// Count segments in user ID ranges (useful for sharding analysis)
	fmt.Println("\n--- Segment Distribution by User ID Range ---")
	ranges := [][2]int{{0, 25000}, {25000, 50000}, {50000, 75000}, {75000, 100000}}
	fmt.Printf("%-20s %10s %10s %10s\n", "Range", "Online", "Premium", "High Score")
	for _, r := range ranges {
		onlineInRange := onlineIndex.Rank(r[1]) - onlineIndex.Rank(r[0])
		premiumInRange := premiumIndex.Rank(r[1]) - premiumIndex.Rank(r[0])
		highInRange := highScorerIndex.Rank(r[1]) - highScorerIndex.Rank(r[0])
		fmt.Printf("[%5d, %5d)       %10d %10d %10d\n",
			r[0], r[1], onlineInRange, premiumInRange, highInRange)
	}

	// Find users with multiple attributes
	fmt.Println("\n--- Online Premium High-Scorers ---")
	count := 0
	for i := 1; i <= onlineIndex.Rank(numUsers) && count < 5; i++ {
		pos := onlineIndex.Select(i)
		u := users[pos]
		if u.IsPremium && u.Score >= 1000 {
			count++
			fmt.Printf("  %d. %-15s Score: %d\n", count, u.Username, u.Score)
		}
	}
	if count == 0 {
		fmt.Println("  No users found matching all criteria")
	}

	// Performance comparison hint
	fmt.Println("\n--- Query Performance ---")
	fmt.Println("All queries above execute in O(1) or O(log n) time,")
	fmt.Println("compared to O(n) for naive filtering approaches.")
}

func generateUsers(n int) []User {
	users := make([]User, n)
	adjectives := []string{"Swift", "Brave", "Silent", "Mighty", "Clever", "Noble", "Fierce", "Calm"}
	nouns := []string{"Wolf", "Eagle", "Tiger", "Dragon", "Phoenix", "Bear", "Hawk", "Lion"}

	baseTime := time.Now()

	for i := range users {
		adj := adjectives[rand.Intn(len(adjectives))]
		noun := nouns[rand.Intn(len(nouns))]

		users[i] = User{
			ID:             i,
			Username:       fmt.Sprintf("%s%s%d", adj, noun, rand.Intn(1000)),
			LastActive:     baseTime.Add(-time.Duration(rand.Intn(72)) * time.Hour),
			IsOnline:       rand.Float64() < 0.15, // 15% online
			IsPremium:      rand.Float64() < 0.10, // 10% premium
			Score:          rand.Intn(2000),
			AccountCreated: baseTime.AddDate(0, 0, -rand.Intn(365)),
		}
	}

	// Sort by ID for consistent ordering
	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})

	return users
}

func pct(count, total int) float64 {
	return float64(count) * 100 / float64(total)
}
