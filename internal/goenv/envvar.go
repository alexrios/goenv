package goenv

// EnvVar represents a Go environment variable with its current value and change status.
type EnvVar struct {
	Key      string
	Value    string
	Changed  bool
	Favorite bool
}
