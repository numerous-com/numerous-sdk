package wizard

import (
	"testing"

	"numerous.com/cli/internal/manifest"

	"github.com/stretchr/testify/assert"
)

func TestGetQuestions(t *testing.T) {
	t.Run("given empty manifest gets all questions", func(t *testing.T) {
		qs := getQuestions(&manifest.Manifest{
			Python: &manifest.Python{},
		})

		assert.Len(t, qs, 5)
	})

	t.Run("given full manifest gets no questions", func(t *testing.T) {
		qs := getQuestions(&manifest.Manifest{
			App: manifest.App{
				Name:        "Some name",
				Description: "Some description",
			},
			Python: &manifest.Python{
				Library:          manifest.LibraryNumerous,
				AppFile:          "app.py",
				RequirementsFile: "requirements.txt",
			},
		})

		assert.Empty(t, qs)
	})
}
