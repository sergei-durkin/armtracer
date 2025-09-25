//+build arm64,!gccgo,!noasm,!appengine

#include "textflag.h"

// func getCntvct() uint64
TEXT ·getCntvct(SB), NOSPLIT, $0
    MRS     CNTVCT_EL0, R0
    MOVD    R0,         ret+0(FP)
    RET

// func getCntfrq() uint64
TEXT ·getCntfrq(SB), NOSPLIT, $0
    MRS     CNTFRQ_EL0, R0
    MOVD    R0,         ret+0(FP)
    RET
