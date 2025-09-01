package collection

import (
	"github.com/madhouselabs/anybase/internal/database/types"
	"github.com/madhouselabs/anybase/internal/governance"
	"github.com/madhouselabs/anybase/internal/validator"
)

// AdapterService is a new implementation that uses the database adapter
type AdapterService struct {
	db          types.DB
	rbacService governance.RBACService
	validator   *validator.SchemaValidator
}

// NewAdapterService creates a new adapter-based service
func NewAdapterService(db types.DB, rbacService governance.RBACService) *AdapterService {
	return &AdapterService{
		db:          db,
		rbacService: rbacService,
		validator:   validator.NewSchemaValidator(),
	}
}

// QueryOptions is defined in interfaces.go