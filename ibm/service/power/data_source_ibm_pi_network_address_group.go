// Copyright IBM Corp. 2024 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
)

func DataSourceIBMPINetworkAddressGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIBMPINetworkAddressGroupRead,

		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Description:  "The GUID of the service instance associated with an account.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},

			Arg_NetworkAddressGroupID: {
				Description: "Network Address Group ID.",
				Required:    true,
				Type:        schema.TypeString,
			},
			// Attributes
			Attr_CRN: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Network Address Group's crn.",
			},

			Attr_Members: {
				Computed:    true,
				Description: "The list of IP addresses in CIDR notation (for example 192.168.66.2/32) in the Network Address Group.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Attr_CIDR: {
							Computed:    true,
							Description: "The IP addresses in CIDR notation for example 192.168.1.5/32.",
							Type:        schema.TypeString,
						},
						Attr_ID: {
							Computed:    true,
							Description: "The id of the Network Address Group member IP addresses.",
							Type:        schema.TypeString,
						},
					},
				},
				Type: schema.TypeList,
			},
			Attr_Name: {
				Computed:    true,
				Description: "The name of the Network Address Group.",
				Type:        schema.TypeString,
			},
			Attr_UserTags: {
				Computed:    true,
				Description: "The user tags associated with this resource.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Type: schema.TypeList,
			},
		},
	}
}

func dataSourceIBMPINetworkAddressGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	nagID := d.Get(Arg_NetworkAddressGroupID).(string)
	nagC := instance.NewIBMPINetworkAddressGroupClient(ctx, sess, cloudInstanceID)
	networkAddressGroup, err := nagC.Get(nagID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*networkAddressGroup.ID)

	d.Set(Attr_CRN, networkAddressGroup.Crn)

	members := []map[string]interface{}{}
	if networkAddressGroup.Members != nil {
		for _, mbr := range networkAddressGroup.Members {
			member := memberToMap(mbr)
			members = append(members, member)
		}
	}
	d.Set(Attr_Members, members)
	d.Set(Attr_Name, networkAddressGroup.Name)
	if len(networkAddressGroup.UserTags) > 0 {
		d.Set(Attr_UserTags, networkAddressGroup.UserTags)
	}

	return nil
}
