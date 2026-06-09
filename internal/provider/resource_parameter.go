package provider

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ParameterResource{}

type ParameterResource struct{}

func NewParameterResource() resource.Resource {
	return &ParameterResource{}
}

type ParameterOption struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type ParameterResourceModel struct {
	ID          types.String      `tfsdk:"id"`
	Name        types.String      `tfsdk:"name"`
	DisplayName types.String      `tfsdk:"display_name"`
	Description types.String      `tfsdk:"description"`
	Type        types.String      `tfsdk:"type"`
	Default     types.String      `tfsdk:"default"`
	Mutable     types.Bool        `tfsdk:"mutable"`
	Option      []ParameterOption `tfsdk:"option"`
	Value       types.String      `tfsdk:"value"`
}

func (r *ParameterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_parameter"
}

func (r *ParameterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Declares a workspace parameter that users fill in when creating a workspace.

The Igloo provisioner injects each parameter as ` + "`IGLOO_PARAMETER_<NAME>`" + ` before running Terraform.
The ` + "`value`" + ` attribute contains the user-supplied value (or the default).

` + "```hcl\n" + `resource "igloo_parameter" "instance_type" {
  name         = "instance_type"
  display_name = "Instance Type"
  description  = "The resource profile for this workspace."
  type         = "string"
  default      = "small"
  mutable      = true

  option {
    name  = "Small (0.5 CPU, 512Mi)"
    value = "small"
  }
  option {
    name  = "Large (2 CPU, 2Gi)"
    value = "large"
  }
}
` + "```",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The parameter name. Must match the key used in the Igloo template variables.",
			},
			"display_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Human-readable label shown in the workspace creation form.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Helper text shown below the input field in the UI.",
			},
			"type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("string"),
				MarkdownDescription: "The value type: `string`, `number`, or `bool`.",
				Validators: []validator.String{
					stringvalidator.OneOf("string", "number", "bool"),
				},
			},
			"default": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Default value used when the user does not supply one.",
			},
			"mutable": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the parameter can be changed after workspace creation. Defaults to true.",
			},
			"value": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The resolved value: user-supplied input or the default.",
			},
		},
		Blocks: map[string]schema.Block{
			"option": schema.ListNestedBlock{
				MarkdownDescription: "Allowed values rendered as a dropdown in the UI. When set, the user must pick one of these.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Human-readable label for this option.",
						},
						"value": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The value submitted when this option is selected.",
						},
					},
				},
			},
		},
	}
}

func (r *ParameterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ParameterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(data.Name.ValueString())
	data.Value = types.StringValue(resolveValue(data))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ParameterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ParameterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Value = types.StringValue(resolveValue(data))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ParameterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ParameterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Value = types.StringValue(resolveValue(data))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ParameterResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {}

// resolveValue reads IGLOO_PARAMETER_<NAME> (uppercased, hyphens→underscores).
// Falls back to the declared default.
func resolveValue(data ParameterResourceModel) string {
	envKey := fmt.Sprintf("IGLOO_PARAMETER_%s", strings.ToUpper(strings.ReplaceAll(data.Name.ValueString(), "-", "_")))
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return data.Default.ValueString()
}
