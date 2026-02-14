// Example: Prime number indexing
//
// Run with: go run ./examples/primes
package main

import (
	"fmt"

	"github.com/shaia/succincter"
)

func main() {
	fmt.Println("=== Prime Numbers Example ===")

	// Generate numbers 0-99
	numbers := make([]int, 100)
	for i := range numbers {
		numbers[i] = i
	}

	// Create index on prime numbers
	primeIndex := succincter.NewSuccincter(numbers, isPrime)

	// Count primes
	totalPrimes := primeIndex.Rank(100)
	fmt.Printf("Prime numbers in [0, 100): %d\n\n", totalPrimes)

	// Find specific primes by rank
	fmt.Println("First 10 primes:")
	for i := 1; i <= 10; i++ {
		pos := primeIndex.Select(i)
		fmt.Printf("  Prime #%d = %d\n", i, numbers[pos])
	}

	// Count primes in ranges
	fmt.Println("\nPrimes in ranges:")
	ranges := [][2]int{{0, 10}, {10, 20}, {20, 50}, {50, 100}}
	for _, r := range ranges {
		count := primeIndex.Rank(r[1]) - primeIndex.Rank(r[0])
		fmt.Printf("  Primes in [%d, %d): %d\n", r[0], r[1], count)
	}

	// Find the 25th prime
	pos25 := primeIndex.Select(25)
	fmt.Printf("\nThe 25th prime is: %d\n", numbers[pos25])
}

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}
	for i := 3; i*i <= n; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}
