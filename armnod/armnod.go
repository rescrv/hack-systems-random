// armnod generates random strings for use in benchmarking.
//
// The core idea of armnod is that a configuration provides three different
// parameters that give deterministic string generation from a deterministic
// PRNG (specifically, guacamole).
//
// Start with a Charset to specify the runes that can appear in an output
// string.  Charsets are UTF8 aware and the generated strings will pull
// runes from the charset and use those runes to create the output.  This means
// that the length of the output is measured rune-wise, not byte-wise.
//
// To specify how the Charset turns into random strings, provide a
// StringChooser and an LengthChooser.  The StringChooser defines the strategy
// for selecting the next string's random seed and the LengthChooser defines how
// many runes to select using said seed.  It may sound strange to have two
// levels of random number generators, but this makes it possible to draw output
// strings from a fixed-size set using a Zipf distribution.
package armnod

import (
	"math"

	"hack.systems/random/guacamole"
)

// Characters to be selected from when composing random strings.
type Charset string

// Pre-defined character sets to generate common patterns of strings.
const (
	// ASCII-centric definitions
	LowerLetters Charset = "abcdefghijklmnopqrstuvwxyz"
	UpperLetters Charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Letters      Charset = LowerLetters + UpperLetters
	Digits       Charset = "0123456789"
	Alphanumeric Charset = Letters + Digits
	Punctuation  Charset = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
	// Programmer-ish
	HexLower  Charset = "0123456789abcdef"
	HexUpper  Charset = "0123456789ABCDEF"
	ModHex    Charset = "cbdefghijklnrtuv"
	Base64    Charset = UpperLetters + LowerLetters + Digits + "+/"
	Base64URL Charset = UpperLetters + LowerLetters + Digits + "-_"
	// TODO(rescrv): find some good unicode tests

	// Default
	Default Charset = Alphanumeric + Punctuation
)

// Configuration of a random string generator.  If the configuration is used to
// create multiple generators that could be called concurrently, the provided
// Choosers must be reentrant and not embed any memory.  All choosers provided
// by this package satisfy this requirement.
//
// The *Chooser interfaces are required to consume a constant number of bytes
// of guacamole on each call to Next.  This is done to assure deterministic
// generation of strings for the more complex chooser arrangments.
type Configuration struct {
	Charset
	StringChooser
	LengthChooser
}

// StringChooser uses the provided PRNG to specify the next string to generate.
// Returns (X, true) to indicate that the next generated string should begin at
// seed X and (_, false) to indicate that no more strings should be generated.
type StringChooser interface {
	NextArmnodString(*guacamole.Guacamole) (uint64, bool)
}

// LengthChooser uses the provided PRNG to specify the length of the next string
// to generate.  As an optimization, the generator uses MaxArmnodLength to
// preallocate a buffer, and it is an error for NextArmnodLength to ever return
// a value larger than the max.
type LengthChooser interface {
	MaxArmnodLength() uint64
	NextArmnodLength(*guacamole.Guacamole) uint64
}

// Generator builds a generator object from the provided configuration.
func (c Configuration) Generator() *Generator {
	g := &Generator{
		configuration: c,
	}
	g.initialize()
	return g
}

// Generator creates and returns random strings according to a provided
// configuration.  Do not instantiate a generator directly; instead, use the
// Generator method on a configuration object to create and initialize a
// generator.
type Generator struct {
	configuration Configuration
	// stretched runes
	runes [runeStretchLength]rune
	// buffer to avoid repeated allocation
	bbuf []byte
	rbuf []rune
	// random bytes used to pump the StringChooser
	random *guacamole.Guacamole
	// random bytes seeded by the StringChooser and used in subsequent choosers
	strings *guacamole.Guacamole
}

func (g *Generator) initialize() {
	if len(g.configuration.Charset) > runeStretchLength/2 {
		panic("large charset support not implemented")
	}
	if g.configuration.StringChooser == nil {
		g.configuration.StringChooser = &DefaultStringChooser{}
	}
	if g.configuration.LengthChooser == nil {
		g.configuration.LengthChooser = &ConstantLengthChooser{10}
	}
	if g.random == nil {
		g.random = guacamole.New()
	}
	if g.strings == nil {
		g.strings = guacamole.New()
	}
	charset := []rune(string(g.configuration.Charset))
	// TODO(rescrv): Consider moving this to C; it's about an order of magnitude
	// slower in Go.  I assume this is because of bounds checks as unrolling the
	// loop had a positive effect in the C code and has zero effect in Go,
	// hinting that maybe there's some branching getting in the way.
	for i := 0; i < runeStretchLength; i++ {
		d := int(float64(i) / float64(runeStretchLength) * float64(len(charset)))
		if d < 0 || d >= runeStretchLength {
			panic("invariant violated")
		}
		g.runes[i] = charset[d]
	}
	length := g.configuration.LengthChooser.MaxArmnodLength()
	g.bbuf = make([]byte, length)
	g.rbuf = make([]rune, length)
	g.Seed(0)
}

