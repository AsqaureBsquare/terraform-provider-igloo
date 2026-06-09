package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = &IglooProvider{}

type IglooProvider struct {
	version string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &IglooProvider{version: version}
	}
}

func (p *IglooProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "igloo"
	resp.Version = p.version
}

func (p *IglooProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Igloo provider gives Terraform templates access to workspace metadata injected by the Igloo provisioner.",
	}
}

func (p *IglooProvider) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
	// No configuration needed — all data comes from env vars injected by Igloo.
}

func (p *IglooProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAgentResource,
		NewParameterResource,
	}
}

func (p *IglooProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
