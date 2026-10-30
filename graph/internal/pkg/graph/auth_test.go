package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthFromEnv(t *testing.T) {
	t.Setenv("GRAPH_USER", "alice")
	t.Setenv("GRAPH_PASSWORD", "secret")

	tokens := authFromEnv().Tokens
	assert.Equal(t, "alice", tokens["principal"])
	assert.Equal(t, "secret", tokens["credentials"])
}

func TestAuthFromEnvDefault(t *testing.T) {
	t.Setenv("GRAPH_USER", "")
	t.Setenv("GRAPH_PASSWORD", "")

	tokens := authFromEnv().Tokens
	assert.Equal(t, "", tokens["principal"])
	assert.Equal(t, "", tokens["credentials"])
}
