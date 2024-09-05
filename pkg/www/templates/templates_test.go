package templates

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTemplates(t *testing.T) {
	templates := New()

	var buf bytes.Buffer
	err := templates.Render(&buf, "index.html", Dashboard{
		IsAuthenticated: true,
	}, nil)
	require.NoError(t, err)
}
