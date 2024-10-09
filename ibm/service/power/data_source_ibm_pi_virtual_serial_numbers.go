// Copyright IBM Corp. 2024 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"

	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Datasource to list Cloud Connections in a power instance
func DataSourceIBMPIVirtualSerialNumbers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIBMPIVirtualSerialNumbersRead,
		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Description:  "The GUID of the service instance associated with an account.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_PVMInstanceId: {
				Description: "ID of PVM instance to get virtual serial number attached to.",
				Optional:    true,
				Type:        schema.TypeString,
			},

			// Attributes
			// Attr_VirtualSerialNumbers: {
			// 	Computed: true,
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			Attr_Description: {
			// 				Computed:    true,
			// 				Description: "Description of virtual serial number.",
			// 				Type:        schema.TypeString,
			// 			},
			// 			Attr_PVMInstanceID: {
			// 				Computed:    true,
			// 				Description: "ID of PVM instance virtual serial number is attached to.",
			// 				Type:        schema.TypeString,
			// 			},
			// 			Attr_Serial: {
			// 				Computed:    true,
			// 				Description: "Virtual Serial Number.",
			// 				Type:        schema.TypeString,
			// 			},
			// 		},
			// 	},
			// 	Type: schema.TypeList,
			// },
		},
	}
}

func dataSourceIBMPIVirtualSerialNumbersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	_, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.Errorf("Placeholder")
}

func flattenVirtualSerialNumbers() {}
