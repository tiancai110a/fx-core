package gravity

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/functionx/fx-core/app"
	"github.com/functionx/fx-core/x/gravity/keeper"
	"github.com/functionx/fx-core/x/gravity/types"
)

// EndBlocker is called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	params := k.GetParams(ctx)
	slashing(ctx, k, params)
	attestationTally(ctx, k)
	cleanupTimedOutBatches(ctx, k)
	createValsets(ctx, k)
	if ctx.BlockHeight() > app.GravityPruneValsetsAndAttestationBlock() {
		pruneValsets(ctx, k, params)
		pruneAttestations(ctx, k)
	}
}

func createValsets(ctx sdk.Context, k keeper.Keeper) {
	// Auto ValsetRequest Creation.
	// WARNING: do not use k.GetLastObservedValset in this function, it *will* result in losing control of the bridge
	// 1. If there are no valset requests, create a new one.
	// 2. If there is at least one validator who started unbonding in current block. (we persist last unbonded block height in hooks.go)
	//      This will make sure the unbonding validator has to provide an attestation to a new Valset
	//	    that excludes him before he completely Unbonds.  Otherwise he will be slashed
	// 3. If power change between validators of CurrentValset and latest valset request is > 10%
	if currentValset, is := needUpdateValset(ctx, k); is {
		k.SetValsetRequest(ctx, currentValset)
	}
}

func needUpdateValset(ctx sdk.Context, k keeper.Keeper) (*types.Valset, bool) {
	currentValset := k.GetCurrentValset(ctx)
	latestValset := k.GetLatestValset(ctx)
	if latestValset == nil {
		return currentValset, true
	}
	if k.GetLastUnBondingBlockHeight(ctx) == uint64(ctx.BlockHeight()) {
		ctx.Logger().Info("Gravity", "valset change, has validator UnBonding", ctx.BlockHeight())
		return currentValset, true
	}
	params := k.GetParams(ctx)
	if powerDiff, isChange := valSetPowerIsChanger(latestValset.Members, currentValset.Members, params.ValsetUpdatePowerChangePercent); isChange {
		ctx.Logger().Info("Gravity", "valset power change", "", "change threshold", params.ValsetUpdatePowerChangePercent.String(), "powerDiff", powerDiff)
		return currentValset, true
	}
	return currentValset, false
}

func valSetPowerIsChanger(latestValset []*types.BridgeValidator, currentValset []*types.BridgeValidator, powerChangeThreshold sdk.Dec) (float64, bool) {
	powerDiff := types.BridgeValidators(currentValset).PowerDiff(latestValset)
	powerDiffStr := fmt.Sprintf("%.8f", powerDiff)
	powerDiffDec, err := sdk.NewDecFromStr(powerDiffStr)
	if err != nil {
		panic(fmt.Errorf("covert power diff to dec err!!!powerDiff:%v, err:%v", powerDiffStr, err))
	}
	return powerDiff, powerDiffDec.GTE(powerChangeThreshold)
}

// pruneValsets
func pruneValsets(ctx sdk.Context, k keeper.Keeper, params types.Params) {
	// Validator set pruning
	// prune all validator sets with a nonce less than the
	// last observed nonce, they can't be submitted any longer
	//
	// Only prune valsets after the signed valsets window has passed
	// so that slashing can occur the block before we remove them
	lastObserved := k.GetLastObservedValset(ctx)
	currentBlock := uint64(ctx.BlockHeight())
	tooEarly := currentBlock < params.SignedValsetsWindow
	if lastObserved != nil && !tooEarly {
		earliestToPrune := currentBlock - params.SignedValsetsWindow
		sets := k.GetValsets(ctx)
		for _, set := range sets {
			if set.Nonce < lastObserved.Nonce && set.Height < earliestToPrune {
				k.DeleteValset(ctx, set.Nonce)
			}
		}
	}
}
func slashing(ctx sdk.Context, k keeper.Keeper, params types.Params) {
	// Slash validator for not confirming valset requests, batch requests
	ValsetSlashing(ctx, k, params)
	BatchSlashing(ctx, k, params)
}

