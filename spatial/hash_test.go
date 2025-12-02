package spatial_test

import (
	"testing"

	"github.com/rafibayer/ants-again/spatial"
	vec "github.com/rafibayer/ants-again/vector"
	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	sp := spatial.NewHash[vec.Vector](1.0)
	sp.Insert(vec.Vector{X: 5, Y: 5})

	p := sp.Points()
	require.Len(t, p, 1)
	require.Equal(t, vec.Vector{X: 5, Y: 5}, p[0])

	sp.Remove(vec.Vector{X: 4, Y: 4})
	p = sp.Points()
	require.Len(t, p, 1)
	require.Equal(t, vec.Vector{X: 5, Y: 5}, p[0])

	r := sp.Remove(vec.Vector{X: 5, Y: 5})
	require.Equal(t, vec.Vector{X: 5, Y: 5}, r)
	p = sp.Points()
	require.Len(t, p, 0)
}
