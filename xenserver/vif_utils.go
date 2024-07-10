package xenserver

import (
	"context"
	"errors"

	"xenapi"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type vifResourceModel struct {
	Network     types.String `tfsdk:"network_uuid"`
	VIF         types.String `tfsdk:"vif_ref"`
	MTU         types.Int64  `tfsdk:"mtu"`
	MAC         types.String `tfsdk:"mac"`
	OtherConfig types.Map    `tfsdk:"other_config"`
}

var vifResourceModelAttrTypes = map[string]attr.Type{
	"network_uuid": types.StringType,
	"vif_ref":      types.StringType,
	"mtu":          types.Int64Type,
	"mac":          types.StringType,
	"other_config": types.MapType{ElemType: types.StringType},
}

func VIFSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"network_uuid": schema.StringAttribute{
			MarkdownDescription: "Network UUID to attach to VIF",
			Required:            true,
		},
		"vif_ref": schema.StringAttribute{
			MarkdownDescription: "VIF Reference",
			Computed:            true,
		},
		"mtu": schema.Int64Attribute{
			MarkdownDescription: "MTU in octets, default: 1500",
			Optional:            true,
			Computed:            true,
		},
		"mac": schema.StringAttribute{
			MarkdownDescription: "MAC address of the VIF, if not provided, XenServer will generate a random MAC address.",
			Optional:            true,
			Computed:            true,
		},
		"other_config": schema.MapAttribute{
			MarkdownDescription: "The additional configuration, default to be {}",
			ElementType:         types.StringType,
			Optional:            true,
			Computed:            true,
		},
	}
}

func createVIF(ctx context.Context, vifData vifResourceModel, vmRef xenapi.VMRef, session *xenapi.Session) error {
	networkRef, err := xenapi.Network.GetByUUID(session, vifData.Network.ValueString())
	if err != nil {
		return errors.New(err.Error())
	}

	userDevices, err := xenapi.VM.GetAllowedVIFDevices(session, vmRef)
	if err != nil {
		return errors.New(err.Error())
	}

	if len(userDevices) == 0 {
		return errors.New("No available vif devices to attach to vm " + string(vmRef))
	}

	otherConfig := make(map[string]string)
	if !vifData.OtherConfig.IsUnknown() {
		diags := vifData.OtherConfig.ElementsAs(ctx, &otherConfig, false)
		if diags.HasError() {
			return errors.New("unable to get VIF other config")
		}
	}

	mac := ""
	macAuto := false
	if !vifData.MAC.IsUnknown() && vifData.MAC.ValueString() != "" {
		mac = vifData.MAC.ValueString()
		macAuto = true
	}

	mtu := 1500 
	if !vifData.MTU.IsUnknown() {
		err = checkMTU(int(vifData.MTU.ValueInt64()))
		if err != nil {
			return err
		}
	}

	vifRecord := xenapi.VIFRecord{
		VM:               vmRef,
		Network:          networkRef,
		MAC:              mac,
		MTU:              mtu,
		OtherConfig:      otherConfig,
		Device:           userDevices[0],
		LockingMode:      xenapi.VifLockingModeNetworkDefault,
		MACAutogenerated: macAuto,
	}

	vifRef, err := xenapi.VIF.Create(session, vifRecord)
	if err != nil {
		return errors.New(err.Error())
	}

	tflog.Debug(ctx, "+++++++++++++VIF created with ref: "+string(vifRef))

	vmPowerState, err := xenapi.VM.GetPowerState(session, vmRef)
	if err != nil {
		return errors.New(err.Error())
	}

	if vmPowerState == xenapi.VMPowerStateRunning {
		err = xenapi.VIF.Plug(session, vifRef)
		if err != nil {
			return errors.New(err.Error())
		}
	}

	return nil
}

func createVIFs(ctx context.Context, data vmResourceModel, vmRef xenapi.VMRef, session *xenapi.Session) error {
	elements := make([]vifResourceModel, 0, len(data.NetworkInterface.Elements()))
	diags := data.NetworkInterface.ElementsAs(ctx, &elements, false)
	if diags.HasError() {
		return errors.New("unable to get Network Interface elements")
	}

	for _, vifData := range elements {
		err := createVIF(ctx, vifData, vmRef, session)
		if err != nil {
			return err
		}
	}
	return nil
}


// updateVIF updates the VIFs in the VM based on the plan and state, the logic is similar to updateVBDs
func updateVIFs(ctx context.Context, plan vmResourceModel, state vmResourceModel, vmRef xenapi.VMRef, session *xenapi.Session) error {
	// Get VIFs from plan and state
	planVIFs := make([]vifResourceModel, 0, len(plan.NetworkInterface.Elements()))
	diags := plan.NetworkInterface.ElementsAs(ctx, &planVIFs, false)
	if diags.HasError() {
		return errors.New("unable to get VIFs in plan data")
	}

	stateVIFs := make([]vifResourceModel, 0, len(state.NetworkInterface.Elements()))
	diags = state.NetworkInterface.ElementsAs(ctx, &stateVIFs, false)
	if diags.HasError() {
		return errors.New("unable to get VIFs in state data")
	}

	var err error
	planVIFsMap := make(map[string]vifResourceModel)
	for _, vif := range planVIFs {
		planVIFsMap[vif.Network.ValueString()] = vif
	}

	stateVIFsMap := make(map[string]vifResourceModel)
	for _, vif := range stateVIFs {
		stateVIFsMap[vif.Network.ValueString()] = vif
	}

	// Create VIFs that are in plan but not in state, Update VIFs if already exists and attributes changed
	for networkUUID, planVIF := range planVIFsMap {
		stateVIF, ok := stateVIFsMap[networkUUID]
		if !ok {
			tflog.Debug(ctx, "---> Create VIF for Network: "+networkUUID+" <---")
			err = createVIF(ctx, planVIF, vmRef, session)
			if err != nil {
				return err
			}
		} else {
			if planVIF.MTU != stateVIF.MTU {
				return errors.New(`"network_interface.mtu" doesn't expected to be updated`)
			}

			if planVIF.MAC != stateVIF.MAC {
				return errors.New(`"network_interface.mac" doesn't expected to be updated`)
			}

			if !planVIF.OtherConfig.Equal(stateVIF.OtherConfig) {
				tflog.Debug(ctx, "---> Update VIF other config "+stateVIF.VIF.String()+" for Network: "+networkUUID+" <---")
				otherConfig := make(map[string]string)
				if !planVIF.OtherConfig.IsUnknown() {
					diags := planVIF.OtherConfig.ElementsAs(ctx, &otherConfig, false)
					if diags.HasError() {
						return errors.New("unable to get VIF other config")
					}
				}

				err = xenapi.VIF.SetOtherConfig(session, xenapi.VIFRef(stateVIF.VIF.ValueString()), otherConfig)
				if err != nil {
					return errors.New(err.Error())
				}
			}
		}
	}

	// Destroy VIFs that are not in plan
	for networkUUID, stateVIF := range stateVIFsMap {
		if _, ok := planVIFsMap[networkUUID]; !ok {
			tflog.Debug(ctx, "---> Destroy VIF:	"+stateVIF.VIF.String())
			err = xenapi.VIF.Destroy(session, xenapi.VIFRef(stateVIF.VIF.ValueString()))
			if err != nil {
				return errors.New(err.Error())
			}
		}
	}

	return nil
}
