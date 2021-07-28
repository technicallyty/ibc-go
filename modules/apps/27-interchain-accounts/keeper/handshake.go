package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	"github.com/cosmos/ibc-go/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/modules/core/05-port/types"
	host "github.com/cosmos/ibc-go/modules/core/24-host"
)

// OnChanOpenInit performs basic validation of channel initialization.
// The channel order must be ORDERED, the counterparty port identifier
// must be the host chain representation as defined in the types package,
// the channel version must be equal to the version in the types package,
// there must not be an active channel for the specfied port identifier,
// and the interchain accounts module must be able to claim the channel
// capability.
func (k Keeper) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) error {
	fmt.Print(portID, channelID)
	if order != channeltypes.ORDERED {
		return sdkerrors.Wrapf(channeltypes.ErrInvalidChannelOrdering, "invalid channel ordering: %s, expected %s", order.String(), channeltypes.ORDERED.String())
	}
	if counterparty.PortId != types.PortID {
		return sdkerrors.Wrapf(porttypes.ErrInvalidPort, "counterparty port-id must be '%s', (%s != %s)", types.PortID, counterparty.PortId, types.PortID)
	}
	if version != types.Version {
		return sdkerrors.Wrapf(channeltypes.ErrInvalidChannelVersion, "channel version must be '%s' (%s != %s)", types.Version, version, types.Version)
	}
	channelID, found := k.GetActiveChannel(ctx, portID)
	if found {
		return sdkerrors.Wrapf(porttypes.ErrInvalidPort, "existing active channel (%s) for portID (%s)", channelID, portID)
	}

	// Claim channel capability passed back by IBC module
	if err := k.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
		return sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, err.Error())
	}

	return nil
}

// register account (if it doesn't exist)
// check if counterpary version is the same
// TODO: remove ics27-1 hardcoded
func (k Keeper) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version,
	counterpartyVersion string,
) error {
	if order != channeltypes.ORDERED {
		return sdkerrors.Wrapf(channeltypes.ErrInvalidChannelOrdering, "invalid channel ordering: %s, expected %s", order.String(), channeltypes.ORDERED.String())
	}

	// TODO: Check counterparty version
	// if counterpartyVersion != types.Version {
	// 	return sdkerrors.Wrapf(porttypes.ErrInvalidPort, "invalid counterparty version: %s, expected %s", counterpartyVersion, "ics20-1")
	// }

	// Claim channel capability passed back by IBC module
	if err := k.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
		return sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, err.Error())
	}

	// Register interchain account if it does not already exist
	_, _ = k.RegisterInterchainAccount(ctx, counterparty.PortId)
	return nil
}

func (k Keeper) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyVersion string,
) error {
	k.SetActiveChannel(ctx, portID, channelID)

	return nil
}

// Set active channel
func (k Keeper) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {

	return nil
}

// May want to use these for re-opening a channel when it is closed
//// OnChanCloseInit implements the IBCModule interface
//func (am AppModule) OnChanCloseInit(
//	ctx sdk.Context,
//	portID,
//	channelID string,
//) error {
//	// Disallow user-initiated channel closing for transfer channels
//	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "user cannot close channel")
//}

//// OnChanCloseConfirm implements the IBCModule interface
//func (am AppModule) OnChanCloseConfirm(
//	ctx sdk.Context,
//	portID,
//	channelID string,
//) error {
//	return nil
//}