// Copyright IBM Corp. 2026 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
)

func ResourceIBMPIInstanceNetworkAttachment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPIInstanceNetworkAttachmentCreate,
		ReadContext:   resourceIBMPIInstanceNetworkAttachmentRead,
		UpdateContext: resourceIBMPIInstanceNetworkAttachmentUpdate,
		DeleteContext: resourceIBMPIInstanceNetworkAttachmentDelete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Description:  "The GUID of the service instance associated with an account.",
				ForceNew:     true,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_InstanceID: {
				Description:  "The unique identifier of the instance to attach the network attachment to.",
				ForceNew:     true,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_Name: {
				Computed:    true,
				Description: "The name of this network attachment. Names must be unique within the PVM instance the network attachment resides in. If unspecified, the name will be a hyphenated list of randomly-selected words.",
				Optional:    true,
				Type:        schema.TypeString,
			},
			Arg_VPCVirtualNetworkInterface: {
				Description: "The VPC Virtual Network Interface to attach.",
				ForceNew:    true,
				MaxItems:    1,
				MinItems:    1,
				Required:    true,
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Attr_Bandwidth: {
							Computed:     true,
							Description:  "Bandwidth in Gbps for the network interface. Valid range is 1 to 100. Default is 1.",
							Optional:     true,
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntBetween(1, 100),
						},
						Attr_CRN: {
							Description: "CRN of the VPC Virtual Network Interface to attach.",
							ForceNew:    true,
							Required:    true,
							Type:        schema.TypeString,
						},
						Attr_IPAddress: {
							Computed:    true,
							Description: "The IP address of the network interface.",
							Type:        schema.TypeString,
						},
						Attr_MacAddress: {
							Computed:    true,
							Description: "The MAC address of the network interface.",
							Type:        schema.TypeString,
						},
					},
				},
			},

			// Attributes
			Attr_NetworkAttachmentID: {
				Computed:    true,
				Description: "ID of the network attachment.",
				Type:        schema.TypeString,
			},
			Attr_Type: {
				Computed:    true,
				Description: "The type of network that is attached.",
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceIBMPIInstanceNetworkAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_network_attachment", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	instanceID := d.Get(Arg_InstanceID).(string)

	vni := d.Get(Arg_VPCVirtualNetworkInterface).([]any)[0].(map[string]any)
	crn := vni[Attr_CRN].(string)
	body := &models.PVMInstanceAddNetworkAttachment{
		VpcVirtualNetworkInterface: &models.PVMInstanceAddVPCVirtualNetworkInterface{
			Bandwidth: int64(vni[Attr_Bandwidth].(int)),
			Crn:       &crn,
		},
	}
	if v, ok := d.GetOk(Arg_Name); ok {
		body.Name = v.(string)
	}

	client := instance.NewIBMPIInstanceClient(ctx, sess, cloudInstanceID)
	attachment, err := client.CreateNetworkAttachment(instanceID, body)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("Create network attachment failed: %s", err.Error()), "ibm_pi_instance_network_attachment", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", cloudInstanceID, instanceID, *attachment.NetworkAttachmentID))

	return resourceIBMPIInstanceNetworkAttachmentRead(ctx, d, meta)
}

func resourceIBMPIInstanceNetworkAttachmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_network_attachment", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	parts, err := flex.IdParts(d.Id())
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IdParts failed: %s", err.Error()), "ibm_pi_instance_network_attachment", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	cloudInstanceID := parts[0]
	instanceID := parts[1]
	networkAttachmentID := parts[2]

	client := instance.NewIBMPIInstanceClient(ctx, sess, cloudInstanceID)
	attachment, err := client.GetNetworkAttachment(instanceID, networkAttachmentID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), NotFound) {
			log.Printf("[WARNING] network attachment was not found or removed outside of terraform\n")
			d.SetId("")
			return nil
		}
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("Get network attachment failed: %s", err.Error()), "ibm_pi_instance_network_attachment", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	d.Set(Arg_CloudInstanceID, cloudInstanceID)
	d.Set(Arg_InstanceID, instanceID)
	d.Set(Attr_NetworkAttachmentID, networkAttachmentID)
	d.Set(Arg_Name, attachment.Name)
	d.Set(Attr_Type, attachment.Type)
	if vni := attachment.VpcVirtualNetworkInterface; vni != nil {
		vniMap := map[string]any{
			Attr_CRN:        flex.StringValue(vni.Crn),
			Attr_IPAddress:  flex.StringValue(vni.IPAddress),
			Attr_MacAddress: flex.StringValue(vni.MacAddress),
		}
		if vni.Bandwidth != nil {
			vniMap[Attr_Bandwidth] = int(*vni.Bandwidth)
		}
		d.Set(Arg_VPCVirtualNetworkInterface, []map[string]any{vniMap})
	}

	return nil
}

func resourceIBMPIInstanceNetworkAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_network_attachment", "update")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	parts, err := flex.IdParts(d.Id())
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IdParts failed: %s", err.Error()), "ibm_pi_instance_network_attachment", "update")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	cloudInstanceID := parts[0]
	instanceID := parts[1]
	networkAttachmentID := parts[2]

	if d.HasChange(Arg_Name) || d.HasChange(fmt.Sprintf("%s.0.%s", Arg_VPCVirtualNetworkInterface, Attr_Bandwidth)) {
		vni := d.Get(Arg_VPCVirtualNetworkInterface).([]any)[0].(map[string]any)
		body := &models.PVMInstanceUpdateNetworkAttachment{
			Name: d.Get(Arg_Name).(string),
			VpcVirtualNetworkInterface: &models.PVMInstanceUpdateVPCVirtualNetworkInterface{
				Bandwidth: int64(vni[Attr_Bandwidth].(int)),
			},
		}

		client := instance.NewIBMPIInstanceClient(ctx, sess, cloudInstanceID)
		_, err = client.UpdateNetworkAttachment(instanceID, networkAttachmentID, body)
		if err != nil {
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("Update network attachment failed: %s", err.Error()), "ibm_pi_instance_network_attachment", "update")
			log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
		}
	}

	return resourceIBMPIInstanceNetworkAttachmentRead(ctx, d, meta)
}

func resourceIBMPIInstanceNetworkAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_network_attachment", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	parts, err := flex.IdParts(d.Id())
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IdParts failed: %s", err.Error()), "ibm_pi_instance_network_attachment", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	cloudInstanceID := parts[0]
	instanceID := parts[1]
	networkAttachmentID := parts[2]

	client := instance.NewIBMPIInstanceClient(ctx, sess, cloudInstanceID)
	err = client.DeleteNetworkAttachment(instanceID, networkAttachmentID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), NotFound) {
			d.SetId("")
			return nil
		}
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("Delete network attachment failed: %s", err.Error()), "ibm_pi_instance_network_attachment", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	d.SetId("")
	return nil
}
