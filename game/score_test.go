package game

import (
    "reflect"
    "testing"
)

func TestScore(t *testing.T) {
    tests := []struct {
        name, guess, answer string
        want []Mark
    }{
        {"all letters correct", "CRANE", "CRANE", []Mark{Correct, Correct, Correct, Correct, Correct}},
        {"all letters absent", "XXXXX", "CRANE", []Mark{Absent, Absent, Absent, Absent, Absent}},
        {"duplicate letters", "EERIE", "CRANE", []Mark{Absent, Absent, Present, Absent, Correct}},
        {"case insensitive", "crane", "CRANE", []Mark{Correct, Correct, Correct, Correct, Correct}},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Score(tt.guess, tt.answer)
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Score(%q, %q) = %v; want %v", tt.guess, tt.answer, got, tt.want)
            }
        })
    }
}

func TestIsWin(t *testing.T) {
    tests := []struct {
        name string
        marks []Mark
        want bool
    }{
        {"all correct", []Mark{Correct, Correct, Correct, Correct, Correct}, true},
        {"one absent", []Mark{Correct, Correct, Absent, Correct, Correct}, false},
        {"one present", []Mark{Correct, Present, Correct, Correct, Correct}, false},
        {"empty marks", []Mark{}, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := IsWin(tt.marks)
            if got != tt.want {
                t.Errorf("IsWin(%v) = %v; want %v", tt.marks, got, tt.want)
            }
        })
    }
}
