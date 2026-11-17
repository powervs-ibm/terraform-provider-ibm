// Copyright IBM Corp. 2026 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func DataSourceIBMPIInstanceNetworkAttachments() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIBMPIInstanceNetworkAttachmentsRead,
		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Description:  "The GUID of the service instance associated with an account.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_InstanceID: {
				Description:  "The unique identifier or ID of the instance.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},

			// Attributes
			Attr_NetworkAttachments: {
				Computed:    true,
				Description: "List of network attachments associated with this instance.",
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Attr_Name: {
							Computed:    true,
							Description: "The name of the network attachment.",
							Type:        schema.TypeString,
						},
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
				},
			},
		},
	}
}

func dataSourceIBMPIInstanceNetworkAttachmentsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "(Data) ibm_pi_instance_network_attachments", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	instanceID := d.Get(Arg_InstanceID).(string)

	client := instance.NewIBMPIInstanceClient(ctx, sess, cloudInstanceID)
	attachmentsWrap, err := client.GetAllNetworkAttachments(instanceID)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("GetAll network attachments failed: %s", err.Error()), "(Data) ibm_pi_instance_network_attachments", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	d.SetId(fmt.Sprintf("%s/%s", cloudInstanceID, instanceID))
	d.Set(Attr_NetworkAttachments, flattenPvmInstanceNetworkAttachments(attachmentsWrap.NetworkAttachments))

	return nil
}

func flattenPvmInstanceNetworkAttachments(list []*models.PVMInstanceNetworkAttachment) []map[string]any {
	attachments := make([]map[string]any, len(list))
	for i, a := range list {
		p := map[string]any{
			Attr_Name:                flex.StringValue(a.Name),
			Attr_NetworkAttachmentID: flex.StringValue(a.NetworkAttachmentID),
			Attr_Type:                flex.StringValue(a.Type),
		}
		if vni := a.VpcVirtualNetworkInterface; vni != nil {
			vniMap := map[string]any{
				Attr_CRN:        flex.StringValue(vni.Crn),
				Attr_IPAddress:  flex.StringValue(vni.IPAddress),
				Attr_MacAddress: flex.StringValue(vni.MacAddress),
				Attr_Name:       flex.StringValue(vni.Name),
			}
			if vni.Bandwidth != nil {
				vniMap[Attr_Bandwidth] = int(*vni.Bandwidth)
			}
			p[Attr_VPCVirtualNetworkInterface] = []map[string]any{vniMap}
		}
		attachments[i] = p
	}
	return attachments
}
