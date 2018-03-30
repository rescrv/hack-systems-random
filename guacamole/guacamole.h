/* Copyright (c) 2013-2018, Robert Escriva
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 *     * Redistributions of source code must retain the above copyright notice,
 *       this list of conditions and the following disclaimer.
 *     * Redistributions in binary form must reproduce the above copyright
 *       notice, this list of conditions and the following disclaimer in the
 *       documentation and/or other materials provided with the distribution.
 *     * Neither the name of this project nor the names of its contributors
 *       may be used to endorse or promote products derived from this software
 *       without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */

#ifndef guacamole_h_
#define guacamole_h_

/* C */
#include <stdint.h>
#include <stdlib.h>

/* generate random bytes */
struct guacamole
{
    uint64_t nonce;
    unsigned index;
    union {
        uint32_t u32[16];
        uint32_t u32v[16][4];
        unsigned char byte[256];
    } __attribute__ ((aligned (64))) buffer;
};
void guacamole_seed(struct guacamole* g, uint64_t seed);
void guacamole_generate(struct guacamole* g, void* bytes, size_t bytes_sz);
uint32_t guacamole_uint32(struct guacamole* g);
double guacamole_double(struct guacamole* g);

/* draw numbers froma  Zipf distribution with the given parameters */
struct guacamole_zipf_params
{
    uint64_t n;
    double alpha;
    double theta;
    double zetan;
    double zeta2;
    double eta;
};
void guacamole_zipf_init_alpha(uint64_t n, double alpha, struct guacamole_zipf_params* p);
void guacamole_zipf_init_theta(uint64_t n, double theta, struct guacamole_zipf_params* p);
uint64_t guacamole_zipf(struct guacamole* g, struct guacamole_zipf_params* p);

/* scramble the given value through the specified bijection
 * useful for turning zipf output into values spread out in space
 */
#define BLF_N   16          /* Number of Subkeys */
struct guacamole_scrambler
{
    uint32_t S[4][256]; /* S-Boxes */
    uint32_t P[BLF_N + 2];  /* Subkeys */
};
void guacamole_scrambler_change(struct guacamole_scrambler* gs, uint64_t bijection);
uint64_t guacamole_scramble(struct guacamole_scrambler* gs, uint64_t value);

/* low level 64-bit number to 64-byte output; safe to sequentially increment #
 *
 * this is derived fromt the salsa encryption scheme, except the key and
 * ciphertext were made constant in order to speed up the routine.
 */
void guacamole_mash(uint64_t number, uint32_t output[16]);
void guacamole_disable_assembly();
void guacamole_maybe_enable_assembly();

#endif /* guacamole_h_ */
