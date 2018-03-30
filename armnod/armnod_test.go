package armnod_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"hack.systems/random/armnod"
	"hack.systems/random/guacamole"
)

func TestArmnodConstantLength(t *testing.T) {
	require := require.New(t)

	c := armnod.Configuration{}
	c.Charset = armnod.Default
	c.LengthChooser = armnod.ConstantLengthChooser{
		Length: 10,
	}
	g := c.Generator()
	require.NotNil(g)

	s1, ok := g.String()
	require.True(ok)
	s2, ok := g.String()
	require.True(ok)
	require.NotEqual(s1, s2)
}

func TestArmnodUniform(t *testing.T) {
	require := require.New(t)

	c := armnod.Configuration{}
	c.Charset = armnod.Default
	c.LengthChooser = armnod.UniformLengthChooser{
		Min: 8,
		Max: 16,
	}
	g := c.Generator()
	require.NotNil(g)

	s1, ok := g.String()
	require.True(ok)
	s2, ok := g.String()
	require.True(ok)
	require.NotEqual(s1, s2)
}

func TestArmnodZipf(t *testing.T) {
	require := require.New(t)

	c := armnod.Configuration{}
	c.Charset = armnod.Default
	c.StringChooser = armnod.ChooseFromFixedSetZipf(guacamole.ZipfTheta(10, 0.99))
	g := c.Generator()
	require.NotNil(g)

	strings := make(map[string]int)
	for i := 0; i < 10000; i++ {
		s, ok := g.String()
		require.True(ok)
		strings[s]++
	}
	require.True(len(strings) < 10)
}

var result string

func BenchmarkArmnodDefault(b *testing.B) {
	c := armnod.Configuration{}
	c.Charset = armnod.Default
	g := c.Generator()

	for n := 0; n < b.N; n++ {
		if s, ok := g.String(); ok {
			result = s
		}
	}
}

func BenchmarkArmnodZipf(b *testing.B) {
	c := armnod.Configuration{}
	c.Charset = armnod.Default
	c.StringChooser = armnod.ChooseFromFixedSetZipf(guacamole.ZipfTheta(10, 0.99))
	g := c.Generator()

	for n := 0; n < b.N; n++ {
		if s, ok := g.String(); ok {
			result = s
		}
	}
}
