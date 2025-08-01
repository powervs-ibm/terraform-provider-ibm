// Copyright IBM Corp. 2025 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func DataSourceIBMPINetworkPeer() *schema.Resource {

	return &schema.Resource{
		ReadContext: dataSourceIBMPINetworkPeerRead,

		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Description:  "The GUID of the service instance associated with an account.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_NetworkPeerID: {
				Description:  "Network peer ID.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			// Attributes
			Attr_CustomerASN: {
				Computed:    true,
				Description: "ASN number at customer network side.",
				Type:        schema.TypeInt,
			},
			Attr_CustomerCIDR: {
				Computed:    true,
				Description: "IP address used for configuring customer network interface with network subnet mask.",
				Type:        schema.TypeString,
			},
			Attr_DefaultExportRouteFilter: {
				Computed:    true,
				Description: "Default action for export route filter.",
				Type:        schema.TypeString,
			},
			Attr_DefaultImportRouteFilter: {
				Computed:    true,
				Description: "Default action for import route filter.",
				Type:        schema.TypeString,
			},
			Attr_ExportRouteFilters: {
				Computed:    true,
				Description: "List of export route filters.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Attr_Action: {
							Computed:    true,
							Description: "Action of the filter.",
							Type:        schema.TypeString,
						},
						Attr_Direction: {
							Computed:    true,
							Description: "Direction of the filter.",
							Type:        schema.TypeString,
						},
						Attr_GE: {
							Computed:    true,
							Description: "The minimum matching length of the prefix-set.",
							Type:        schema.TypeInt,
						},
						Attr_Index: {
							Computed:    true,
							Description: "Priority or order of the filter.",
							Type:        schema.TypeInt,
						},
						Attr_LE: {
							Computed:    true,
							Description: "The maximum matching length of the prefix-set.",
							Type:        schema.TypeInt,
						},
						Attr_Prefix: {
							Computed:    true,
							Description: "IP prefix representing an address and mask length of the prefix-set.",
							Type:        schema.TypeString,
						},
						Attr_RouteFilterID: {
							Computed:    true,
							Description: "Route filter ID.",
							Type:        schema.TypeString,
						},
					},
				},
				Type: schema.TypeList,
			},
			Attr_IBMASN: {
				Computed:    true,
				Description: "ASN number at IBM PowerVS side.",
				Type:        schema.TypeInt,
			},
			Attr_IBMCIDR: {
				Computed:    true,
				Description: "IP address used for configuring IBM network interface with network subnet mask.",
				Type:        schema.TypeString,
			},
			Attr_ImportRouteFilters: {
				Computed:    true,
				Description: "List of import route filters.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Attr_Action: {
							Computed:    true,
							Description: "Action of the filter.",
							Type:        schema.TypeString,
						},
						Attr_Direction: {
							Computed:    true,
							Description: "Direction of the filter.",
							Type:        schema.TypeString,
						},
						Attr_GE: {
							Computed:    true,
							Description: "The minimum matching length of the prefix-set.",
							Type:        schema.TypeInt,
						},
						Attr_Index: {
							Computed:    true,
							Description: "Priority or order of the filter.",
							Type:        schema.TypeInt,
						},
						Attr_LE: {
							Computed:    true,
							Description: "The maximum matching length of the prefix-set.",
							Type:        schema.TypeInt,
						},
						Attr_Prefix: {
							Computed:    true,
							Description: "IP prefix representing an address and mask length of the prefix-set.",
							Type:        schema.TypeString,
						},
						Attr_RouteFilterID: {
							Computed:    true,
							Description: "Route filter ID.",
							Type:        schema.TypeString,
						},
					},
				},
				Type: schema.TypeList,
			},
			Attr_Name: {
				Computed:    true,
				Description: "User defined name.",
				Type:        schema.TypeString,
			},
			Attr_PeerInterfaceID: {
				Computed:    true,
				Description: "Peer interface id.",
				Type:        schema.TypeString,
			},
			Attr_Type: {
				Computed:    true,
				Description: "Type of the peer network.",
				Type:        schema.TypeString,
			},
			Attr_VLAN: {
				Computed:    true,
				Description: "A vlan configured at the customer network.",
				Type:        schema.TypeInt,
			},
		},
	}

}

func dataSourceIBMPINetworkPeerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}
	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	networkPeerID := d.Get(Arg_NetworkPeerID).(string)
	networkC := instance.NewIBMPINetworkPeerClient(ctx, sess, cloudInstanceID)
	networkdata, err := networkC.GetNetworkPeer(networkPeerID)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(*networkdata.ID)
	d.Set(Attr_CustomerASN, networkdata.CustomerASN)
	d.Set(Attr_CustomerCIDR, networkdata.CustomerCidr)

	d.Set(Attr_DefaultExportRouteFilter, networkdata.DefaultExportRouteFilter)
	d.Set(Attr_DefaultImportRouteFilter, networkdata.DefaultImportRouteFilter)
	exportRouteFilters := []map[string]interface{}{}
	if networkdata.ExportRouteFilters != nil {
		for _, erp := range networkdata.ExportRouteFilters {
			exportRouteFilter := dataSourceIBMPINetworkPeerRouteFilterToMap(erp)
			exportRouteFilters = append(exportRouteFilters, exportRouteFilter)
		}
	}
	d.Set(Attr_ExportRouteFilters, exportRouteFilters)
	d.Set(Attr_IBMASN, networkdata.IbmASN)
	d.Set(Attr_IBMCIDR, networkdata.IbmCidr)
	importRouteFilters := []map[string]interface{}{}
	if networkdata.ImportRouteFilters != nil {
		for _, irp := range networkdata.ImportRouteFilters {
			importRouteFilter := dataSourceIBMPINetworkPeerRouteFilterToMap(irp)
			importRouteFilters = append(importRouteFilters, importRouteFilter)
		}
	}
	d.Set(Attr_ImportRouteFilters, importRouteFilters)
	if networkdata.Name != nil {
		d.Set(Attr_Name, networkdata.Name)
	}
	d.Set(Attr_PeerInterfaceID, networkdata.PeerInterfaceID)
	if networkdata.Type != nil {
		d.Set(Attr_Type, networkdata.Type)
	}
	d.Set(Attr_VLAN, networkdata.Vlan)

	return nil
}
