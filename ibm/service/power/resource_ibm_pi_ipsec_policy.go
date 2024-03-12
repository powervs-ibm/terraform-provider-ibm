// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceIBMPIIPSecPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPIIPSecPolicyCreate,
		ReadContext:   resourceIBMPIIPSecPolicyRead,
		UpdateContext: resourceIBMPIIPSecPolicyUpdate,
		DeleteContext: resourceIBMPIIPSecPolicyDelete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "PI cloud instance ID",
			},
			PIVPNPolicyName: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the IPSec Policy",
			},
			PIVPNPolicyDhGroup: {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validate.ValidateAllowedIntValues([]int{1, 2, 5, 14, 19, 20, 24}),
				Description:  "DH group of the IPSec Policy",
			},
			PIVPNPolicyEncryption: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.ValidateAllowedStringValues([]string{"aes-256-cbc", "aes-192-cbc", "aes-128-cbc", "aes-256-gcm", "aes-128-gcm", "3des-cbc"}),
				Description:  "Encryption of the IPSec Policy",
			},
			PIVPNPolicyKeyLifetime: {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validate.ValidateAllowedRangeInt(180, 86400),
				Description:  "Policy key lifetime",
			},
			PIVPNPolicyPFS: {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Perfect Forward Secrecy",
			},

			// Optional Attributes
			PIVPNPolicyAuthentication: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "none",
				ValidateFunc: validate.ValidateAllowedStringValues([]string{"hmac-sha-256-128", "hmac-sha1-96", "none"}),
				Description:  "Authentication for the IPSec Policy",
			},

			//Computed Attributes
			PIPolicyId: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IPSec policy ID",
			},
		},
	}
}

func resourceIBMPIIPSecPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	name := d.Get(PIVPNPolicyName).(string)
	dhGroup := int64(d.Get(PIVPNPolicyDhGroup).(int))
	encryption := d.Get(PIVPNPolicyEncryption).(string)
	pfs := d.Get(PIVPNPolicyPFS).(bool)
	keyLifetime := int64(d.Get(PIVPNPolicyKeyLifetime).(int))
	klt := models.KeyLifetime(keyLifetime)

	body := &models.IPSecPolicyCreate{
		DhGroup:     &dhGroup,
		Encryption:  &encryption,
		KeyLifetime: &klt,
		Name:        &name,
		Pfs:         &pfs,
	}

	if v, ok := d.GetOk(PIVPNPolicyAuthentication); ok {
		body.Authentication = models.IPSECPolicyAuthentication(v.(string))
	}

	client := instance.NewIBMPIVpnPolicyClient(ctx, sess, cloudInstanceID)
	ipsecPolicy, err := client.CreateIPSecPolicy(body)
	if err != nil {
		log.Printf("[DEBUG] create ipsec policy failed %v", err)
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s", cloudInstanceID, *ipsecPolicy.ID))

	return resourceIBMPIIPSecPolicyRead(ctx, d, meta)
}

func resourceIBMPIIPSecPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID, policyID, err := splitID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	client := instance.NewIBMPIVpnPolicyClient(ctx, sess, cloudInstanceID)
	body := &models.IPSecPolicyUpdate{}

	if d.HasChange(PIVPNPolicyName) {
		name := d.Get(PIVPNPolicyName).(string)
		body.Name = name
	}
	if d.HasChange(PIVPNPolicyDhGroup) {
		dhGroup := int64(d.Get(PIVPNPolicyDhGroup).(int))
		body.DhGroup = dhGroup
	}
	if d.HasChange(PIVPNPolicyEncryption) {
		encryption := d.Get(PIVPNPolicyEncryption).(string)
		body.Encryption = encryption
	}
	if d.HasChange(PIVPNPolicyKeyLifetime) {
		keyLifetime := int64(d.Get(PIVPNPolicyKeyLifetime).(int))
		body.KeyLifetime = models.KeyLifetime(keyLifetime)
	}
	if d.HasChange(PIVPNPolicyPFS) {
		pfs := d.Get(PIVPNPolicyPFS).(bool)
		body.Pfs = &pfs
	}
	if d.HasChange(PIVPNPolicyAuthentication) {
		authentication := d.Get(PIVPNPolicyAuthentication).(string)
		body.Authentication = models.IPSECPolicyAuthentication(authentication)
	}

	_, err = client.UpdateIPSecPolicy(policyID, body)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceIBMPIIPSecPolicyRead(ctx, d, meta)
}

func resourceIBMPIIPSecPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID, policyID, err := splitID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	client := instance.NewIBMPIVpnPolicyClient(ctx, sess, cloudInstanceID)
	ipsecPolicy, err := client.GetIPSecPolicy(policyID)
	if err != nil {
		// FIXME: Uncomment when 404 error is available
		// switch err.(type) {
		// case *p_cloud_v_p_n_policies.PcloudIPSecpoliciesGetNotFound:
		// 	log.Printf("[DEBUG] VPN policy does not exist %v", err)
		// 	d.SetId("")
		// 	return nil
		// }
		log.Printf("[DEBUG] get VPN policy failed %v", err)
		return diag.FromErr(err)
	}

	d.Set(PIPolicyId, ipsecPolicy.ID)
	d.Set(PIVPNPolicyName, ipsecPolicy.Name)
	d.Set(PIVPNPolicyDhGroup, ipsecPolicy.DhGroup)
	d.Set(PIVPNPolicyEncryption, ipsecPolicy.Encryption)
	d.Set(PIVPNPolicyKeyLifetime, ipsecPolicy.KeyLifetime)
	d.Set(PIVPNPolicyPFS, ipsecPolicy.Pfs)
	d.Set(PIVPNPolicyAuthentication, ipsecPolicy.Authentication)

	return nil
}

func resourceIBMPIIPSecPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID, policyID, err := splitID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	client := instance.NewIBMPIVpnPolicyClient(ctx, sess, cloudInstanceID)

	err = client.DeleteIPSecPolicy(policyID)
	if err != nil {
		// FIXME: Uncomment when 404 error is available
		// switch err.(type) {
		// case *p_cloud_v_p_n_policies.PcloudIPSecpoliciesDeleteNotFound:
		// 	log.Printf("[DEBUG] VPN policy does not exist %v", err)
		// 	d.SetId("")
		// 	return nil
		// }
		log.Printf("[DEBUG] delete VPN policy failed %v", err)
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
