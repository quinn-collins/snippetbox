package main

import "snippetbox.qcollins.net/internal/models"

type templateData struct {
  Snippet *models.Snippet
  Snippets []*models.Snippet
}
