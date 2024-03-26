// Copyright IBM Corp. 2022 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"time"

	st "github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/helpers"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	piOnboardingVolumes   = "pi_onboarding_volumes"
	piAuxiliaryVolumes    = "pi_auxiliary_volumes"
	piAuxiliaryVolumeName = "pi_auxiliary_volume_name"
	piSourceCRN           = "pi_source_crn"
	piDisplayName         = "pi_display_name"
	piDescription         = "pi_description"
)

func ResourceIBMPIVolumeOnboarding() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPIVolumeOnboardingCreate,
		ReadContext:   resourceIBMPIVolumeOnboardingRead,
		DeleteContext: resourceIBMPIVolumeOnboardingDelete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{

// Arguments
			Arg_CloudInstanceID: {
				Description: "Cloud Instance ID - This is the service_instance_id.",
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
			},

			piOnboardingVolumes: {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						piSourceCRN: {
							Description: "CRN of source ServiceBroker instance from where auxiliary volumes need to be onboarded",
							Required:    true,
							Type:        schema.TypeString,
						},
						piAuxiliaryVolumes: {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									piAuxiliaryVolumeName: {
										Description: "Auxiliary volume name at storage host level",
										Required:    true,
										Type:        schema.TypeString,
									},
									piDisplayName: {
										Description: "Display name of auxVolumeName once onboarded,auxVolumeName will be set to display name if not provided.",
										Optional:    true,
										Type:        schema.TypeString,
									},
								},
							},
						},
					},
				},
			},
			piDescription: {
				Computed:    true,
				Description: "Description of the volume onboarding operation",
				Optional:    true,
				Type:        schema.TypeString,
			},

			// Computed Attribute
			Attr_CreateTime: {
				Computed:    true,
				Description: "Indicates the create-time of volume onboarding operation",
				Type:        schema.TypeString,
			},
			Attr_OnboardingId: {
				Computed:    true,
				Description: "Indicates the volume onboarding operation id",
				Type:        schema.TypeString,
			},
			Attr_InputVolumes: {
				Computed:    true,
				Description: "List of volumes requested to be onboarded",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Type:        schema.TypeList,
			},
			Attr_Progress: {
				Computed:    true,
				Description: "Indicates the progress of volume onboarding operation",
				Type:        schema.TypeFloat,
			},
			Attr_ResultsOnboardedVolumes: {
				Computed:    true,
				Description: "List of volumes which are onboarded successfully",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Type:        schema.TypeList,
			},
			Attr_ResultsVolumeOnboardingFailures: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Attr_FailureMessage: {
							Computed:    true,
							Description: "The failure reason for the volumes which have failed to be onboarded",
							Type:        schema.TypeString,
						},
						Attr_Volumes: {
							Computed:    true,
							Description: "List of volumes which have failed to be onboarded",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Type:        schema.TypeList,
						},
					}},
			},
			Attr_Status: {
				Computed:    true,
				Description: "Indicates the status of volume onboarding operation",
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceIBMPIVolumeOnboardingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}
	cloudInstanceID := d.Get(helpers.PICloudInstanceId).(string)
	client := st.NewIBMPIVolumeOnboardingClient(ctx, sess, cloudInstanceID)

	vol, err := expandCreateVolumeOnboarding(d.Get(piOnboardingVolumes).([]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}

	body := &models.VolumeOnboardingCreate{
		Volumes: vol,
	}

	if v, ok := d.GetOk(piDescription); ok {
		body.Description = v.(string)
	}

	resOnboarding, err := client.CreateVolumeOnboarding(body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s", cloudInstanceID, resOnboarding.ID))

	return resourceIBMPIVolumeOnboardingRead(ctx, d, meta)
}

func resourceIBMPIVolumeOnboardingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID, onboardingID, err := splitID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	client := st.NewIBMPIVolumeOnboardingClient(ctx, sess, cloudInstanceID)

	onboardingData, err := client.Get(onboardingID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("onboarding_id", *onboardingData.ID)
	d.Set("create_time", onboardingData.CreationTimestamp.String())
	d.Set(piDescription, onboardingData.Description)
	d.Set("input_volumes", onboardingData.InputVolumes)
	d.Set("progress", onboardingData.Progress)
	d.Set("status", onboardingData.Status)
	d.Set("results_onboarded_volumes", onboardingData.Results.OnboardedVolumes)
	d.Set("results_volume_onboarding_failures", flattenVolumeOnboardingFailures(onboardingData.Results.VolumeOnboardingFailures))
	return nil
}

func resourceIBMPIVolumeOnboardingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// There is no delete or unset concept for instance action
	d.SetId("")
	return nil
}

// expandCreateVolumeOnboarding expands create volume onboarding resource
func expandCreateVolumeOnboarding(data []interface{}) ([]*models.AuxiliaryVolumesForOnboarding, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("[ERROR] no pi_onboarding_volumes received")
	}

	auxVolForOnboarding := make([]*models.AuxiliaryVolumesForOnboarding, 0)

	for _, d := range data {
		resource := d.(map[string]interface{})

		var crn string
		auxVolumes := make([]interface{}, 0)

		if v, ok := resource["pi_source_crn"]; ok && v != "" {
			crn = resource["pi_source_crn"].(string)
		}

		if v, ok := resource["pi_auxiliary_volumes"]; ok && len(v.([]interface{})) != 0 {
			auxVolumes = resource["pi_auxiliary_volumes"].([]interface{})
		}

		auxVolForOnboarding = append(auxVolForOnboarding, &models.AuxiliaryVolumesForOnboarding{
			SourceCRN:        &crn,
			AuxiliaryVolumes: expandAuxiliaryVolumeForOnboarding(auxVolumes),
		})

	}

	return auxVolForOnboarding, nil
}

func expandAuxiliaryVolumeForOnboarding(data []interface{}) []*models.AuxiliaryVolumeForOnboarding {
	auxVolumeForOnboarding := make([]*models.AuxiliaryVolumeForOnboarding, 0)

	for _, d := range data {
		var auxVolumeName, displayName string
		resource := d.(map[string]interface{})

		if v, ok := resource["pi_auxiliary_volume_name"]; ok && v != "" {
			auxVolumeName = resource["pi_auxiliary_volume_name"].(string)
		}

		if v, ok := resource["pi_display_name"]; ok && v != "" {
			displayName = resource["pi_display_name"].(string)
		}

		auxVolumeForOnboarding = append(auxVolumeForOnboarding, &models.AuxiliaryVolumeForOnboarding{
			AuxVolumeName: &auxVolumeName,
			Name:          displayName,
		})
	}

	return auxVolumeForOnboarding
}
