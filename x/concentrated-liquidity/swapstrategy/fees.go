package swapstrategy

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// computeFeeChargePerSwapStepOutGivenIn returns the total fee charge per swap step given the parameters.
// Assumes swapping for token out given token in.
//
// - hasReachedTarget is the boolean flag indicating whether the sqrtPriceTarget has been reached during the swap step.
//   - the sqrtPriceTarget can be one of:
//   - sqrtPriceLimit
//   - nextTickSqrtPrice
//
// - amountIn the amount of token in to be consumed during the swap step
//
// - amountSpecifiedRemaining is the total remaining amount of token in that needs to be consumed to complete the swap.

// - spreadFactor the spread factor to be charged.
//
// If spread factor is negative, it panics.
// If spread factor is 0, returns 0. Otherwise, computes and returns the fee charge per step.
func computeFeeChargePerSwapStepOutGivenIn(hasReachedTarget bool, amountIn, amountSpecifiedRemaining, spreadFactor sdk.Dec) sdk.Dec {
	feeChargeTotal := sdk.ZeroDec()

	if spreadFactor.IsNegative() {
		// This should never happen but is added as a defense-in-depth measure.
		panic(fmt.Errorf("spread factor must be non-negative, was (%s)", spreadFactor))
	}

	if spreadFactor.IsZero() {
		return feeChargeTotal
	}

	if hasReachedTarget {
		// This branch implies two options:
		// 1) either sqrtPriceNextTick is reached
		// 2) or sqrtPriceLimit is reached
		// In both cases, we charge the fee on the amount in actually consumed before
		// hitting the target.
		feeChargeTotal = computeFeeChargeFromAmountIn(amountIn, spreadFactor)
	} else {
		// Otherwise, the current tick had enough liquidity to fulfill the swap
		// and we ran out of amount remaining before reaching either the next tick or the limit.
		// As a result, the fee is the difference between
		// the amount needed to fulfill and the actual amount we ended up charging.
		feeChargeTotal = amountSpecifiedRemaining.Sub(amountIn)
	}

	if feeChargeTotal.IsNegative() {
		// This should never happen but is added as a defense-in-depth measure.
		panic(fmt.Errorf("fee charge must be non-negative, was (%s)", feeChargeTotal))
	}

	return feeChargeTotal
}

// computeFeeChargeFromAmountIn returns the fee charge given the amount in and spread factor.
// Computes amountIn * spreadFactor / (1 - spreadFactor) where math operations round up
// at precision end. This is necessary to ensure that the fee charge is always
// rounded in favor of the pool.
func computeFeeChargeFromAmountIn(amountIn sdk.Dec, spreadFactor sdk.Dec) sdk.Dec {
	return amountIn.MulRoundUp(spreadFactor).QuoRoundupMut(sdk.OneDec().SubMut(spreadFactor))
}
