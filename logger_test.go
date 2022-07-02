package nblogger

import (
	"regexp"
	"testing"
)

func TestHello(t *testing.T) {
	name := "Test"
	want := regexp.MustCompile(`\b` + name + `\b`)
	msg, err := Hello("Test")

	if !want.MatchString(msg) || err != nil {
		t.Fatalf(`Hello("Test") = %q, %v, want match for %#q, nil`, msg, err, want)
	}
}

func TestEmpty(t *testing.T) {
	msg, err := Hello("")
	if msg != "" || err == nil {
		t.Fatalf(`Hello("") = %q, %v, want "", error`, msg, err)
	}
}