// Iterate over all attestations currently being voted on in order of nonce and
// "Observe" those who have passed the threshold. Break the loop once we see
// an attestation that has not passed the threshold
func attestationTally(ctx sdk.Context, k keeper.Keeper) {
	attMap := k.GetAttestationMapping(ctx)
	// We make a slice with all the event nonces that are in the attestation mapping
	nonces := make([]uint64, 0, len(attMap))
	for k := range attMap {
		nonces = append(nonces, k)
	}
	// Then we sort it
	sort.Slice(nonces, func(i, j int) bool {
		return nonces[i] < nonces[j]
	})

	// This iterates over all nonces (event nonces) in the attestation mapping. Each value contains
	// a slice with one or more attestations at that event nonce. There can be multiple attestations
	// at one event nonce when validators disagree about what event happened at that nonce.
	for _, nonce := range nonces {
		// This iterates over all attestations at a particular event nonce.
		// They are ordered by when the first attestation at the event nonce was received.
		// This order is not important.
		for _, att := range attMap[nonce] {
			// We check if the event nonce is exactly 1 higher than the last attestation that was
			// observed. If it is not, we just move on to the next nonce. This will skip over all
			// attestations that have already been observed.
			//
			// Once we hit an event nonce that is one higher than the last observed event, we stop
			// skipping over this conditional and start calling tryAttestation (counting votes)
			// Once an attestation at a given event nonce has enough votes and becomes observed,
			// every other attestation at that nonce will be skipped, since the lastObservedEventNonce
			// will be incremented.
			//
			// Then we go to the next event nonce in the attestation mapping, if there is one. This
			// nonce will once again be one higher than the lastObservedEventNonce.
			// If there is an attestation at this event nonce which has enough votes to be observed,
			// we skip the other attestations and move on to the next nonce again.
			// If no attestation becomes observed, when we get to the next nonce, every attestation in
			// it will be skipped. The same will happen for every nonce after that.
			if nonce == k.GetLastObservedEventNonce(ctx)+1 {
				k.TryAttestation(ctx, &att)
			}
		}
	}
}

// cleanupTimedOutBatches deletes batches that have passed their expiration on Ethereum
// keep in mind several things when modifying this function
// A) unlike nonces timeouts are not monotonically increasing, meaning batch 5 can have a later timeout than batch 6
//    this means that we MUST only cleanup a single batch at a time
// B) it is possible for ethereumHeight to be zero if no events have ever occurred, make sure your code accounts for this
// C) When we compute the timeout we do our best to estimate the Ethereum block height at that very second. But what we work with
//    here is the Ethereum block height at the time of the last Deposit or Withdraw to be observed. It's very important we do not
//    project, if we do a slowdown on ethereum could cause a double spend. Instead timeouts will *only* occur after the timeout period
//    AND any deposit or withdraw has occurred to update the Ethereum block height.
func cleanupTimedOutBatches(ctx sdk.Context, k keeper.Keeper) {
	ethereumHeight := k.GetLastObservedEthBlockHeight(ctx).EthBlockHeight
	batches := k.GetOutgoingTxBatches(ctx)
	for _, batch := range batches {
		if batch.BatchTimeout < ethereumHeight {
			_ = k.CancelOutgoingTXBatch(ctx, batch.TokenContract, batch.BatchNonce)
		}
	}
}

