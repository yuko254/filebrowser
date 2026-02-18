package users

// Shortcut represents a named path shortcut for a user.
type Shortcut struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	System bool   `json:"system,omitempty"`
}
