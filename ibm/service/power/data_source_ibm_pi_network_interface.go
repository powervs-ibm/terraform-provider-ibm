// Copyright IBM Corp. 2024 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
)

func DataSourceIBMPINetworkInterface() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIBMPINetworkInterfaceRead,

		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Description:  "The GUID of the service instance associated with an account.",
				ForceNew:     true,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_NetworkID: {
				Description:  "Network ID.",
				ForceNew:     true,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_NetworkInterfaceID: {
				Description:  "Network Interface ID.",
				ForceNew:     true,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			// Attributes
			Attr_CRN: {
				Computed:    true,
				Description: "The Network Interface's crn.",
				Type:        schema.TypeString,
			},
			Attr_IPAddress: {
				Computed:    true,
				Description: "The ip address of this Network Interface.",
				Type:        schema.TypeString,
			},
			Attr_MacAddress: {
				Computed:    true,
				Description: "The mac address of the Network Interface.",
				Type:        schema.TypeString,
			},
			Attr_Name: {
				Computed:    true,
				Description: "Name of the Network Interface (not unique or indexable).",
				Type:        schema.TypeString,
			},
			Attr_NetworkSecurityGroupID: {
				Computed:    true,
				Description: "ID of the Network Security Group the network interface will be added to.",
				Type:        schema.TypeString,
			},
			Attr_PVMInstance: {
				Computed:    true,
				Description: "The attached pvm-instance to this Network Interface.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Attr_Href: {
							Computed:    true,
							Description: "Link to pvm-instance resource.",
							Type:        schema.TypeString,
						},
						Attr_InstanceID: {
							Computed:    true,
							Description: "The attahed instance ID.",
							Type:        schema.TypeString,
						},
					},
				},
				Type: schema.TypeList,
			},
			Attr_Status: {
				Computed:    true,
				Description: "The status of the network address group.",
				Type:        schema.TypeString,
			},
		},
	}
}

func dataSourceIBMPINetworkInterfaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	networkID := d.Get(Arg_NetworkID).(string)
	networkInterfaceID := d.Get(Arg_NetworkInterfaceID).(string)
	networkC := instance.NewIBMPINetworkClient(ctx, sess, cloudInstanceID)
	networkInterface, err := networkC.GetNetworkInterface(networkID, networkInterfaceID)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(*networkInterface.ID)
	d.Set(Attr_CRN, networkInterface.Crn)
	d.Set(Attr_IPAddress, networkInterface.IPAddress)
	d.Set(Attr_MacAddress, networkInterface.MacAddress)
	d.Set(Attr_Name, networkInterface.Name)
	d.Set(Attr_NetworkSecurityGroupID, networkInterface.NetworkSecurityGroupID)

	if networkInterface.Instance != nil {
		pvmInstance := []map[string]interface{}{}
		instanceMap := pvmInstanceToMap(networkInterface.Instance)
		pvmInstance = append(pvmInstance, instanceMap)
		d.Set(Attr_PVMInstance, pvmInstance)
	}
	d.Set(Attr_Status, networkInterface.Status)

	return nil
}

func pvmInstanceToMap(pvm *models.NetworkInterfaceInstance) map[string]interface{} {
	instanceMap := make(map[string]interface{})
	if pvm.Href != "" {
		instanceMap[Attr_Href] = pvm.Href
	}
	if pvm.InstanceID != "" {
		instanceMap[Attr_InstanceID] = pvm.InstanceID
	}
	return instanceMap
}
