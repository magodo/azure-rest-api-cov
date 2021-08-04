package operation

import (
	"github.com/magodo/azure-rest-api-cov/analyzer/specpath"
	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
)

type Info struct {
	RP             string                  `json:"-"`
	APIPath        string                  `json:"-"`
	APIVersion     string                  `json:"-"`
	IsPreview      bool                    `json:"is_preview"`
	IsActive       bool                    `json:"is_active"`
	Operation      workspace.OperationKind `json:"-"`
	SwaggerRelPath specpath.Path           `json:"swagger_rel_path"`
	Model          *swagger.Schema         `json:"model,omitempty"`
}

