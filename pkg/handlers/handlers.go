package handlers

import (
	"net/http"

	"github.com/shoheiKU/golang_poker/pkg/config"
	"github.com/shoheiKU/golang_poker/pkg/models"
	"github.com/shoheiKU/golang_poker/pkg/render"
)

// Repo the repository used by the handlers.
var Repo *Repository

// Repository is the repository type.
type Repository struct {
	App       *config.AppConfig
	PokerRepo *PokerRepository
}

// NewRepo make a Repository.
func NewRepo(
	a *config.AppConfig,
	p *PokerRepository,
) *Repository {
	return &Repository{
		App:       a,
		PokerRepo: p,
	}
}

// NewHandlers sets the repository for the handlers.
func NewHandlers(repo *Repository) {
	Repo = repo
}

// Home is the handler for the home page.
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About is the handler for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "about.page.tmpl", &models.TemplateData{})
}

// Contact is the handler for the contact page.
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "contact.page.tmpl", &models.TemplateData{})
}
