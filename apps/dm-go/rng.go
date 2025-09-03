package main

import (
	"math/rand"
)

// SeededRNG provides seeded random number generation
type SeededRNG struct {
	rng *rand.Rand
}

// NewSeededRNG creates a new seeded RNG
func NewSeededRNG(seed int64) *SeededRNG {
	source := rand.NewSource(seed)
	return &SeededRNG{
		rng: rand.New(source),
	}
}

// RollD20 rolls a d20 (1-20)
func (s *SeededRNG) RollD20() int {
	return s.rng.Intn(20) + 1
}

// RollD6 rolls a d6 (1-6)
func (s *SeededRNG) RollD6() int {
	return s.rng.Intn(6) + 1
}

// RollD100 rolls a d100 (1-100)
func (s *SeededRNG) RollD100() int {
	return s.rng.Intn(100) + 1
}

// RandomInt returns a random int between min and max inclusive
func (s *SeededRNG) RandomInt(min, max int) int {
	if min > max {
		return min
	}
	return s.rng.Intn(max-min+1) + min
}