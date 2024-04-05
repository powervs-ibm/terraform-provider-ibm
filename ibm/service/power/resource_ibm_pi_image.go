// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/errors"
	"github.com/IBM-Cloud/power-go-client/power/client/p_cloud_images"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceIBMPIImage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPIImageCreate,
		ReadContext:   resourceIBMPIImageRead,
		DeleteContext: resourceIBMPIImageDelete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_AffinityInstance: {
				ConflictsWith: []string{Arg_AffinityVolume},
				Description:   "PVM Instance (ID or Name) to base storage affinity policy against; required if requesting 'affinity' and 'pi_affinity_volume' is not provided.",
				ForceNew:      true,
				Optional:      true,
				Type:          schema.TypeString,
			},
			Arg_AffinityPolicy: {
				Description:  "Affinity policy for image; ignored if 'pi_image_storage_pool' provided; for policy affinity requires one of 'pi_affinity_instance' or 'pi_affinity_volume' to be specified; for policy anti-affinity requires one of 'pi_anti_affinity_instances' or 'pi_anti_affinity_volumes' to be specified; Allowable values: 'affinity', 'anti-affinity'.",
				ForceNew:     true,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validate.ValidateAllowedStringValues([]string{"affinity", "anti-affinity"}),
			},
			Arg_AffinityVolume: {
				ConflictsWith: []string{Arg_AffinityInstance},
				Description:   "Volume (ID or Name) to base storage affinity policy against; required if requesting 'affinity' and 'pi_affinity_instance' is not provided.",
				ForceNew:      true,
				Optional:      true,
				Type:          schema.TypeString,
			},
			Arg_AntiAffinityInstances: {
				ConflictsWith: []string{Arg_AntiAffinityVolumes},
				Description:   "List of pvmInstances to base storage anti-affinity policy against; required if requesting 'anti-affinity' and 'pi_anti_affinity_volumes' is not provided.",
				Elem:          &schema.Schema{Type: schema.TypeString},
				ForceNew:      true,
				Optional:      true,
				Type:          schema.TypeList,
			},
			Arg_AntiAffinityVolumes: {
				ConflictsWith: []string{Arg_AntiAffinityInstances},
				Description:   "List of volumes to base storage anti-affinity policy against; required if requesting 'anti-affinity' and 'pi_anti_affinity_instances' is not provided.",
				Elem:          &schema.Schema{Type: schema.TypeString},
				ForceNew:      true,
				Optional:      true,
				Type:          schema.TypeList,
			},
			Arg_CloudInstanceID: {
				Description: "The GUID of the service instance associated with an account.",
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
			},
			Arg_ImageAccessKey: {
				Description:  "Cloud Object Storage access key; required for buckets with private access.",
				ForceNew:     true,
				Optional:     true,
				RequiredWith: []string{Arg_ImageSecretKey},
				Sensitive:    true,
				Type:         schema.TypeString,
			},
			Arg_ImageBucketAccess: {
				ConflictsWith: []string{Arg_ImageID},
				Description:   "Indicates if the bucket has public or private access. The default value is 'public'.",
				Default:       "public",
				ForceNew:      true,
				Optional:      true,
				Type:          schema.TypeString,
				ValidateFunc:  validate.ValidateAllowedStringValues([]string{"public", "private"}),
			},
			Arg_ImageBucketFileName: {
				ConflictsWith: []string{Arg_ImageID},
				Description:   "Cloud Object Storage image filename.",
				ForceNew:      true,
				Optional:      true,
				RequiredWith:  []string{Arg_ImageBucketName},
				Type:          schema.TypeString,
			},
			Arg_ImageBucketName: {
				ConflictsWith: []string{Arg_ImageID},
				Description:   "Cloud Object Storage bucket name; 'bucket-name[/optional/folder]'.",
				ExactlyOneOf:  []string{Arg_ImageID, Arg_ImageBucketName},
				ForceNew:      true,
				Optional:      true,
				RequiredWith:  []string{Arg_ImageBucketRegion, Arg_ImageBucketFileName},
				Type:          schema.TypeString,
			},
			Arg_ImageBucketRegion: {
				ConflictsWith: []string{Arg_ImageID},
				Description:   "Cloud Object Storage region.",
				Optional:      true,
				ForceNew:      true,
				RequiredWith:  []string{Arg_ImageBucketName},
				Type:          schema.TypeString,
			},
			Arg_ImageID: {
				ConflictsWith:    []string{Arg_ImageBucketName},
				Description:      "Image ID of existing source image; required for copy image.",
				DiffSuppressFunc: flex.ApplyOnce,
				ExactlyOneOf:     []string{Arg_ImageID, Arg_ImageBucketName},
				ForceNew:         true,
				Optional:         true,
				Type:             schema.TypeString,
			},
			Arg_ImageName: {
				Description:      "The name of an image.",
				DiffSuppressFunc: flex.ApplyOnce,
				ForceNew:         true,
				Required:         true,
				Type:             schema.TypeString,
			},
			Arg_ImageSecretKey: {
				Description:  "Cloud Object Storage secret key; required for buckets with private access.",
				ForceNew:     true,
				Optional:     true,
				RequiredWith: []string{Arg_ImageAccessKey},
				Sensitive:    true,
				Type:         schema.TypeString,
			},
			Arg_ImageStoragePool: {
				Description: "Storage pool where the image will be loaded, if provided then `pi_affinity_policy` will be ignored. Used only when importing an image from cloud storage.",
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeString,
			},
			Arg_ImageStorageType: {
				Description: "Type of storage; If not provided the storage type will default to 'tier3'. Used only when importing an image from cloud storage.",
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeString,
			},

			// Attributes
			Attr_ImageID: {
				Computed:    true,
				Description: "The unique identifier of an image.",
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceIBMPIImageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		log.Printf("Failed to get the session")
		return diag.FromErr(err)
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	imageName := d.Get(Arg_ImageName).(string)

	client := instance.NewIBMPIImageClient(ctx, sess, cloudInstanceID)
	// image copy
	if v, ok := d.GetOk(Arg_ImageID); ok {
		imageid := v.(string)
		source := "root-project"
		var body = &models.CreateImage{
			ImageName: imageName,
			ImageID:   imageid,
			Source:    &source,
		}
		imageResponse, err := client.Create(body)
		if err != nil {
			return diag.FromErr(err)
		}

		IBMPIImageID := imageResponse.ImageID
		d.SetId(fmt.Sprintf("%s/%s", cloudInstanceID, *IBMPIImageID))

		_, err = isWaitForIBMPIImageAvailable(ctx, client, *IBMPIImageID, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			log.Printf("[DEBUG]  err %s", err)
			return diag.FromErr(err)
		}
	}

	// COS image import
	if v, ok := d.GetOk(Arg_ImageBucketName); ok {
		bucketName := v.(string)
		bucketImageFileName := d.Get(Arg_ImageBucketFileName).(string)
		bucketRegion := d.Get(Arg_ImageBucketRegion).(string)
		bucketAccess := d.Get(Arg_ImageBucketAccess).(string)

		body := &models.CreateCosImageImportJob{
			ImageName:     &imageName,
			BucketName:    &bucketName,
			BucketAccess:  &bucketAccess,
			ImageFilename: &bucketImageFileName,
			Region:        &bucketRegion,
		}

		if v, ok := d.GetOk(Arg_ImageAccessKey); ok {
			body.AccessKey = v.(string)
		}
		if v, ok := d.GetOk(Arg_ImageSecretKey); ok {
			body.SecretKey = v.(string)
		}

		if v, ok := d.GetOk(Arg_ImageStorageType); ok {
			body.StorageType = v.(string)
		}
		if v, ok := d.GetOk(Arg_ImageStoragePool); ok {
			body.StoragePool = v.(string)
		}
		if ap, ok := d.GetOk(Arg_AffinityPolicy); ok {
			policy := ap.(string)
			affinity := &models.StorageAffinity{
				AffinityPolicy: &policy,
			}

			if policy == "affinity" {
				if ai, ok := d.GetOk(Arg_AffinityInstance); ok {
					afins := ai.(string)
					affinity.AffinityPVMInstance = &afins
				}
				if av, ok := d.GetOk(Arg_AffinityVolume); ok {
					afvol := av.(string)
					affinity.AffinityVolume = &afvol
				}
			} else {
				if ais, ok := d.GetOk(Arg_AntiAffinityInstances); ok {
					afinss := flex.ExpandStringList(ais.([]interface{}))
					affinity.AntiAffinityPVMInstances = afinss
				}
				if avs, ok := d.GetOk(Arg_AntiAffinityVolumes); ok {
					afvols := flex.ExpandStringList(avs.([]interface{}))
					affinity.AntiAffinityVolumes = afvols
				}
			}
			body.StorageAffinity = affinity
		}
		imageResponse, err := client.CreateCosImage(body)
		if err != nil {
			return diag.FromErr(err)
		}

		jobClient := instance.NewIBMPIJobClient(ctx, sess, cloudInstanceID)
		_, err = waitForIBMPIJobCompleted(ctx, jobClient, *imageResponse.ID, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}

		// Once the job is completed find by name
		image, err := client.Get(imageName)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(fmt.Sprintf("%s/%s", cloudInstanceID, *image.ImageID))
	}

	return resourceIBMPIImageRead(ctx, d, meta)
}

func resourceIBMPIImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID, imageID, err := splitID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	imageC := instance.NewIBMPIImageClient(ctx, sess, cloudInstanceID)
	imagedata, err := imageC.Get(imageID)
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
	d.Set(Arg_CloudInstanceID, cloudInstanceID)

	return nil
}

func resourceIBMPIImageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID, imageID, err := splitID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	imageC := instance.NewIBMPIImageClient(ctx, sess, cloudInstanceID)
	err = imageC.Delete(imageID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func isWaitForIBMPIImageAvailable(ctx context.Context, client *instance.IBMPIImageClient, id string, timeout time.Duration) (interface{}, error) {
	log.Printf("Waiting for Power Image (%s) to be available.", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{Status_Retry, Status_Queued},
		Target:     []string{Status_Active},
		Refresh:    isIBMPIImageRefreshFunc(ctx, client, id),
		Timeout:    timeout,
		Delay:      20 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	return stateConf.WaitForStateContext(ctx)
}

func isIBMPIImageRefreshFunc(ctx context.Context, client *instance.IBMPIImageClient, id string) resource.StateRefreshFunc {
	log.Printf("Calling the isIBMPIImageRefreshFunc Refresh Function....")
	return func() (interface{}, string, error) {
		image, err := client.Get(id)
		if err != nil {
			return nil, "", err
		}

		if image.State == Status_Active {
			return image, Status_Active, nil
		}

		return image, Status_Queued, nil
	}
}

func waitForIBMPIJobCompleted(ctx context.Context, client *instance.IBMPIJobClient, jobID string, timeout time.Duration) (interface{}, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{Status_Queued, Status_ReadyForProcessing, Status_InProgress, Status_Running, Status_Waiting},
		Target:  []string{Status_Completed, Status_Failed},
		Refresh: func() (interface{}, string, error) {
			job, err := client.Get(jobID)
			if err != nil {
				log.Printf("[DEBUG] get job failed %v", err)
				return nil, "", fmt.Errorf(errors.GetJobOperationFailed, jobID, err)
			}
			if job == nil || job.Status == nil {
				log.Printf("[DEBUG] get job failed with empty response")
				return nil, "", fmt.Errorf("failed to get job status for job id %s", jobID)
			}
			if *job.Status.State == Status_Failed {
				log.Printf("[DEBUG] job status failed with message: %v", job.Status.Message)
				return nil, Status_Failed, fmt.Errorf("job status failed for job id %s with message: %v", jobID, job.Status.Message)
			}
			return job, *job.Status.State, nil
		},
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}
	return stateConf.WaitForStateContext(ctx)
}
