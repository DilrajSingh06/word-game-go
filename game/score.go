// Package game holds the pure rules of the word game.
//
// No database, no HTTP, no Redis. Everything here is a plain function over
// plain data, which is exactly what makes it worth unit testing.
package game

import "strings"

// Mark is the colour given to one letter of a guess.
type Mark string

const (
	Correct Mark = "correct" // right letter, right position  (green)
	Present Mark = "present" // right letter, wrong position  (yellow)
	Absent  Mark = "absent"  // letter not in the answer       (grey)
)

// WordLength is how many letters every word has.
const WordLength = 5

// MaxGuesses is how many attempts a player gets.
const MaxGuesses = 6

// Score compares a guess against the answer and returns one Mark per letter.
//
// The subtle part is duplicate letters. A letter in the answer can only be
// "claimed" once. If the answer is CRANE and the guess is EERIE, there is
// exactly one E in the answer, so only ONE of the guess's three E's may be
// coloured - and greens get first claim.
//
// The algorithm is therefore two passes:
//  1. mark every exact position match, and consume that letter from a count
//  2. walk the rest, handing out Present only while a letter remains unclaimed
//
// Doing it in one pass is the classic bug.
func Score(guess, answer string) []Mark {
	guess = strings.ToUpper(guess)
	answer = strings.ToUpper(answer)

	marks := make([]Mark, len(guess))
	remaining := make(map[rune]int)

	g := []rune(guess)
	a := []rune(answer)

	// pass 1 - greens
	for i := range g {
		if i < len(a) && g[i] == a[i] {
			marks[i] = Correct
		} else if i < len(a) {
			remaining[a[i]]++
		}
	}

	// pass 2 - yellows, only while that letter is still unclaimed
	for i := range g {
		if marks[i] == Correct {
			continue
		}
		if remaining[g[i]] > 0 {
			marks[i] = Present
			remaining[g[i]]--
		} else {
			marks[i] = Absent
		}
	}

	return marks
}

// IsWin reports whether every letter was marked Correct.
func IsWin(marks []Mark) bool {
	if len(marks) == 0 {
		return false
	}
	for _, m := range marks {
		if m != Correct {
			return false
		}
	}
	return true
}

// ValidateGuess checks a guess before it is scored.
func ValidateGuess(guess string, allowed map[string]bool) error {
	g := strings.ToUpper(strings.TrimSpace(guess))
	if len(g) != WordLength {
		return ErrWrongLength
	}
	for _, r := range g {
		if r < 'A' || r > 'Z' {
			return ErrNotLetters
		}
	}
	if allowed != nil && !allowed[g] {
		return ErrNotAWord
	}
	return nil
}
