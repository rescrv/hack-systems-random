package guacamole_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"hack.systems/random/guacamole"
)

const (
	First8Bytes string  = "\f\xedYO\xb6\x19K\xe6"
	FirstDouble float64 = 0.3099659152297931
)

func TestGuacamoleC(t *testing.T) {
	require := require.New(t)
	g := guacamole.New()
	require.NotNil(g)

	guacamole.DisableAssembly()

	g.Seed(0)
	s := g.String(8)
	require.Equal(First8Bytes, s)

	g.Seed(0)
	b := g.Bytes(8)
	require.Equal([]byte(First8Bytes), b)

	b = make([]byte, 8)
	require.NotEqual([]byte(First8Bytes), b)
	g.Seed(0)
	g.Fill(b)
	require.Equal([]byte(First8Bytes), b)

	g.Seed(0)
	require.InEpsilon(FirstDouble, g.Float64(), 0.001)
}

func TestGuacamoleAssembly(t *testing.T) {
	require := require.New(t)
	g := guacamole.New()
	require.NotNil(g)

	guacamole.MaybeEnableAssembly()

	g.Seed(0)
	s := g.String(8)
	require.Equal(First8Bytes, s)

	g.Seed(0)
	b := g.Bytes(8)
	require.Equal([]byte(First8Bytes), b)

	b = make([]byte, 8)
	require.NotEqual([]byte(First8Bytes), b)
	g.Seed(0)
	g.Fill(b)
	require.Equal([]byte(First8Bytes), b)

	g.Seed(0)
	require.InEpsilon(FirstDouble, g.Float64(), 0.001)
}

func TestGuacamolePassToFunction(t *testing.T) {
	require := require.New(t)
	g := guacamole.New()
	require.NotNil(g)

	x := func(guac *guacamole.Guacamole) []byte {
		buf := make([]byte, 8)
		guac.Fill(buf)
		return buf
	}(g)
	require.NotNil(x)
}

func TestZipf(t *testing.T) {
	require := require.New(t)
	g := guacamole.New()
	require.NotNil(g)

	type parameter struct {
		theta    float64
		expected uint64
	}
	parameters := []parameter{
		{0.1, 1751},
		{0.2, 3199},
		{0.3, 5518},
		{0.4, 9596},
		{0.5, 16253},
		{0.6, 26834},
		{0.7, 42145},
		{0.8, 64628},
		{0.9, 95126},
		{1.0, 133414},
	}

	for _, p := range parameters {
		zp := guacamole.ZipfTheta(1000, p.theta)
		counts := [2]uint64{}
		for i := 0; i < 1000000; i++ {
			if g.Zipf(zp) == 1 {
				counts[0]++
			} else {
				counts[1]++
			}
		}
		require.Equal(p.expected, counts[0])
	}
}

func TestScrambler(t *testing.T) {
	require := require.New(t)

	s := guacamole.NewScrambler()

	require.Equal(uint64(0x4ef997456198dd78), s.Scramble(0))
	require.Equal(uint64(0x64ed065757511fa7), s.Scramble(1))
	require.Equal(uint64(0xad63166cadf9811e), s.Scramble(2))
	require.Equal(uint64(0x731ca6b4131b7ef1), s.Scramble(3))
	require.Equal(uint64(0xac25f3282f17f3b2), s.Scramble(4))
}

func benchmarkGuacamoleBytes(num int, maybeASM bool, b *testing.B) []byte {
	if maybeASM {
		guacamole.MaybeEnableAssembly()
	} else {
		guacamole.DisableAssembly()
	}
	g := guacamole.New()
	bytes := make([]byte, num)
	for n := 0; n < b.N; n++ {
		g.Fill(bytes)
	}
	return bytes
}

func BenchmarkCGuacamole64B(b *testing.B)  { benchmarkGuacamoleBytes(64, false, b) }
func BenchmarkCGuacamole1KB(b *testing.B)  { benchmarkGuacamoleBytes(1024, false, b) }
func BenchmarkCGuacamole4KB(b *testing.B)  { benchmarkGuacamoleBytes(4096, false, b) }
func BenchmarkCGuacamole64KB(b *testing.B) { benchmarkGuacamoleBytes(65536, false, b) }
func BenchmarkCGuacamole1MB(b *testing.B)  { benchmarkGuacamoleBytes(1048576, false, b) }

func BenchmarkASMGuacamole64B(b *testing.B)  { benchmarkGuacamoleBytes(64, true, b) }
func BenchmarkASMGuacamole1KB(b *testing.B)  { benchmarkGuacamoleBytes(1024, true, b) }
func BenchmarkASMGuacamole4KB(b *testing.B)  { benchmarkGuacamoleBytes(4096, true, b) }
func BenchmarkASMGuacamole64KB(b *testing.B) { benchmarkGuacamoleBytes(65536, true, b) }
func BenchmarkASMGuacamole1MB(b *testing.B)  { benchmarkGuacamoleBytes(1048576, true, b) }

var result uint64

func benchmarkZipfTheta(N uint64, theta float64, b *testing.B) {
	guacamole.DisableAssembly()
	zp := guacamole.ZipfTheta(N, theta)
	g := guacamole.New()
	sum := uint64(0)
	for n := 0; n < b.N; n++ {
		sum += g.Zipf(zp)
	}
	result = sum
}

func BenchmarkZipfTheta_1e3_01(b *testing.B) { benchmarkZipfTheta(1000, 0.1, b) }
func BenchmarkZipfTheta_1e3_05(b *testing.B) { benchmarkZipfTheta(1000, 0.1, b) }
func BenchmarkZipfTheta_1e3_09(b *testing.B) { benchmarkZipfTheta(1000, 0.1, b) }

func BenchmarkZipfTheta_1e6_01(b *testing.B) { benchmarkZipfTheta(1000, 0.1, b) }
func BenchmarkZipfTheta_1e6_05(b *testing.B) { benchmarkZipfTheta(1000, 0.1, b) }
func BenchmarkZipfTheta_1e6_09(b *testing.B) { benchmarkZipfTheta(1000, 0.1, b) }

func BenchmarkZipfTheta_1e9_01(b *testing.B) { benchmarkZipfTheta(1000, 0.1, b) }
func BenchmarkZipfTheta_1e9_05(b *testing.B) { benchmarkZipfTheta(1000, 0.1, b) }
func BenchmarkZipfTheta_1e9_09(b *testing.B) { benchmarkZipfTheta(1000, 0.1, b) }

func BenchmarkScramblerChange(b *testing.B) {
	s := guacamole.NewScrambler()
	for n := 0; n < b.N; n++ {
		s.Change(uint64(n))
	}
	result = s.Scramble(uint64(0))
}

func BenchmarkScramblerScramble(b *testing.B) {
	s := guacamole.NewScrambler()
	sum := uint64(0)
	for n := 0; n < b.N; n++ {
		sum += s.Scramble(uint64(n))
	}
	result = sum
}
