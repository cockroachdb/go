// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// See memmove Go doc for important implementation constraints.

// func memmove(to, from unsafe.Pointer, n uintptr)
TEXT runtime·memmove(SB),NOSPLIT|NOFRAME,$0-24
	MOVD	to+0(FP), R6
	MOVD	from+8(FP), R4
	MOVD	n+16(FP), R5

	CMPBEQ	R6, R4, done

start:
moveLE16:
	CMPBLE	R5, $3, move0to3
	CMPBLE	R5, $7, move4to7
	CMPBLE	R5, $11, move8to11
	CMPBLE	R5, $15, move12to15
	CMPBNE	R5, $16, movemt16
	MOVD	0(R4), R7
	MOVD	8(R4), R8
	MOVD	R7, 0(R6)
	MOVD	R8, 8(R6)
	RET

movemt16:
	CMPBGT	R4, R6, forwards_copy
	ADD	R5, R4, R7
	CMPBLE	R7, R6, forwards_copy
	ADD	R5, R6, R8

// backwards_copy is used in below scenario:
// 1. When src and dst are overlapping, and the dst is at higher address than src.
backwards_copy:
	MOVD	R5, R8
	SRD	$4, R8
	ADD	$-16, R5
	CMPBGE	R8, $8, moveGE128

moveLT128:
	MOVD	8(R4)(R5), R3
	MOVD	R3, 8(R6)(R5)
	MOVD	0(R4)(R5), R3
	MOVD	R3, 0(R6)(R5)
	ADD	$-16, R5
	BRCTG	R8, moveLT128
	ADD	$16, R5
	BR	moveLE16

moveGE128:
	ADD	$-48, R5
	SRD	$2, R8, R0

move_large_64_loop:
	VL	0(R4)(R5), V1
	VL	16(R4)(R5), V2
	VL	32(R4)(R5), V3
	VL	48(R4)(R5), V4
	VST	V1, 0(R6)(R5)
	VST	V2, 16(R6)(R5)
	VST	V3, 32(R6)(R5)
	VST	V4, 48(R6)(R5)
	ADD	$-64, R5
	BRCTG	R0, move_large_64_loop
	ADD	$48, R5
	AND	$3, R8
	BNE	moveLT128
	ADD	$16, R5
	BR	moveLE16

// forwards_copy is used in below scenarios:
// 1. When src and dst are non-overlapping.
// 2. When src and dst are overlapping, but src is at higher address than dst.
forwards_copy:
	MOVD	R5, R8
	SRD	$8, R8
	CMPBNE	R8, $0, moveGE256

use_exrl:
	CMPBEQ	R5, $0, done
	ADD	$-1, R5
	EXRL	$memmove_exrl_mvc<>(SB), R5
	RET

moveGE256:
	CMP	R8, $4096
	BGT	moveGT1MB

mvc_loop:
	MVC	$256, 0(R4), 0(R6)
	ADD	$256, R4
	ADD	$256, R6
	ADD	$-256, R5
	BRCTG	R8, mvc_loop
	BR	use_exrl

moveGT1MB:
	MOVD	R5, R7

mvcle_loop:
	MVCLE	0, R4, R6
	BVS	mvcle_loop
	RET

move0to3:
	CMPBEQ	R5, $0, done
move1:
	CMPBNE	R5, $1, move2
	MOVB	0(R4), R3
	MOVB	R3, 0(R6)
	RET
move2:
	CMPBNE	R5, $2, move3
	MOVH	0(R4), R3
	MOVH	R3, 0(R6)
	RET
move3:
	MOVH	0(R4), R3
	MOVB	2(R4), R7
	MOVH	R3, 0(R6)
	MOVB	R7, 2(R6)
	RET

move4to7:
	CMPBNE	R5, $4, move5
	MOVW	0(R4), R3
	MOVW	R3, 0(R6)
	RET
move5:
	CMPBNE	R5, $5, move6
	MOVW	0(R4), R3
	MOVB	4(R4), R7
	MOVW	R3, 0(R6)
	MOVB	R7, 4(R6)
	RET
move6:
	CMPBNE	R5, $6, move7
	MOVW	0(R4), R3
	MOVH	4(R4), R7
	MOVW	R3, 0(R6)
	MOVH	R7, 4(R6)
	RET
move7:
	MOVW	0(R4), R3
	MOVH	4(R4), R7
	MOVB	6(R4), R8
	MOVW	R3, 0(R6)
	MOVH	R7, 4(R6)
	MOVB	R8, 6(R6)
	RET

move8to11:
	CMPBNE	R5, $8, move9
	MOVD	0(R4), R3
	MOVD	R3, 0(R6)
	RET
move9:
	CMPBNE	R5, $9, move10
	MOVD	0(R4), R3
	MOVB	8(R4), R7
	MOVD	R3, 0(R6)
	MOVB	R7, 8(R6)
	RET
move10:
	CMPBNE	R5, $10, move11
	MOVD	0(R4), R3
	MOVH	8(R4), R7
	MOVD	R3, 0(R6)
	MOVH	R7, 8(R6)
	RET
move11:
	MOVD	0(R4), R3
	MOVH	8(R4), R7
	MOVB	10(R4), R8
	MOVD	R3, 0(R6)
	MOVH	R7, 8(R6)
	MOVB	R8, 10(R6)
	RET

move12to15:
	CMPBNE	R5, $12, move13
	MOVD	0(R4), R3
	MOVW	8(R4), R7
	MOVD	R3, 0(R6)
	MOVW	R7, 8(R6)
	RET
move13:
	CMPBNE	R5, $13, move14
	MOVD	0(R4), R3
	MOVW	8(R4), R7
	MOVB	12(R4), R8
	MOVD	R3, 0(R6)
	MOVW	R7, 8(R6)
	MOVB	R8, 12(R6)
	RET
move14:
	CMPBNE	R5, $14, move15
	MOVD	0(R4), R3
	MOVW	8(R4), R7
	MOVH	12(R4), R8
	MOVD	R3, 0(R6)
	MOVW	R7, 8(R6)
	MOVH	R8, 12(R6)
	RET
move15:
	MOVD	0(R4), R3
	MOVW	8(R4), R7
	MOVH	12(R4), R8
	MOVB	14(R4), R10
	MOVD	R3, 0(R6)
	MOVW	R7, 8(R6)
	MOVH	R8, 12(R6)
	MOVB	R10, 14(R6)
done:
	RET

// DO NOT CALL - target for exrl (execute relative long) instruction.
TEXT memmove_exrl_mvc<>(SB),NOSPLIT|NOFRAME,$0-0
	MVC	$1, 0(R4), 0(R6)
	MOVD	R0, 0(R0)
	RET

