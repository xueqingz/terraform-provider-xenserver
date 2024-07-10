package xenserver

import (
	"context"
	"errors"
	"strconv"
	"xenapi"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type vbdResourceModel struct {
	VDI      types.String `tfsdk:"vdi_uuid"`
	VBD      types.String `tfsdk:"vbd_ref"`
	Mode     types.String `tfsdk:"mode"`
	Bootable types.Bool   `tfsdk:"bootable"`
}

var vbdResourceModelAttrTypes = map[string]attr.Type{
	"vdi_uuid": types.StringType,
	"vbd_ref":  types.StringType,
	"mode":     types.StringType,
	"bootable": types.BoolType,
}

func VBDSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"vdi_uuid": schema.StringAttribute{
			MarkdownDescription: "VDI UUID to attach to VBD",
			Required:            true,
		},
		"vbd_ref": schema.StringAttribute{
			MarkdownDescription: "VBD Reference",
			Computed:            true,
		},
		"bootable": schema.BoolAttribute{
			MarkdownDescription: "Set VBD as bootable, Default: false",
			Optional:            true,
			Computed:            true,
		},
		"mode": schema.StringAttribute{
			MarkdownDescription: "The mode the VBD should be mounted with, Default: RW",
			Optional:            true,
			Computed:            true,
		},
	}
}

func createVBD(ctx context.Context, vbd vbdResourceModel, vmRef xenapi.VMRef, session *xenapi.Session) (error) {
	vdiRef, err := xenapi.VDI.GetByUUID(session, vbd.VDI.ValueString())
	if err != nil {
		return errors.New(err.Error())
	}

	userDevices, err := xenapi.VM.GetAllowedVBDDevices(session, vmRef)
	if err != nil {
		return errors.New(err.Error())
	}

	if len(userDevices) == 0 {
		return errors.New("unable to find available vbd devices to attach to vm " + string(vmRef))
	}

	bootable := false
	if !vbd.Bootable.IsUnknown() {
		bootable = vbd.Bootable.ValueBool()
	}
	tflog.Debug(ctx, "+++++++++++++VBD created with bootable" + strconv.FormatBool(bootable))

	mode := "RW"
	if !vbd.Mode.IsUnknown() {
		mode = vbd.Mode.ValueString()
	}
	tflog.Debug(ctx, "+++++++++++++VBD created with mode: "+string(mode))

	vbdRecord := xenapi.VBDRecord{
		VM:         vmRef,
		VDI:        vdiRef,
		Type:       "Disk",
		Mode:       xenapi.VbdMode(mode),
		Bootable:   bootable,
		Empty:      false,
		Userdevice: userDevices[0],
	}

	vbdRef, err := xenapi.VBD.Create(session, vbdRecord)
	if err != nil {
		return errors.New(err.Error())
	}

	tflog.Debug(ctx, "+++++++++++++VBD created with ref: "+string(vbdRef))

	// plug VBDs if VM is running
	vmPowerState, err := xenapi.VM.GetPowerState(session, vmRef)
	if err != nil {
		return errors.New(err.Error())
	}

	if vmPowerState == xenapi.VMPowerStateRunning {
		err = xenapi.VBD.Plug(session, vbdRef)
		if err != nil {
			return errors.New(err.Error())
		}
	}


	return nil
}

func createVBDs(ctx context.Context, data vmResourceModel, vmRef xenapi.VMRef, session *xenapi.Session) error {
	elements := make([]vbdResourceModel, 0, len(data.HardDrive.Elements()))
	diags := data.HardDrive.ElementsAs(ctx, &elements, false)
	if diags.HasError() {
		return errors.New("unable to get HardDrive elements")
	}
	for _, vbd := range elements {
		err := createVBD(ctx, vbd, vmRef, session)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateVBDs(ctx context.Context, plan vmResourceModel, state vmResourceModel, vmRef xenapi.VMRef, session *xenapi.Session) error {
	// Get VBDs from plan and state
	planVBDs := make([]vbdResourceModel, 0, len(plan.HardDrive.Elements()))
	diags := plan.HardDrive.ElementsAs(ctx, &planVBDs, false)
	if diags.HasError() {
		return errors.New("unable to get VBDs in plan data")
	}

	stateVBDs := make([]vbdResourceModel, 0, len(state.HardDrive.Elements()))
	diags = state.HardDrive.ElementsAs(ctx, &stateVBDs, false)
	if diags.HasError() {
		return errors.New("unable to get VBDs in state data")
	}

	var err error
	planVDIsMap := make(map[string]vbdResourceModel)
	for _, vbd := range planVBDs {
		planVDIsMap[vbd.VDI.ValueString()] = vbd
	}

	stateVDIsMap := make(map[string]vbdResourceModel)
	for _, vbd := range stateVBDs {
		stateVDIsMap[vbd.VDI.ValueString()] = vbd
	}

	// Create VBDs that are in plan but not in state, Update VBDs if already exists and attributes changed
	for vdiUUID, planVBD := range planVDIsMap {
		stateVBD, ok := stateVDIsMap[vdiUUID]
		if !ok {
			tflog.Debug(ctx, "---> Create VBD for VDI: "+vdiUUID+" <---")
			err = createVBD(ctx, planVBD, vmRef, session)
			if err != nil {
				return err
			}
		} else {
			// Update VBD if attributes changed
			if (planVBD.Bootable.IsUnknown() && planVBD.Mode.ValueString() != "RW") || (!planVBD.Bootable.IsUnknown() && planVBD.Mode != stateVBD.Mode) {
				err = xenapi.VBD.SetMode(session, xenapi.VBDRef(stateVBD.VBD.ValueString()), xenapi.VbdMode(planVBD.Mode.ValueString()))
				if err != nil {
					return errors.New(err.Error())
				}
			}

			if (planVBD.Bootable.IsUnknown() && stateVBD.Bootable.ValueBool() != false) || (!planVBD.Bootable.IsUnknown() && planVBD.Bootable != stateVBD.Bootable) {
				err = xenapi.VBD.SetBootable(session, xenapi.VBDRef(stateVBD.VBD.ValueString()), planVBD.Bootable.ValueBool())
				if err != nil {
					return errors.New(err.Error())
				}
			}
		}
	}

	// Destroy VBDs that are not in plan
	for vdiUUID, stateVBD := range stateVDIsMap {
		if _, ok := planVDIsMap[vdiUUID]; !ok {
			tflog.Debug(ctx, "---> Destroy VBD:	"+stateVBD.VBD.String())
			err = xenapi.VBD.Destroy(session, xenapi.VBDRef(stateVBD.VBD.ValueString()))
			if err != nil {
				return errors.New(err.Error())
			}
		}
	}

	return nil
}
