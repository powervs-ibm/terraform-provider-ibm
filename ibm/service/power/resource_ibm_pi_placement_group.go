// Copyright IBM Corp. 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	models "github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func ResourceIBMPIPlacementGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPIPlacementGroupCreate,
		ReadContext:   resourceIBMPIPlacementGroupRead,
		UpdateContext: resourceIBMPIPlacementGroupUpdate,
		DeleteContext: resourceIBMPIPlacementGroupDelete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{

			Arg_PlacementGroupName: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the placement group",
			},

			Attr_SPPPlacementGroupPolicy: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.ValidateAllowedStringValues([]string{"affinity", "anti-affinity"}),
				Description:  "Policy of the placement group",
			},

			Arg_CloudInstanceID: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "PI cloud instance ID",
			},

			Attr_SPPPlacementGroupMembers: {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Server IDs that are the placement group members",
			},

			Attr_SPPPlacementGroupID: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "PI placement group ID",
			},
		},
	}
}

func resourceIBMPIPlacementGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	name := d.Get(Attr_SPPPlacementGroupNamePlacementGroupName).(string)
	policy := d.Get(Attr_SPPPlacementGroupPolicy).(string)
	client := instance.NewIBMPIPlacementGroupClient(ctx, sess, cloudInstanceID)
	body := &models.PlacementGroupCreate{
		Name:   &name,
		Policy: &policy,
	}

	response, err := client.Create(body)
	if err != nil || response == nil {
		return diag.FromErr(fmt.Errorf("error creating the shared processor pool: %s", err))
	}

	log.Printf("Printing the placement group %+v", &response)

	d.SetId(fmt.Sprintf("%s/%s", cloudInstanceID, *response.ID))
	return resourceIBMPIPlacementGroupRead(ctx, d, meta)
}

func resourceIBMPIPlacementGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}
	parts, err := flex.IdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := parts[0]
	client := instance.NewIBMPIPlacementGroupClient(ctx, sess, cloudInstanceID)

	response, err := client.Get(parts[1])
	if err != nil {
		log.Printf("[DEBUG]  err %s", err)
		return diag.FromErr(err)
	}

	d.Set(Attr_SPPPlacementGroupName, response.Name)
	d.Set(Attr_SPPPlacementGroupID, response.ID)
	d.Set(Attr_SPPPlacementGroupPolicy, response.Policy)
	d.Set(Attr_SPPPlacementGroupMembers, response.Members)

	return nil

}

func resourceIBMPIPlacementGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceIBMPIPlacementGroupRead(ctx, d, meta)
}

func resourceIBMPIPlacementGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}
	parts, err := flex.IdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	cloudInstanceID := parts[0]
	client := instance.NewIBMPIPlacementGroupClient(ctx, sess, cloudInstanceID)
	err = client.Delete(parts[1])

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
