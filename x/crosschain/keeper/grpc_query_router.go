package keeper

import (
	"context"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/functionx/fx-core/x/crosschain/types"
)

var _ types.QueryServer = RouterKeeper{}

func (k RouterKeeper) getQueryServerByChainName(chainName string) (types.QueryServer, error) {
	if !k.router.HasRoute(chainName) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "chain name not found:"+chainName)
	}
	return k.router.GetRoute(chainName).QueryServer, nil
}

// Params queries the params of the bsc module
func (k RouterKeeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.Params(c, req)
	}
}

// CurrentOracleSet queries the CurrentOracleSet of the bsc module
func (k RouterKeeper) CurrentOracleSet(c context.Context, req *types.QueryCurrentOracleSetRequest) (*types.QueryCurrentOracleSetResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.CurrentOracleSet(c, req)
	}
}

// OracleSetRequest queries the OracleSetRequest of the bsc module
func (k RouterKeeper) OracleSetRequest(c context.Context, req *types.QueryOracleSetRequestRequest) (*types.QueryOracleSetRequestResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.OracleSetRequest(c, req)
	}
}

// OracleSetConfirm queries the OracleSetConfirm of the bsc module
func (k RouterKeeper) OracleSetConfirm(c context.Context, req *types.QueryOracleSetConfirmRequest) (*types.QueryOracleSetConfirmResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.OracleSetConfirm(c, req)
	}
}

// OracleSetConfirmsByNonce queries the OracleSetConfirmsByNonce of the bsc module
func (k RouterKeeper) OracleSetConfirmsByNonce(c context.Context, req *types.QueryOracleSetConfirmsByNonceRequest) (*types.QueryOracleSetConfirmsByNonceResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.OracleSetConfirmsByNonce(c, req)
	}
}

// LastOracleSetRequests queries the LastOracleSetRequests of the bsc module
func (k RouterKeeper) LastOracleSetRequests(c context.Context, req *types.QueryLastOracleSetRequestsRequest) (*types.QueryLastOracleSetRequestsResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.LastOracleSetRequests(c, req)
	}
}

// LastPendingOracleSetRequestByAddr queries the LastPendingOracleSetRequestByAddr of the bsc module
func (k RouterKeeper) LastPendingOracleSetRequestByAddr(c context.Context, req *types.QueryLastPendingOracleSetRequestByAddrRequest) (*types.QueryLastPendingOracleSetRequestByAddrResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.LastPendingOracleSetRequestByAddr(c, req)
	}
}

// BatchFees queries the batch fees from unbatched pool
func (k RouterKeeper) BatchFees(c context.Context, req *types.QueryBatchFeeRequest) (*types.QueryBatchFeeResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.BatchFees(c, req)
	}
}

// LastPendingBatchRequestByAddr queries the LastPendingBatchRequestByAddr of the bsc module
func (k RouterKeeper) LastPendingBatchRequestByAddr(c context.Context, req *types.QueryLastPendingBatchRequestByAddrRequest) (*types.QueryLastPendingBatchRequestByAddrResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.LastPendingBatchRequestByAddr(c, req)
	}
}

// OutgoingTxBatches queries the OutgoingTxBatches of the bsc module
func (k RouterKeeper) OutgoingTxBatches(c context.Context, req *types.QueryOutgoingTxBatchesRequest) (*types.QueryOutgoingTxBatchesResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.OutgoingTxBatches(c, req)
	}
}

// BatchRequestByNonce queries the BatchRequestByNonce of the bsc module
func (k RouterKeeper) BatchRequestByNonce(c context.Context, req *types.QueryBatchRequestByNonceRequest) (*types.QueryBatchRequestByNonceResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.BatchRequestByNonce(c, req)
	}
}

func (k RouterKeeper) BatchConfirm(c context.Context, req *types.QueryBatchConfirmRequest) (*types.QueryBatchConfirmResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.BatchConfirm(c, req)
	}
}

// BatchConfirms returns the batch confirmations by nonce and token contract
func (k RouterKeeper) BatchConfirms(c context.Context, req *types.QueryBatchConfirmsRequest) (*types.QueryBatchConfirmsResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.BatchConfirms(c, req)
	}
}

// LastEventNonceByAddr returns the last event nonce for the given validator address, this allows eth oracles to figure out where they left off
func (k RouterKeeper) LastEventNonceByAddr(c context.Context, req *types.QueryLastEventNonceByAddrRequest) (*types.QueryLastEventNonceByAddrResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.LastEventNonceByAddr(c, req)
	}
}

func (k RouterKeeper) DenomToToken(c context.Context, req *types.QueryDenomToTokenRequest) (*types.QueryDenomToTokenResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.DenomToToken(c, req)
	}
}

func (k RouterKeeper) TokenToDenom(c context.Context, req *types.QueryTokenToDenomRequest) (*types.QueryTokenToDenomResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.TokenToDenom(c, req)
	}
}

func (k RouterKeeper) GetOracleByAddr(c context.Context, req *types.QueryOracleByAddrRequest) (*types.QueryOracleResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.GetOracleByAddr(c, req)
	}
}

func (k RouterKeeper) GetOracleByOrchestrator(c context.Context, req *types.QueryOracleByOrchestratorRequest) (*types.QueryOracleResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.GetOracleByOrchestrator(c, req)
	}
}

func (k RouterKeeper) GetOracleByExternalAddr(c context.Context, req *types.QueryOracleByExternalAddrRequest) (*types.QueryOracleResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.GetOracleByExternalAddr(c, req)
	}
}

func (k RouterKeeper) GetPendingSendToExternal(c context.Context, req *types.QueryPendingSendToExternalRequest) (*types.QueryPendingSendToExternalResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.GetPendingSendToExternal(c, req)
	}
}

func (k RouterKeeper) GetIbcSequenceHeightByChannel(c context.Context, req *types.QueryIbcSequenceHeightRequest) (*types.QueryIbcSequenceHeightResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.GetIbcSequenceHeightByChannel(c, req)
	}
}

func (k RouterKeeper) LastObservedBlockHeight(c context.Context, req *types.QueryLastObservedBlockHeightRequest) (*types.QueryLastObservedBlockHeightResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.LastObservedBlockHeight(c, req)
	}
}

func (k RouterKeeper) LastEventBlockHeightByAddr(c context.Context, req *types.QueryLastEventBlockHeightByAddrRequest) (*types.QueryLastEventBlockHeightByAddrResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.LastEventBlockHeightByAddr(c, req)
	}
}

func (k RouterKeeper) Oracles(c context.Context, req *types.QueryOraclesRequest) (*types.QueryOraclesResponse, error) {
	if queryServer, err := k.getQueryServerByChainName(req.ChainName); err != nil {
		return nil, err
	} else {
		return queryServer.Oracles(c, req)
	}
}
