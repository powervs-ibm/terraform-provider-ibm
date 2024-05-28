// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/client/p_cloud_images"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceIBMPICapture() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPICaptureCreate,
		ReadContext:   resourceIBMPICaptureRead,
		DeleteContext: resourceIBMPICaptureDelete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(75 * time.Minute),
			Delete: schema.DefaultTimeout(50 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CaptureCloudStorageAccessKey: {
				Description: "Name of Cloud Storage Access Key",
				ForceNew:    true,
				Optional:    true,
				Sensitive:   true,
				Type:        schema.TypeString,
			},
			Arg_CaptureCloudStorageRegion: {
				Description: "List of Regions to use",
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeString,
			},
			Arg_CaptureCloudStorageSecretKey: {
				Description: "Name of the Cloud Storage Secret Key",
				ForceNew:    true,
				Optional:    true,
				Sensitive:   true,
				Type:        schema.TypeString,
			},
			Arg_CaptureDestination: {
				Description:  "Destination for the deployable image",
				ForceNew:     true,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validate.ValidateAllowedStringValues([]string{ImageCatalog, CloudStorage, Both}),
			},
			Arg_CaptureName: {
				Description:  "Name of the capture to create. Note : this must be unique",
				ForceNew:     true,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_CaptureStorageImagePath: {
				Description: "Cloud Storage Image Path (bucket-name [/folder/../..])",
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeString,
			},
			Arg_CaptureVolumeIDs: {
				Description:      "List of Data volume IDs",
				DiffSuppressFunc: flex.ApplyOnce,
				Elem:             &schema.Schema{Type: schema.TypeString},
				ForceNew:         true,
				Optional:         true,
				Set:              schema.HashString,
				Type:             schema.TypeSet,
			},
			Arg_CloudInstanceID: {
				Description:  "The GUID of the service instance associated with an account.",
				ForceNew:     true,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_InstanceName: {
				Description:  "Instance Name of the Power VM",
				ForceNew:     true,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},

			//  Attribute
			Attr_ImageID: {
				Computed:    true,
				Description: "The image id of the capture instance.",
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceIBMPICaptureCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get(Arg_InstanceName).(string)
	capturename := d.Get(Arg_CaptureName).(string)
	capturedestination := d.Get(Arg_CaptureDestination).(string)
	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)

	client := instance.NewIBMPIInstanceClient(context.Background(), sess, cloudInstanceID)

	captureBody := &models.PVMInstanceCapture{
		CaptureDestination: &capturedestination,
		CaptureName:        &capturename,
	}
	if capturedestination != ImageCatalog {
		if v, ok := d.GetOk(Arg_CaptureCloudStorageRegion); ok {
			captureBody.CloudStorageRegion = v.(string)
		} else {
			return diag.Errorf("%s is required when capture destination is %s", Arg_CaptureCloudStorageRegion, capturedestination)
		}
		if v, ok := d.GetOk(Arg_CaptureCloudStorageAccessKey); ok {
			captureBody.CloudStorageAccessKey = v.(string)
		} else {
			return diag.Errorf("%s is required when capture destination is %s ", Arg_CaptureCloudStorageAccessKey, capturedestination)
		}
		if v, ok := d.GetOk(Arg_CaptureStorageImagePath); ok {
			captureBody.CloudStorageImagePath = v.(string)
		} else {
			return diag.Errorf("%s is required when capture destination is %s ", Arg_CaptureStorageImagePath, capturedestination)
		}
		if v, ok := d.GetOk(Arg_CaptureCloudStorageSecretKey); ok {
			captureBody.CloudStorageSecretKey = v.(string)
		} else {
			return diag.Errorf("%s is required when capture destination is %s ", Arg_CaptureCloudStorageSecretKey, capturedestination)
		}
	}

	if v, ok := d.GetOk(Arg_CaptureVolumeIDs); ok {
		volids := flex.ExpandStringList((v.(*schema.Set)).List())
		if len(volids) > 0 {
			captureBody.CaptureVolumeIDs = volids
		}
	}

	captureResponse, err := client.CaptureInstanceToImageCatalogV2(name, captureBody)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", cloudInstanceID, capturename, capturedestination))
	jobClient := instance.NewIBMPIJobClient(ctx, sess, cloudInstanceID)
	_, err = waitForIBMPIJobCompleted(ctx, jobClient, *captureResponse.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceIBMPICaptureRead(ctx, d, meta)
}

func resourceIBMPICaptureRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}
	parts, err := flex.IdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	cloudInstanceID := parts[0]
	captureID := parts[1]
	capturedestination := parts[2]
	if capturedestination != CloudStorage {
		imageClient := instance.NewIBMPIImageClient(ctx, sess, cloudInstanceID)
		imagedata, err := imageClient.Get(captureID)
		if err != nil {
			uErr := errors.Unwrap(err)
			switch uErr.(type) {
			case *p_cloud_images.PcloudCloudinstancesImagesGetNotFound:
				log.Printf("[DEBUG] image does not exist %v", err)
				d.SetId("")
				return nil
			}
			log.Printf("[DEBUG] get image failed %v", err)
			return diag.FromErr(err)
		}
		imageid := *imagedata.ImageID
		d.Set(Attr_ImageID, imageid)
	}
	d.Set(Arg_CloudInstanceID, cloudInstanceID)
	return nil
}

func resourceIBMPICaptureDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}
	parts, err := flex.IdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	cloudInstanceID := parts[0]
	captureID := parts[1]
	capturedestination := parts[2]
	if capturedestination != CloudStorage {
		imageClient := instance.NewIBMPIImageClient(ctx, sess, cloudInstanceID)
		err = imageClient.Delete(captureID)
		if err != nil {
			uErr := errors.Unwrap(err)
			switch uErr.(type) {
			case *p_cloud_images.PcloudCloudinstancesImagesGetNotFound:
				log.Printf("[DEBUG] image does not exist while deleting %v", err)
				d.SetId("")
				return nil
			}
			log.Printf("[DEBUG] delete image failed %v", err)
			return diag.FromErr(err)
		}
	}
	d.SetId("")
	return nil
}
