package templates

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplates(t *testing.T) {
	t.Parallel()

	templates := New()

	var buf bytes.Buffer
	err := templates.Render(&buf, "index.html", Dashboard{
		IsAuthenticated: true,
	}, nil)
	require.NoError(t, err)
}
