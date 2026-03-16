package resources

import "github.com/hashicorp/terraform-plugin-framework/resource"

func NewExtMySQLResource() resource.Resource {
	return NewExtensionResource(ExtensionConfig{
		TypeName:    "mysql",
		Description: "Manages a Fly.io MySQL extension.",
		HasOrg:      true,
		HasRegion:   true,
	})()
}

func NewExtKubernetesResource() resource.Resource {
	return NewExtensionResource(ExtensionConfig{
		TypeName:    "kubernetes",
		Description: "Manages a Fly.io Kubernetes extension.",
		HasOrg:      true,
		HasRegion:   true,
	})()
}

func NewExtSentryResource() resource.Resource {
	return NewExtensionResource(ExtensionConfig{
		TypeName:    "sentry",
		Description: "Manages a Fly.io Sentry extension.",
		HasApp:      true,
	})()
}

func NewExtArcjetResource() resource.Resource {
	return NewExtensionResource(ExtensionConfig{
		TypeName:    "arcjet",
		Description: "Manages a Fly.io Arcjet extension.",
		HasApp:      true,
	})()
}

func NewExtWafrisResource() resource.Resource {
	return NewExtensionResource(ExtensionConfig{
		TypeName:    "wafris",
		Description: "Manages a Fly.io Wafris extension.",
		HasApp:      true,
	})()
}

func NewExtVectorResource() resource.Resource {
	return NewExtensionResource(ExtensionConfig{
		TypeName:    "vector",
		Description: "Manages a Fly.io Vector extension.",
		HasApp:      true,
	})()
}
