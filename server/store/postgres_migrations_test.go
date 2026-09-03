package store

import "testing"

func TestPostgresPlaceholders(t *testing.T) {
	got := postgresPlaceholders("SELECT * FROM parts WHERE value = ? AND status = '?' AND id = ?")
	want := "SELECT * FROM parts WHERE value = $1 AND status = '?' AND id = $2"
	if got != want {
		t.Fatalf("postgresPlaceholders() = %q, want %q", got, want)
	}
}
