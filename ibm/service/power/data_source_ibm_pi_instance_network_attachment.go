// Copyright IBM Corp. 2026 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func DataSourceIBMPIInstanceNetworkAttachment() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIBMPIInstanceNetworkAttachmentRead,
		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Description:  "The GUID of the service instance associated with an account.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_InstanceID: {
				Description:  "The unique identifier of the instance.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_NetworkAttachmentID: {
				Description:  "The unique identifier of the network attachment.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},

			// Attributes
			Attr_Name: {
				Computed:    true,
				Description: "The name of the network attachment.",
				Type:        schema.TypeString,
			},
			Attr_Type: {
				Computed:    true,
				Description: "The type of network that is attached.",
				Type:        schema.TypeString,
			},
			Attr_VPCVirtualNetworkInterface: {
				Computed:    true,
				Description: "The VPC Virtual Network Interface that is attached.",
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Attr_Bandwidth: {
							Computed:    true,
							Description: "Bandwidth in Gbps of the network interface.",
							Type:        schema.TypeInt,
						},
						Attr_CRN: {
							Computed:    true,
							Description: "CRN of the VPC Virtual Network Interface.",
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
						Attr_Name: {
							Computed:    true,
							Description: "The name for the network interface.",
							Type:        schema.TypeString,
						},
					},
				},
			},
		},
	}
}

func dataSourceIBMPIInstanceNetworkAttachmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "(Data) ibm_pi_instance_network_attachment", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	instanceID := d.Get(Arg_InstanceID).(string)
	networkAttachmentID := d.Get(Arg_NetworkAttachmentID).(string)

	client := instance.NewIBMPIInstanceClient(ctx, sess, cloudInstanceID)
	attachment, err := client.GetNetworkAttachment(instanceID, networkAttachmentID)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("Get network attachment failed: %s", err.Error()), "(Data) ibm_pi_instance_network_attachment", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", cloudInstanceID, instanceID, networkAttachmentID))
	d.Set(Attr_Name, attachment.Name)
	d.Set(Attr_Type, attachment.Type)
	if vni := attachment.VpcVirtualNetworkInterface; vni != nil {
		vniMap := map[string]any{
			Attr_CRN:        flex.StringValue(vni.Crn),
			Attr_IPAddress:  flex.StringValue(vni.IPAddress),
			Attr_MacAddress: flex.StringValue(vni.MacAddress),
			Attr_Name:       flex.StringValue(vni.Name),
		}
		if vni.Bandwidth != nil {
			vniMap[Attr_Bandwidth] = int(*vni.Bandwidth)
		}
		d.Set(Attr_VPCVirtualNetworkInterface, []map[string]any{vniMap})
	}

	return nil
}
