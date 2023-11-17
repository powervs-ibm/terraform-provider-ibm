package power

import (
	"context"
	"fmt"
	"log"
	"time"

	st "github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/helpers"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceIBMPIWorkspace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPIWorkspaceCreate,
		ReadContext:   resourceIBMPIWorkspaceRead,
		DeleteContext: resourceIBMPIWorkspaceDelete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			helpers.PICloudInstanceId: {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{helpers.PIWorkspaceName, helpers.PIWorkspaceDatacenter, helpers.PIWorkspaceResourceGroup, helpers.PIWorkspacePlan},
				Description:   "PI cloud instance ID",
			},
			helpers.PIWorkspaceName: {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{helpers.PICloudInstanceId},
				RequiredWith:  []string{helpers.PIWorkspaceDatacenter, helpers.PIWorkspaceResourceGroup, helpers.PIWorkspacePlan},
				Description:   "The desired name of the workspace",
			},
			helpers.PIWorkspaceDatacenter: {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{helpers.PICloudInstanceId},
				RequiredWith:  []string{helpers.PIWorkspaceName, helpers.PIWorkspaceResourceGroup, helpers.PIWorkspacePlan},
				Description:   "The datacenter location where the instance should be hosted",
			},
			helpers.PIWorkspaceResourceGroup: {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{helpers.PICloudInstanceId},
				RequiredWith:  []string{helpers.PIWorkspaceDatacenter, helpers.PIWorkspaceName, helpers.PIWorkspacePlan},
				Description:   "The ID of the resource group",
			},
			helpers.PIWorkspacePlan: {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{helpers.PICloudInstanceId},
				RequiredWith:  []string{helpers.PIWorkspaceDatacenter, helpers.PIWorkspaceResourceGroup, helpers.PIWorkspaceName},
				Description:   "Plan associated with the offering; Valid values are \"public\" or \"private\".",
			},
		},
	}
}

func resourceIBMPIWorkspaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get(helpers.PIWorkspaceName).(string)
	datacenter := d.Get(helpers.PIWorkspaceDatacenter).(string)
	resourceGroup := d.Get(helpers.PIWorkspaceResourceGroup).(string)
	plan := d.Get(helpers.PIWorkspacePlan).(string)

	// No need for cloudInstanceID because we are creating a workspace
	client := st.NewIBMPIWorkspacesClient(ctx, sess, "")
	controller, err := client.Create(name, datacenter, resourceGroup, plan)
	if err != nil {
		log.Printf("[DEBUG] create workspace failed %v", err)
		return diag.FromErr(err)
	}
	// d.Set(helpers.PICloudInstanceId, strings.Split(*controller.CRN, ":")[7])
	d.SetId(*controller.GUID)
	_, err = waitForResourceInstanceCreate(ctx, client, *controller.GUID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceIBMPIWorkspaceRead(ctx, d, meta)
}

func resourceIBMPIWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// session
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := d.Get(helpers.PICloudInstanceId).(string)

	client := st.NewIBMPIWorkspacesClient(ctx, sess, cloudInstanceID)
	wsData, err := client.Get(cloudInstanceID)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set(helpers.PICloudInstanceId, cloudInstanceID)
	d.Set(helpers.PIWorkspaceName, wsData.Name)
	// d.Set(helpers.PIWorkspaceDatacenter, helpers.PIWorkspaceDatacenter)
	// d.Set(helpers.PIWorkspaceResourceGroup, helpers.PIWorkspaceResourceGroup)
	// d.Set(helpers.PIWorkspacePlan, helpers.PIWorkspacePlan)

	return nil
}

func resourceIBMPIWorkspaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := d.Get(helpers.PICloudInstanceId).(string)
	client := st.NewIBMPIWorkspacesClient(ctx, sess, cloudInstanceID)
	err = client.Delete(cloudInstanceID)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func waitForResourceInstanceCreate(ctx context.Context, client *st.IBMPIWorkspacesClient, id string, timeout time.Duration) (interface{}, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"provisioning", "in progress", "inactive"},
		Target:     []string{"active"},
		Refresh:    isIBMPIWorkspaceRefreshFunc(client, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	return stateConf.WaitForStateContext(ctx)
}
func isIBMPIWorkspaceRefreshFunc(client *st.IBMPIWorkspacesClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		ws, err := client.GetRC(id)
		// ws, err := client.Get(id)
		if err != nil {
			return nil, "", err
		}
		if *ws.State == "failed" {
			return ws, *ws.State, fmt.Errorf("[ERROR] The resource instance %s failed to provisioned", id)
		}

		return ws, *ws.State, nil

	}
}
