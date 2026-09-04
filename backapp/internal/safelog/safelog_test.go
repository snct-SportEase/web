package safelog

import "testing"

func TestValueEscapesLineBreaks(t *testing.T) {
	got := Value("trusted\nforged\rentry")
	want := `trusted\nforged\rentry`
	if got != want {
		t.Fatalf("Value() = %q, want %q", got, want)
	}
}