func ValsetSlashing(ctx sdk.Context, k keeper.Keeper, params types.Params) {
	logger := k.Logger(ctx)
	maxHeight := uint64(0)

	// don't slash in the beginning before there aren't even SignedValsetsWindow blocks yet
	if uint64(ctx.BlockHeight()) > params.SignedValsetsWindow {
		maxHeight = uint64(ctx.BlockHeight()) - params.SignedValsetsWindow
	}

	unslashedValsets := k.GetUnSlashedValsets(ctx, maxHeight)

	// unslashedValsets are sorted by nonce in ASC order
	// Question: do we need to sort each time? See if this can be epoched
	for _, vs := range unslashedValsets {
		confirms := k.GetValsetConfirms(ctx, vs.Nonce)

		// SLASH BONDED VALIDTORS who didn't attest valset request
		currentBondedSet := k.StakingKeeper.GetBondedValidatorsByPower(ctx)
		for _, val := range currentBondedSet {
			ethAddress, foundEthAddress := k.GetEthAddressByValidator(ctx, val.GetOperator())
			if !foundEthAddress && ctx.BlockHeight() > app.GravityValsetSlashBlock() {
				continue
			}
			consAddr, _ := val.GetConsAddr()
			valSigningInfo, exist := k.SlashingKeeper.GetValidatorSigningInfo(ctx, consAddr)

			//  Slash validator ONLY if he joined after valset is created
			if exist && valSigningInfo.StartHeight < int64(vs.Height) {
				// Check if validator has confirmed valset or not
				found := false
				for _, conf := range confirms {
					if foundEthAddress && conf.EthAddress == ethAddress {
						found = true
						break
					}
				}
				// slash validators for not confirming valsets
				if !found {
					cons, _ := val.GetConsAddr()
					k.StakingKeeper.Slash(ctx, cons, ctx.BlockHeight(), val.ConsensusPower(), params.SlashFractionValset)
					logger.Info("validator set slash bonded validators", "height", ctx.BlockHeight(), "operator", val.OperatorAddress)
					if !val.IsJailed() {
						k.StakingKeeper.Jail(ctx, cons)
						logger.Info("validator set jail bonded validators", "height", ctx.BlockHeight(), "consensus", cons.String())
					}

				}
			}
		}

		// SLASH UNBONDING VALIDATORS who didn't attest valset request
		blockTime := ctx.BlockTime().Add(k.StakingKeeper.GetParams(ctx).UnbondingTime)
		blockHeight := ctx.BlockHeight()
		unbondingValIterator := k.StakingKeeper.ValidatorQueueIterator(ctx, blockTime, blockHeight)
		// All unbonding validators
		for ; unbondingValIterator.Valid(); unbondingValIterator.Next() {
			unbondingValidators := k.GetUnbondingValidators(unbondingValIterator.Value())

			for _, valAddr := range unbondingValidators.Addresses {
				addr, err := sdk.ValAddressFromBech32(valAddr)
				if err != nil {
					panic(err)
				}
				ethAddress, foundValidator := k.GetEthAddressByValidator(ctx, addr)
				if !foundValidator && ctx.BlockHeight() > app.GravityValsetSlashBlock() {
					continue
				}

				validator, _ := k.StakingKeeper.GetValidator(ctx, addr)
				valConsAddr, _ := validator.GetConsAddr()
				valSigningInfo, exist := k.SlashingKeeper.GetValidatorSigningInfo(ctx, valConsAddr)

				// Only slash validators who joined after valset is created and they are unbonding and UNBOND_SLASHING_WINDOW didn't passed
				if exist && valSigningInfo.StartHeight < int64(vs.Nonce) && validator.IsUnbonding() && vs.Nonce < uint64(validator.UnbondingHeight)+params.UnbondSlashingValsetsWindow {
					// Check if validator has confirmed valset or not
					found := false
					for _, conf := range confirms {
						if foundValidator && conf.EthAddress == ethAddress {
							found = true
							break
						}
					}

					// slash validators for not confirming valsets
					if !found {
						k.StakingKeeper.Slash(ctx, valConsAddr, ctx.BlockHeight(), validator.ConsensusPower(), params.SlashFractionValset)
						logger.Info("validator set slash unbonding validators", "height", ctx.BlockHeight(), "operator", validator.OperatorAddress)
						if !validator.IsJailed() {
							k.StakingKeeper.Jail(ctx, valConsAddr)
							logger.Info("validator set jail unbonding validators", "height", ctx.BlockHeight(), "consensus", valConsAddr)
						}
					}
				}
			}
		}
		unbondingValIterator.Close()
		// then we set the latest slashed valset  nonce
		k.SetLastSlashedValsetNonce(ctx, vs.Nonce)
	}
}

