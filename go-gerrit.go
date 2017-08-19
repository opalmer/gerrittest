package gerrittest

// This file contains copies of structs and function from the
// github.com/andygrunwald/go-gerrit project. Because gerrittest may be a
// dependency of go-gerrit we want to limit the possibility of circular
// dependencies.

// AccountInfo entity contains information about an account.
type AccountInfo struct {
	AccountID int    `json:"_account_id,omitempty"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
	Username  string `json:"username,omitempty"`
}

// HTTPPasswordInput entity contains information for setting/generating an HTTP password.
type HTTPPasswordInput struct {
	Generate     bool   `json:"generate,omitempty"`
	HTTPPassword string `json:"http_password,omitempty"`
}
