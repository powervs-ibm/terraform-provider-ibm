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
	"github.com/IBM-Cloud/power-go-client/helpers"
	"github.com/IBM-Cloud/power-go-client/power/client/p_cloud_images"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const cloudStorageDestination string = "cloud-storage"
const imageCatalogDestination string = "image-catalog"

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

			Arg_CloudInstanceID: {
				Description: "The GUID of the service instance associated with an account",
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
			},

			Arg_InstanceCaptureDestination: {
				Description:  "Destination for the deployable image",
				ForceNew:     true,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validate.ValidateAllowedStringValues([]string{"image-catalog", "cloud-storage", "both"}),
			},

			Arg_InstanceCaptureVolumeIds: {
				Description:      "List of Data volume IDs to include in the captured PVMInstance",
				DiffSuppressFunc: flex.ApplyOnce,
				Elem:             &schema.Schema{Type: schema.TypeString},
				ForceNew:         true,
				Optional:         true,
				Set:              schema.HashString,
				Type:             schema.TypeSet,
			},

			Arg_InstanceCaptureCloudStorageRegion: {
				Description: "Cloud Storage Region",
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeString,
			},

			Arg_InstanceCaptureCloudStorageAccessKey: {
				Description: "Cloud Storage Access key",
				ForceNew:    true,
				Optional:    true,
				Sensitive:   true,
				Type:        schema.TypeString,
			},

			Arg_InstanceCaptureCloudStorageSecretKey: {
				Description: "Cloud Storage Secret key",
				ForceNew:    true,
				Optional:    true,
				Sensitive:   true,
				Type:        schema.TypeString,
			},

			Arg_InstanceCaptureCloudStorageImagePath: {
				Description: "Cloud Storage Image Path (bucket-name [/folder/../..])",
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeString,
			},

			Arg_InstanceCaptureName: {
				Description: "Name of the deployable image created for the captured PVMInstance",
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
			},

			Arg_InstanceName: {
				Description: "The name of the instance",
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
			},
			// Computed Attribute
			Attr_ImageID: {
				Computed:    true,
				Description: "The image id of the capture instance. The ID is composed of <pi_cloud_instance_id>/<pi_capture_name>/<pi_capture_destination>.",
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
	capturename := d.Get(Arg_InstanceCaptureName).(string)
	capturedestination := d.Get(Arg_InstanceCaptureDestination).(string)
	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)

	client := instance.NewIBMPIInstanceClient(context.Background(), sess, cloudInstanceID)

	captureBody := &models.PVMInstanceCapture{
		CaptureDestination: &capturedestination,
		CaptureName:        &capturename,
	}
	if capturedestination != imageCatalogDestination {
		if v, ok := d.GetOk(Arg_InstanceCaptureCloudStorageRegion); ok {
			captureBody.CloudStorageRegion = v.(string)
		} else {
			return diag.Errorf("%s is required when capture destination is %s", helpers.PIInstanceCaptureCloudStorageRegion, capturedestination)
		}
		if v, ok := d.GetOk(Arg_InstanceCaptureCloudStorageAccessKey); ok {
			captureBody.CloudStorageAccessKey = v.(string)
		} else {
			return diag.Errorf("%s is required when capture destination is %s ", helpers.PIInstanceCaptureCloudStorageAccessKey, capturedestination)
		}
		if v, ok := d.GetOk(Arg_InstanceCaptureCloudStorageImagePath); ok {
			captureBody.CloudStorageImagePath = v.(string)
		} else {
			return diag.Errorf("%s is required when capture destination is %s ", helpers.PIInstanceCaptureCloudStorageImagePath, capturedestination)
		}
		if v, ok := d.GetOk(Arg_InstanceCaptureCloudStorageSecretKey); ok {
			captureBody.CloudStorageSecretKey = v.(string)
		} else {
			return diag.Errorf("%s is required when capture destination is %s ", helpers.PIInstanceCaptureCloudStorageSecretKey, capturedestination)
		}
	}

	if v, ok := d.GetOk(Arg_InstanceCaptureVolumeIds); ok {
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
	if capturedestination != cloudStorageDestination {
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
		d.Set("image_id", imageid)
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
	if capturedestination != cloudStorageDestination {
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
