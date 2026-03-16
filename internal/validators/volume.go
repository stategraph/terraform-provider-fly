package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type volumeSizeValidator struct{}

func VolumeSizeCannotDecrease() validator.Int32 {
	return volumeSizeValidator{}
}

func (v volumeSizeValidator) Description(_ context.Context) string {
	return "Volume size can only be increased, not decreased."
}

func (v volumeSizeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v volumeSizeValidator) ValidateInt32(_ context.Context, req validator.Int32Request, resp *validator.Int32Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// This validator provides documentation; actual shrink prevention is in the Update method
	// since we need to compare against the current state value, not just validate the config.
	val := req.ConfigValue.ValueInt32()
	if val < 1 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Volume Size",
			fmt.Sprintf("Volume size must be at least 1 GB, got %d", val),
		)
	}
}
