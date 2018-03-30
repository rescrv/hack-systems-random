// guacamole provides a seekable PRNG.
//
// This package provides a pseudo-random number generator (PRNG) that provides
// constant-time indexing into the random number stream.  Where most PRNGs do
// not provide correlation between the seed and the input stream, guacamole
// guaranees that consecutive seeds will produce outputs exactly 64B apart.  For
// example, seeding the generator with some value i and reading 64B of data has
// the exact same effect as seeding with the value i+1.
//
// This linear seeding property makes it possible to use the random number
// stream to reproducibly generate and verify data in tests.  For example, the
// armnod package transforms raw bytes from guacamole to into human-readable
// strings that use a specified character set.  In general, the ability to use
// an integer index into the random stream makes it easier to generate random
// data and reconstruct the data for validation by seeking to the specified
// offset in the random stream.
//
// The implementation of guacamole is derived from the Salsa stream cipher from
// DJB.  The name stems from a misunderstanding DJB's naming conventions in
// which Salsa the dance was confused with Salsa the delicious chip dip.  The
// changes from Salsa were largely to choose a constant key and inline many
// values in the algorithm.  For amd64 hardware, there's a hand-written
// implementation of guacamole that offers even more performance and is
// byte-wise identical to the regular C implementation.
//
// For historical reasons, the package also includes routines for drawing
// numbers from a Zipf distribution and for scrambling integers in pseudo-random
// ways.
//
// In addition to being great at random byte generation, the module gives many
// opportunities for puns about "bytes of guacamole".
package guacamole

// #include <errno.h>
// #include <stdlib.h>
// #include "guacamole.h"
import "C"

import (
	"encoding/binary"
	"unsafe"
)

func init() {
	MaybeEnableAssembly()
}

// DisableAssembly ensures that the assembly implementation cannot be used, even
// on processors with the necessary primitives.  This is largely intended for
// debugging, but is exposed in cases it is more broadly useful.
func DisableAssembly() {
	C.guacamole_disable_assembly()
}

// MaybeEnableAssembly allows the use of the optimized assembly implementation
// if the processor is detected to support the necessary SSE 4.1 instructions.
// This function is called by default because there is should be no harm in
// using the fastest implementation available.
func MaybeEnableAssembly() {
	C.guacamole_maybe_enable_assembly()
}

// New creates a new guacamole generator.  The generator comes seeded at 0 and
// is ready to eat... err... use.
func New() *Guacamole {
	g := &Guacamole{}
	g.Seed(0)
	return g
}

// Guacamole is the central class for generating random bytes.  It is safe to
// initialize this directly instead of allocating it via New, but the behavior
// is undefined until the first call to Seed.  New calls Seed directly before
// returning.
type Guacamole struct {
	guac C.struct_guacamole
}

// Seed the guacamole (would that be "avocado"?).  The seed function is fast and
// safe to call relatively frequently.
func (g *Guacamole) Seed(s uint64) {
	C.guacamole_seed(&g.guac, C.uint64_t(s))
}

// String constructs a string of the next sz random bytes of guacamole.
func (g *Guacamole) String(sz uint64) string {
	return string(g.Bytes(sz))
}

// Bytes constructs a slice of the next sz random bytes of guacamole.
func (g *Guacamole) Bytes(sz uint64) []byte {
	bytes := make([]byte, sz)
	g.Fill(bytes)
	return bytes
}

// Fill fills the provided slice with random guacamole bytes.
func (g *Guacamole) Fill(bytes []byte) {
	C.guacamole_generate(&g.guac, unsafe.Pointer(&bytes[0]), C.size_t(len(bytes)))
}

// Uint64 returns a new uint64 that is uniformly distributed throughout the 2^64
// space.
func (g *Guacamole) Uint64() uint64 {
	var bytes [8]byte
	g.Fill(bytes[:])
	return binary.BigEndian.Uint64(bytes[:])
}

// Float64 generates a random float64 in the range [0, 1).
func (g *Guacamole) Float64() float64 {
	return float64(C.guacamole_double(&g.guac))
}

// ZipfParams specify a set of elements and the parameters to select them
// according to a zipf distribution.  Due to the approximation used, it is
// possible the last couple elements of N may not be generated.  This may be a
// bug, or it may be an expected result of Gray's Zipf algorithm.
type ZipfParams struct {
	gzp C.struct_guacamole_zipf_params
}

// N specifies the number of elements in the set from which values are selected.
func (z *ZipfParams) N() uint64 {
	return uint64(z.gzp.n)
}

// ZipfAlpha returns ZipfParams to draw from n elements with the provided alpha
// parameter.
func ZipfAlpha(n uint64, alpha float64) *ZipfParams {
	zp := &ZipfParams{}
	C.guacamole_zipf_init_alpha(C.uint64_t(n), C.double(alpha), &zp.gzp)
	return zp
}

// BUG(rescrv): Zipf may not return the last few (on the order of 1%) elements
// of the random set.  See the ZipfParams struct for details.

// ZipfAlpha returns ZipfParams to draw from n elements with the provided theta
// parameter.
func ZipfTheta(n uint64, theta float64) *ZipfParams {
	zp := &ZipfParams{}
	C.guacamole_zipf_init_theta(C.uint64_t(n), C.double(theta), &zp.gzp)
	return zp
}

// Zipf returns an element from the provided ZipfParams.  The return value will
// be in the range [1, N].
func (g *Guacamole) Zipf(zp *ZipfParams) uint64 {
	return uint64(C.guacamole_zipf(&g.guac, &zp.gzp))
}

// Scrambler turns any set of uint64 numbers into a completely jumbled
// set of uint64 numbers.  The function guarantees that each input will map to a
// unique output no matter how much of the input space is used.
type Scrambler struct {
	scr C.struct_guacamole_scrambler
}

// Create a new scrambler and initialize it with Change(0)
func Scrambler() *Scrambler {
	s := &Scrambler{}
	s.Change(0)
	return s
}

// Change the bijection used to the scrambler to the one provided.  Each
// bijection is deterministic, so it is always possible to remember the
// bijection number and later recover the same mapping.
func (s *Scrambler) Change(bijection uint64) {
	C.guacamole_scrambler_change(&s.scr, C.uint64_t(bijection))
}

// Scramble x through the bijection to generate a unique value for it.  The
// function is deterministic and will produce the same output for scramblers
// with the same change and x value.  Note that while a scrambler is a
// bijection, the scramble function is one way; there is [currently] no way to
// reverse the mapping efficiently.
func (s *Scrambler) Scramble(x uint64) uint64 {
	return uint64(C.guacamole_scramble(&s.scr, C.uint64_t(x)))
}
