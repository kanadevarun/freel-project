package contracts

import (
	"fmt"

	"github.com/freel/backend/internal/svcerror"
)

// ValidateContractLink ensures the relationship creation rules are followed.
func ValidateContractLink(contract *Contract, req *CreateContractLinkRequest) error {
	if contract.Status == ContractStatusArchived {
		e := svcerror.NewServiceError(svcerror.ErrInvalidArgument)
		e.Message = "Contract is locked or archived and cannot accept new links"
		return e
	}

	if req.LinkedEntityID <= 0 || req.LinkedEntityType == "" || req.LinkType == "" {
		e := svcerror.NewServiceError(svcerror.ErrInvalidArgument)
		e.Message = "linked_entity_id, linked_entity_type, and link_type are required"
		return e
	}

	// Basic compatibility checking
	if err := validateLinkCompatibility(req.LinkedEntityType, req.LinkType); err != nil {
		return err
	}

	return nil
}

func validateLinkCompatibility(entityType LinkedEntityType, linkType LinkType) error {
	switch entityType {
	case EntityTypeRate, EntityTypeRateContract:
		if linkType != LinkTypeCommercialRate && linkType != LinkTypeContractRate && linkType != LinkTypeRelated {
			e := svcerror.NewServiceError(svcerror.ErrInvalidArgument)
			e.Message = fmt.Sprintf("invalid link type %s for entity type %s", linkType, entityType)
			return e
		}
	case EntityTypeQuotation:
		if linkType != LinkTypeQuotation && linkType != LinkTypeRelated {
			e := svcerror.NewServiceError(svcerror.ErrInvalidArgument)
			e.Message = fmt.Sprintf("invalid link type %s for entity type %s", linkType, entityType)
			return e
		}
	case EntityTypeSpotRateRequest, EntityTypeSpotRateResponse:
		if linkType != LinkTypeSpotSourcing && linkType != LinkTypeRelated {
			e := svcerror.NewServiceError(svcerror.ErrInvalidArgument)
			e.Message = fmt.Sprintf("invalid link type %s for entity type %s", linkType, entityType)
			return e
		}
	case EntityTypeCustomer, EntityTypeCarrier, EntityTypeVendor:
		if linkType != LinkTypeCustomer && linkType != LinkTypeCarrier && linkType != LinkTypeVendor && linkType != LinkTypeRelated {
			e := svcerror.NewServiceError(svcerror.ErrInvalidArgument)
			e.Message = fmt.Sprintf("invalid link type %s for entity type %s", linkType, entityType)
			return e
		}
	}
	return nil
}

// CanModifyContractLink checks if an existing link can be modified or removed.
func CanModifyContractLink(contract *Contract) error {
	if contract.Status == ContractStatusArchived {
		e := svcerror.NewServiceError(svcerror.ErrInvalidArgument)
		e.Message = "Cannot modify links on an archived contract"
		return e
	}
	return nil
}
