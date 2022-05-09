package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	fxtypes "github.com/functionx/fx-core/types"
	"github.com/functionx/fx-core/x/crosschain/keeper"
	crosstypes "github.com/functionx/fx-core/x/crosschain/types"
)

func TestGetRequestBatchBaseFee(t *testing.T) {
	fn := func(i sdk.Int) *sdk.Int {
		return &i
	}
	testCases := []struct {
		testName      string
		height        int64
		baseFee       *sdk.Int
		expectBaseFee sdk.Int
		expectPass    bool
		err           error
		errReason     string
	}{
		{
			testName:      "No Support - nil baseFee",
			height:        1,
			baseFee:       fn(sdk.ZeroInt()),
			expectBaseFee: sdk.ZeroInt(),
			expectPass:    true,
			err:           nil,
		},
		{
			testName:      "No Support - has baseFee",
			height:        1,
			baseFee:       fn(sdk.NewInt(1000)),
			expectBaseFee: sdk.ZeroInt(),
			expectPass:    true,
			err:           nil,
		},
		{
			testName:      "Support - no baseFee",
			height:        fxtypes.RequestBatchBaseFeeBlock(),
			baseFee:       fn(sdk.ZeroInt()),
			expectBaseFee: sdk.ZeroInt(),
			expectPass:    true,
			err:           nil,
		},
		{
			testName:      "Support - negative baseFee",
			height:        fxtypes.RequestBatchBaseFeeBlock(),
			baseFee:       fn(sdk.NewInt(-1)),
			expectBaseFee: sdk.ZeroInt(),
			expectPass:    false,
			err:           crosstypes.ErrInvalidRequestBatchBaseFee,
		},
		{
			testName:      "Support - has baseFee",
			height:        fxtypes.RequestBatchBaseFeeBlock(),
			baseFee:       fn(sdk.NewInt(101)),
			expectBaseFee: sdk.NewInt(101),
			expectPass:    true,
			err:           nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			actual, err := keeper.GetRequestBatchBaseFee(testCase.height, testCase.baseFee)
			if testCase.expectPass {
				require.NoError(t, err)
				require.True(t, testCase.expectBaseFee.Equal(actual))
				return
			}

			require.NotNil(t, err)
			require.Equal(t, err, testCase.err)
		})
	}
}
