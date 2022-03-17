package handlers

import (
	"net/http"

	"github.com/shoheiKU/web_application/GO_UDEMY/pkg/config"
	"github.com/shoheiKU/web_application/GO_UDEMY/pkg/models"
	"github.com/shoheiKU/web_application/GO_UDEMY/pkg/render"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
}

// NewHandlers make a Repository
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(repo *Repository) {
	Repo = repo
}

// Home is the handler for the home page
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "home.page.tmpl", &models.TemplateData{})
}

// About is the handler for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "about.page.tmpl", &models.TemplateData{})
}

// Porker is the handler for the porker page
func (m *Repository) Poker(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "poker.page.tmpl", &models.TemplateData{})
}

// Contact is the handler for the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "contact.page.tmpl", &models.TemplateData{})
}
