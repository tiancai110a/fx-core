package app

import (
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/spf13/cobra"
)

const (
	FlagLogFilter = "log_filter"
)

func InterceptConfigsPreRunHandler(cmd *cobra.Command) error {
	if err := server.InterceptConfigsPreRunHandler(cmd); err != nil {
		return err
	}
	return addCmdLogWrapFilterLogType(cmd)
}

func addCmdLogWrapFilterLogType(cmd *cobra.Command) error {
	filterLogTypes, err := cmd.Flags().GetStringSlice(FlagLogFilter)
	if err != nil {
		return err
	}
	if len(filterLogTypes) <= 0 {
		return nil
	}
	serverCtx := server.GetServerContextFromCmd(cmd)
	if zeroLog, ok := serverCtx.Logger.(server.ZeroLogWrapper); ok {
		serverCtx.Logger = NewFxZeroLogWrapper(zeroLog, filterLogTypes)
	}
	return server.SetCmdServerContext(cmd, serverCtx)
}
