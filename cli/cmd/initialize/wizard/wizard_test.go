package wizard

import (
	"testing"

	"numerous/cli/manifest"

	"github.com/stretchr/testify/assert"
)

func TestGetQuestions(t *testing.T) {
	t.Run("given empty manifest gets all questions", func(t *testing.T) {
		qs := getQuestions(&manifest.Manifest{})

		assert.Len(t, qs, 5)
	})

	t.Run("given full manifest gets no questions", func(t *testing.T) {
		qs := getQuestions(&manifest.Manifest{
			Name:             "Some name",
			Description:      "Some description",
			Library:          manifest.LibraryNumerous,
			AppFile:          "app.py",
			RequirementsFile: "requirements.txt",
		})

		assert.Empty(t, qs)
	})
}