func BatchSlashing(ctx sdk.Context, k keeper.Keeper, params types.Params) {
	logger := k.Logger(ctx)
	// We look through the full bonded set (the active set)
	// and we slash users who haven't signed a batch confirmation that is >15hrs in blocks old
	maxHeight := uint64(0)

	// don't slash in the beginning before there aren't even SignedBatchesWindow blocks yet
	if uint64(ctx.BlockHeight()) > params.SignedBatchesWindow {
		maxHeight = uint64(ctx.BlockHeight()) - params.SignedBatchesWindow
	}

	unslashedBatches := k.GetUnSlashedBatches(ctx, maxHeight)
	for _, batch := range unslashedBatches {

		// SLASH BONDED VALIDTORS who didn't attest batch requests
		currentBondedSet := k.StakingKeeper.GetBondedValidatorsByPower(ctx)
		confirms := k.GetBatchConfirmByNonceAndTokenContract(ctx, batch.BatchNonce, batch.TokenContract)
		for _, val := range currentBondedSet {
			ethAddress, foundValidator := k.GetEthAddressByValidator(ctx, val.GetOperator())
			if !foundValidator && ctx.BlockHeight() > app.GravityValsetSlashBlock() {
				continue
			}
			// Don't slash validators who joined after batch is created
			consAddr, _ := val.GetConsAddr()
			valSigningInfo, exist := k.SlashingKeeper.GetValidatorSigningInfo(ctx, consAddr)
			if exist && valSigningInfo.StartHeight > int64(batch.Block) {
				continue
			}

			found := false
			for _, conf := range confirms {
				if foundValidator && conf.EthSigner == ethAddress {
					found = true
					break
				}
			}
			if !found {
				cons, _ := val.GetConsAddr()
				logger.Info("batch slash bonded validator", "height", ctx.BlockHeight(), "operator", val.OperatorAddress, "cons", cons, "slashFraction", params.SlashFractionBatch)
				k.StakingKeeper.Slash(ctx, cons, ctx.BlockHeight(), val.ConsensusPower(), params.SlashFractionBatch)
				if !val.IsJailed() {
					logger.Info("batch jail bonded validators", "height", ctx.BlockHeight(), "consensus", cons.String())
					k.StakingKeeper.Jail(ctx, cons)
				}
			}
		}
		// then we set the latest slashed batch block
		k.SetLastSlashedBatchBlock(ctx, batch.Block)
	}
}

// Iterate over all attestations currently being voted on in order of nonce
// and prune those that are older than the current nonce and no longer have any
// use. This could be combined with create attestation and save some computation
// but (A) pruning keeps the iteration small in the first place and (B) there is
// already enough nuance in the other handler that it's best not to complicate it further
func pruneAttestations(ctx sdk.Context, k keeper.Keeper) {
	attmap := k.GetAttestationMapping(ctx)
	// We make a slice with all the event nonces that are in the attestation mapping
	keys := make([]uint64, 0, len(attmap))
	for k := range attmap {
		keys = append(keys, k)
	}
	// Then we sort it
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	// This iterates over all keys (event nonces) in the attestation mapping. Each value contains
	// a slice with one or more attestations at that event nonce. There can be multiple attestations
	// at one event nonce when validators disagree about what event happened at that nonce.
	for _, nonce := range keys {
		// This iterates over all attestations at a particular event nonce.
		// They are ordered by when the first attestation at the event nonce was received.
		// This order is not important.
		for _, att := range attmap[nonce] {
			// we delete all attestations earlier than the current event nonce
			if nonce < k.GetLastObservedEventNonce(ctx) {
				k.DeleteAttestation(ctx, att)
			}
		}
	}
}
