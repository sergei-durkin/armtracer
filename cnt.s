//+build arm64,!gccgo,!noasm,!appengine

#include "textflag.h"

// func getCnt() uint64
TEXT ·getCnt(SB), NOSPLIT, $0
    MRS     CNTPCT_EL0, R0
    MOVD    R0,         ret+0(FP)
    RET
