package main

import (
	"fmt"

	"hack.systems/random/guacamole"
)

func dump(zp *guacamole.ZipfParams) {
	n, alpha, theta, zetan, zeta2, eta := zp.Dump()
	fmt.Printf("\t{%d, %g, %g, %g, %g, %g},\n", n, alpha, theta, zetan, zeta2, eta)
}

func main() {
	fmt.Printf("static struct guacamole_zipf_params precomputed[] = {\n")
	for _, N := range []uint64{1e7, 1e8, 1e9, 1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16} {
		for _, theta := range []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9} {
			zp := guacamole.ZipfTheta(N, theta)
			dump(zp)
		}
		for _, alpha := range []float64{1, 10, 100, 1000, 10000} {
			zp := guacamole.ZipfAlpha(N, alpha)
			dump(zp)
		}
	}
	fmt.Printf("}\n")
}
