// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package catalogmanagement

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/IBM/platform-services-go-sdk/catalogmanagementv1"
)

func DataSourceIBMCmOfferingInstance() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIBMCmOfferingInstanceRead,

		Schema: map[string]*schema.Schema{
			"instance_identifier": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID for this instance",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "url reference to this object.",
			},
			"crn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "platform CRN for this instance.",
			},
			"_rev": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Cloudant Revision for this instance",
			},
			"label": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "the label for this instance.",
			},
			"catalog_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Catalog ID this instance was created from.",
			},
			"offering_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Offering ID this instance was created from.",
			},
			"kind_format": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "the format this instance has (helm, operator, ova...).",
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version this instance was installed from (not version id).",
			},
			"cluster_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Cluster ID.",
			},
			"cluster_region": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Cluster region (e.g., us-south).",
			},
			"cluster_namespaces": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of target namespaces to install into.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cluster_all_namespaces": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "designate to install into all namespaces.",
			},
			"schematics_workspace_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "id of the schematics workspace, for offerings installed through schematics",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "id of the resource group",
			},
			"install_plan": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "install plan for the subscription of the operator- can be either Automatic or Manual. Required for operator bundles",
			},
			"channel": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "channel to target for the operator subscription. Required for operator bundles",
			},
			"plan_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "id of the plan",
			},
			"parent_crn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "CRN of parent instance",
			},
		},
	}
}

func dataSourceIBMCmOfferingInstanceRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	catalogManagementClient, err := meta.(conns.ClientSession).CatalogManagementV1()
	if err != nil {
		return diag.FromErr(err)
	}

	getOfferingInstanceOptions := &catalogmanagementv1.GetOfferingInstanceOptions{}

	getOfferingInstanceOptions.SetInstanceIdentifier(d.Get("instance_identifier").(string))

	offeringInstance, response, err := catalogManagementClient.GetOfferingInstanceWithContext(context, getOfferingInstanceOptions)
	if err != nil {
		log.Printf("[DEBUG] GetOfferingInstanceWithContext failed %s\n%s", err, response)
		return diag.FromErr(err)
	}

	d.SetId(*offeringInstance.ID)

	if err = d.Set("url", offeringInstance.URL); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting url: %s", err))
	}
	if err = d.Set("crn", offeringInstance.CRN); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting crn: %s", err))
	}
	if err = d.Set("_rev", offeringInstance.Rev); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting _rev: %s", err))
	}
	if err = d.Set("label", offeringInstance.Label); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting label: %s", err))
	}
	if err = d.Set("catalog_id", offeringInstance.CatalogID); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting catalog_id: %s", err))
	}
	if err = d.Set("offering_id", offeringInstance.OfferingID); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting offering_id: %s", err))
	}
	if err = d.Set("kind_format", offeringInstance.KindFormat); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting kind_format: %s", err))
	}
	if err = d.Set("version", offeringInstance.Version); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting version: %s", err))
	}
	if err = d.Set("cluster_id", offeringInstance.ClusterID); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting cluster_id: %s", err))
	}
	if err = d.Set("cluster_region", offeringInstance.ClusterRegion); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting cluster_region: %s", err))
	}
	if err = d.Set("cluster_namespaces", offeringInstance.ClusterNamespaces); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting cluster_namespaces: %s", err))
	}
	if err = d.Set("cluster_all_namespaces", offeringInstance.ClusterAllNamespaces); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting cluster_all_namespaces: %s", err))
	}
	if err = d.Set("schematics_workspace_id", offeringInstance.SchematicsWorkspaceID); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting schematics_workspace_id: %s", err))
	}
	if err = d.Set("resource_group_id", offeringInstance.ResourceGroupID); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting resource_group_id: %s", err))
	}
	if err = d.Set("install_plan", offeringInstance.InstallPlan); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting install_plan: %s", err))
	}
	if err = d.Set("channel", offeringInstance.Channel); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting channel: %s", err))
	}
	if err = d.Set("plan_id", offeringInstance.PlanID); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting plan_id: %s", err))
	}
	if err = d.Set("parent_crn", offeringInstance.ParentCRN); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting parent_crn: %s", err))
	}

	return nil
}
