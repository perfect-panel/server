package orm

import "testing"

func TestLikePrefixPatternEscapesWithPortableEscapeChar(t *testing.T) {
	got := LikePrefixPattern("alice_100%=x")
	want := "alice=_100=%==x%"
	if got != want {
		t.Fatalf("LikePrefixPattern() = %q, want %q", got, want)
	}
}

func TestLikeContainsPatternEscapesWithPortableEscapeChar(t *testing.T) {
	got := LikeContainsPattern("50%_off")
	want := "%50=%=_off%"
	if got != want {
		t.Fatalf("LikeContainsPattern() = %q, want %q", got, want)
	}
}
