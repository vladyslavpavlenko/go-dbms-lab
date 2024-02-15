package handlers

import "github.com/vladyslavpavlenko/go-dbms-lab/internal/config"

// Repo is a global variable that holds a pointer to a Repository instance.
// This allows for easy access to the repository across the handlers package.
var Repo *Repository

// Repository encapsulates the application configuration, providing a structured way
// to pass around application state and dependencies.
type Repository struct {
	App *config.AppConfig
}

// NewRepo creates and returns a new instance of Repository.
// This function is used to initialize the repository with the application configuration.
func NewRepo(app *config.AppConfig) *Repository {
	return &Repository{
		App: app,
	}
}

// NewHandlers initializes the global Repo variable with the provided repository instance.
// This function is typically called at application startup to set up the handlers with
// the necessary application state.
func NewHandlers(r *Repository) {
	Repo = r
}