// Seed the random number generator with the provided constant.  Seed is not the
// same as a typical random number generator because the integer distance
// separating two seeds directly correlates with the number of random strings
// between them.
//
// See the guacamole Seed function for more discussion of this atypical
// behavior.
func (g *Generator) Seed(seed uint64) {
	g.random.Seed(seed)
}

// String returns the next string to be generated or signals that all such
// strings have been generated.  Returns (s, true) when s is the next random
// string to be generated and (_, false) when generation is finished.
func (g *Generator) String() (string, bool) {
	idx, ok := g.configuration.StringChooser.NextArmnodString(g.random)
	if !ok {
		return "", false
	}
	g.strings.Seed(idx)
	length := g.configuration.LengthChooser.NextArmnodLength(g.strings)
	g.strings.Fill(g.bbuf)
	for i := uint64(0); i < length; i++ {
		g.rbuf[i] = g.runes[g.bbuf[i]]
	}
	s := string(g.rbuf[:length])
	return s, true
}

// DefaultStringChooser provides the default behavior of string choice being
// unconstrained.  It is theoretically possible for this chooser to generate
// every string imaginable for a given charset and LengthChooser.  This is not
// guaranteed, but is possible and very likely for shorter strings.
type DefaultStringChooser struct{}

// NextArmnodString returns the next string in the random string chooser.  The
// next string is randomly pulled from the provided PRNG.
func (c *DefaultStringChooser) NextArmnodString(g *guacamole.Guacamole) (uint64, bool) {
	return g.Uint64(), true
}

// ChooseFromFixedSet constructs a StringChooser that will repeatedly select
// from a set of N random strings.  Strings will be chosen indefinitely and
// uniformly at random.
func ChooseFromFixedSet(N uint64) StringChooser {
	return &fixedStringChooser{
		N: N,
	}
}

// ChooseFromFixedSetZipf constructs a StringChooser that will repeatedly select
// from a set of N random strings.  Strings will be chosen indefinitely and the
// selection will be an approximate Zipf distribution according to the given
// parameters.
func ChooseFromFixedSetZipf(params *guacamole.ZipfParams) StringChooser {
	return &fixedStringChooserZipf{
		zp: params,
	}
}

// InitializedFixedSet constructions a string chooser that returns every string
// in a set of N random strings exactly once.  After all strings are returned,
// the string chooser will stop generating strings.
func InitializeFixedSet(N uint64) StringChooser {
	return InitializeFixedSlice(N, 0, N)
}

// InitializedFixedSet constructions a string chooser that returns every string
// in a subset of N random strings exactly once.  After all strings are
// returned, the string chooser will stop generating strings.
func InitializeFixedSlice(N, start, limit uint64) StringChooser {
	return &initFixedStringChooser{
		N:     N,
		limit: limit,
		idx:   start,
	}
}

// ConstantLengthChooser specifies that all strings must be the specified
// length.
type ConstantLengthChooser struct {
	Length uint64
}

// MaxArmnodLength implements LengthChooser
func (c ConstantLengthChooser) MaxArmnodLength() uint64 {
	return c.Length
}

// NextArmnodLength implements LengthChooser
func (c ConstantLengthChooser) NextArmnodLength(*guacamole.Guacamole) uint64 {
	return c.Length
}

// ConstantLengthChooser specifies that all strings have a length chosen at
// random and uniformly distributed between Min and Max.
type UniformLengthChooser struct {
	Min uint64
	Max uint64
}

// MaxArmnodLength implements LengthChooser
func (c UniformLengthChooser) MaxArmnodLength() uint64 {
	return c.Max
}

// NextArmnodLength implements LengthChooser
func (c UniformLengthChooser) NextArmnodLength(g *guacamole.Guacamole) uint64 {
	return c.Min + uint64(float64(c.Max-c.Min)*g.Float64())
}

// implementation

const (
	runeStretchLength int = 256
)

// BUG(rescrv): Internal constant runeStretchLength specifies an upper bound on
// the number of characters in a charset and the closer the charset length gets
// to that limit, the less even the representation of characters in the output
// will be.

func distribute(x, c uint64) uint64 {
	return x * (math.MaxUint64 / c)
}

type fixedStringChooser struct {
	N uint64
}

func (c *fixedStringChooser) NextArmnodString(g *guacamole.Guacamole) (uint64, bool) {
	return distribute(uint64(float64(c.N)*g.Float64()), c.N), true
}

type fixedStringChooserZipf struct {
	zp *guacamole.ZipfParams
}

func (c *fixedStringChooserZipf) NextArmnodString(g *guacamole.Guacamole) (uint64, bool) {
	return distribute(g.Zipf(c.zp)-1, c.zp.N()), true
}

type initFixedStringChooser struct {
	N     uint64
	limit uint64
	idx   uint64
}

func (c *initFixedStringChooser) NextArmnodString(g *guacamole.Guacamole) (uint64, bool) {
	if c.idx < c.limit {
		x, done := distribute(c.idx, c.N), true
		c.idx++
		return x, done
	}
	return 0, false
}
