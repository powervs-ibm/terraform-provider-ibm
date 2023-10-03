package power

import (
	"context"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/helpers"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	PERCapability         = "per"
	CCCapability          = "cloud_connections"
	VPNCapability         = "vpn_connections"
	WorkspaceCreationDate = "creation_date"
	WorkspaceCRN          = "crn"
	WorkspaceHref         = "href"
	WorkspaceRegion       = "region"
	WorkspaceType         = "type"
	WorkspaceUrl          = "url"
)

func DatasourceIBMPIWorkspace() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIBMPIWorkspaceRead,
		Schema: map[string]*schema.Schema{
			Arg_CloudInstanceID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			Attr_WorkspaceCapabilities: {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Workspace Capabilities",
				Elem: &schema.Schema{
					Type: schema.TypeBool,
				},
			},
			Attr_WorkspaceDetails: {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The workspace information",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			Attr_WorkspaceID: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Workspace ID",
			},
			Attr_WorkspaceLocation: {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Workspace location",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			Attr_WorkspaceName: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Workspace name",
			},
			Attr_WorkspaceStatus: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Workspace status",
			},
			Attr_WorkspaceType: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Workspace type",
			},
		},
	}
}

func dataSourceIBMPIWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// session
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := d.Get(helpers.PICloudInstanceId).(string)
	workspaceClient := instance.NewIBMPIWorkspacesClient(ctx, sess, cloudInstanceID)
	wsData, err := workspaceClient.Get()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(*wsData.ID)
	d.Set(Attr_WorkspaceName, wsData.Name)
	d.Set(Attr_WorkspaceID, wsData.ID)
	d.Set(Attr_WorkspaceStatus, wsData.Status)
	d.Set(Attr_WorkspaceType, wsData.Type)
	d.Set(Attr_WorkspaceCapabilities, wsData.Capabilities)
	wsdetails := map[string]string{
		WorkspaceCreationDate: wsData.Details.CreationDate.String(),
		WorkspaceCRN:          *wsData.Details.Crn,
	}
	d.Set(Attr_WorkspaceDetails, wsdetails)
	wslocation := map[string]string{
		WorkspaceRegion: *wsData.Location.Region,
		WorkspaceType:   *wsData.Location.Type,
		WorkspaceUrl:    *wsData.Location.URL,
	}
	d.Set(Attr_WorkspaceLocation, wslocation)
	return nil
}
