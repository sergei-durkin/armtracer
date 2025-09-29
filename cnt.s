//+build arm64,armtracer,!gccgo,!noasm

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

// func isb()
TEXT ·isb(SB), NOSPLIT, $0-0
    WORD $0xD5033FDF // ISB SY instruction (ARM64)
    RET
