// Code generated by ent, DO NOT EDIT.

package ent

import (
	"errors"
	"fmt"
	"time"

	"github.com/theopenlane/entx/vanilla/_example/ent/enums"
	"github.com/theopenlane/entx/vanilla/_example/ent/organization"
	"github.com/theopenlane/entx/vanilla/_example/ent/orgmembership"
	"github.com/theopenlane/entx/vanilla/_example/ent/predicate"
)

// OrgMembershipWhereInput represents a where input for filtering OrgMembership queries.
type OrgMembershipWhereInput struct {
	Predicates []predicate.OrgMembership  `json:"-"`
	Not        *OrgMembershipWhereInput   `json:"not,omitempty"`
	Or         []*OrgMembershipWhereInput `json:"or,omitempty"`
	And        []*OrgMembershipWhereInput `json:"and,omitempty"`

	// "id" field predicates.
	ID             *string  `json:"id,omitempty"`
	IDNEQ          *string  `json:"idNEQ,omitempty"`
	IDIn           []string `json:"idIn,omitempty"`
	IDNotIn        []string `json:"idNotIn,omitempty"`
	IDGT           *string  `json:"idGT,omitempty"`
	IDGTE          *string  `json:"idGTE,omitempty"`
	IDLT           *string  `json:"idLT,omitempty"`
	IDLTE          *string  `json:"idLTE,omitempty"`
	IDEqualFold    *string  `json:"idEqualFold,omitempty"`
	IDContainsFold *string  `json:"idContainsFold,omitempty"`

	// "role" field predicates.
	Role      *enums.Role  `json:"role,omitempty"`
	RoleNEQ   *enums.Role  `json:"roleNEQ,omitempty"`
	RoleIn    []enums.Role `json:"roleIn,omitempty"`
	RoleNotIn []enums.Role `json:"roleNotIn,omitempty"`

	// "organization_id" field predicates.
	OrganizationID             *string  `json:"organizationID,omitempty"`
	OrganizationIDNEQ          *string  `json:"organizationIDNEQ,omitempty"`
	OrganizationIDIn           []string `json:"organizationIDIn,omitempty"`
	OrganizationIDNotIn        []string `json:"organizationIDNotIn,omitempty"`
	OrganizationIDGT           *string  `json:"organizationIDGT,omitempty"`
	OrganizationIDGTE          *string  `json:"organizationIDGTE,omitempty"`
	OrganizationIDLT           *string  `json:"organizationIDLT,omitempty"`
	OrganizationIDLTE          *string  `json:"organizationIDLTE,omitempty"`
	OrganizationIDContains     *string  `json:"organizationIDContains,omitempty"`
	OrganizationIDHasPrefix    *string  `json:"organizationIDHasPrefix,omitempty"`
	OrganizationIDHasSuffix    *string  `json:"organizationIDHasSuffix,omitempty"`
	OrganizationIDEqualFold    *string  `json:"organizationIDEqualFold,omitempty"`
	OrganizationIDContainsFold *string  `json:"organizationIDContainsFold,omitempty"`

	// "user_id" field predicates.
	UserID             *string  `json:"userID,omitempty"`
	UserIDNEQ          *string  `json:"userIDNEQ,omitempty"`
	UserIDIn           []string `json:"userIDIn,omitempty"`
	UserIDNotIn        []string `json:"userIDNotIn,omitempty"`
	UserIDGT           *string  `json:"userIDGT,omitempty"`
	UserIDGTE          *string  `json:"userIDGTE,omitempty"`
	UserIDLT           *string  `json:"userIDLT,omitempty"`
	UserIDLTE          *string  `json:"userIDLTE,omitempty"`
	UserIDContains     *string  `json:"userIDContains,omitempty"`
	UserIDHasPrefix    *string  `json:"userIDHasPrefix,omitempty"`
	UserIDHasSuffix    *string  `json:"userIDHasSuffix,omitempty"`
	UserIDEqualFold    *string  `json:"userIDEqualFold,omitempty"`
	UserIDContainsFold *string  `json:"userIDContainsFold,omitempty"`

	// "organization" edge predicates.
	HasOrganization     *bool                     `json:"hasOrganization,omitempty"`
	HasOrganizationWith []*OrganizationWhereInput `json:"hasOrganizationWith,omitempty"`
}

// AddPredicates adds custom predicates to the where input to be used during the filtering phase.
func (i *OrgMembershipWhereInput) AddPredicates(predicates ...predicate.OrgMembership) {
	i.Predicates = append(i.Predicates, predicates...)
}

// Filter applies the OrgMembershipWhereInput filter on the OrgMembershipQuery builder.
func (i *OrgMembershipWhereInput) Filter(q *OrgMembershipQuery) (*OrgMembershipQuery, error) {
	if i == nil {
		return q, nil
	}
	p, err := i.P()
	if err != nil {
		if err == ErrEmptyOrgMembershipWhereInput {
			return q, nil
		}
		return nil, err
	}
	return q.Where(p), nil
}

// ErrEmptyOrgMembershipWhereInput is returned in case the OrgMembershipWhereInput is empty.
var ErrEmptyOrgMembershipWhereInput = errors.New("ent: empty predicate OrgMembershipWhereInput")

// P returns a predicate for filtering orgmemberships.
// An error is returned if the input is empty or invalid.
func (i *OrgMembershipWhereInput) P() (predicate.OrgMembership, error) {
	var predicates []predicate.OrgMembership
	if i.Not != nil {
		p, err := i.Not.P()
		if err != nil {
			return nil, fmt.Errorf("%w: field 'not'", err)
		}
		predicates = append(predicates, orgmembership.Not(p))
	}
	switch n := len(i.Or); {
	case n == 1:
		p, err := i.Or[0].P()
		if err != nil {
			return nil, fmt.Errorf("%w: field 'or'", err)
		}
		predicates = append(predicates, p)
	case n > 1:
		or := make([]predicate.OrgMembership, 0, n)
		for _, w := range i.Or {
			p, err := w.P()
			if err != nil {
				return nil, fmt.Errorf("%w: field 'or'", err)
			}
			or = append(or, p)
		}
		predicates = append(predicates, orgmembership.Or(or...))
	}
	switch n := len(i.And); {
	case n == 1:
		p, err := i.And[0].P()
		if err != nil {
			return nil, fmt.Errorf("%w: field 'and'", err)
		}
		predicates = append(predicates, p)
	case n > 1:
		and := make([]predicate.OrgMembership, 0, n)
		for _, w := range i.And {
			p, err := w.P()
			if err != nil {
				return nil, fmt.Errorf("%w: field 'and'", err)
			}
			and = append(and, p)
		}
		predicates = append(predicates, orgmembership.And(and...))
	}
	predicates = append(predicates, i.Predicates...)
	if i.ID != nil {
		predicates = append(predicates, orgmembership.IDEQ(*i.ID))
	}
	if i.IDNEQ != nil {
		predicates = append(predicates, orgmembership.IDNEQ(*i.IDNEQ))
	}
	if len(i.IDIn) > 0 {
		predicates = append(predicates, orgmembership.IDIn(i.IDIn...))
	}
	if len(i.IDNotIn) > 0 {
		predicates = append(predicates, orgmembership.IDNotIn(i.IDNotIn...))
	}
	if i.IDGT != nil {
		predicates = append(predicates, orgmembership.IDGT(*i.IDGT))
	}
	if i.IDGTE != nil {
		predicates = append(predicates, orgmembership.IDGTE(*i.IDGTE))
	}
	if i.IDLT != nil {
		predicates = append(predicates, orgmembership.IDLT(*i.IDLT))
	}
	if i.IDLTE != nil {
		predicates = append(predicates, orgmembership.IDLTE(*i.IDLTE))
	}
	if i.IDEqualFold != nil {
		predicates = append(predicates, orgmembership.IDEqualFold(*i.IDEqualFold))
	}
	if i.IDContainsFold != nil {
		predicates = append(predicates, orgmembership.IDContainsFold(*i.IDContainsFold))
	}
	if i.Role != nil {
		predicates = append(predicates, orgmembership.RoleEQ(*i.Role))
	}
	if i.RoleNEQ != nil {
		predicates = append(predicates, orgmembership.RoleNEQ(*i.RoleNEQ))
	}
	if len(i.RoleIn) > 0 {
		predicates = append(predicates, orgmembership.RoleIn(i.RoleIn...))
	}
	if len(i.RoleNotIn) > 0 {
		predicates = append(predicates, orgmembership.RoleNotIn(i.RoleNotIn...))
	}
	if i.OrganizationID != nil {
		predicates = append(predicates, orgmembership.OrganizationIDEQ(*i.OrganizationID))
	}
	if i.OrganizationIDNEQ != nil {
		predicates = append(predicates, orgmembership.OrganizationIDNEQ(*i.OrganizationIDNEQ))
	}
	if len(i.OrganizationIDIn) > 0 {
		predicates = append(predicates, orgmembership.OrganizationIDIn(i.OrganizationIDIn...))
	}
	if len(i.OrganizationIDNotIn) > 0 {
		predicates = append(predicates, orgmembership.OrganizationIDNotIn(i.OrganizationIDNotIn...))
	}
	if i.OrganizationIDGT != nil {
		predicates = append(predicates, orgmembership.OrganizationIDGT(*i.OrganizationIDGT))
	}
	if i.OrganizationIDGTE != nil {
		predicates = append(predicates, orgmembership.OrganizationIDGTE(*i.OrganizationIDGTE))
	}
	if i.OrganizationIDLT != nil {
		predicates = append(predicates, orgmembership.OrganizationIDLT(*i.OrganizationIDLT))
	}
	if i.OrganizationIDLTE != nil {
		predicates = append(predicates, orgmembership.OrganizationIDLTE(*i.OrganizationIDLTE))
	}
	if i.OrganizationIDContains != nil {
		predicates = append(predicates, orgmembership.OrganizationIDContains(*i.OrganizationIDContains))
	}
	if i.OrganizationIDHasPrefix != nil {
		predicates = append(predicates, orgmembership.OrganizationIDHasPrefix(*i.OrganizationIDHasPrefix))
	}
	if i.OrganizationIDHasSuffix != nil {
		predicates = append(predicates, orgmembership.OrganizationIDHasSuffix(*i.OrganizationIDHasSuffix))
	}
	if i.OrganizationIDEqualFold != nil {
		predicates = append(predicates, orgmembership.OrganizationIDEqualFold(*i.OrganizationIDEqualFold))
	}
	if i.OrganizationIDContainsFold != nil {
		predicates = append(predicates, orgmembership.OrganizationIDContainsFold(*i.OrganizationIDContainsFold))
	}
	if i.UserID != nil {
		predicates = append(predicates, orgmembership.UserIDEQ(*i.UserID))
	}
	if i.UserIDNEQ != nil {
		predicates = append(predicates, orgmembership.UserIDNEQ(*i.UserIDNEQ))
	}
	if len(i.UserIDIn) > 0 {
		predicates = append(predicates, orgmembership.UserIDIn(i.UserIDIn...))
	}
	if len(i.UserIDNotIn) > 0 {
		predicates = append(predicates, orgmembership.UserIDNotIn(i.UserIDNotIn...))
	}
	if i.UserIDGT != nil {
		predicates = append(predicates, orgmembership.UserIDGT(*i.UserIDGT))
	}
	if i.UserIDGTE != nil {
		predicates = append(predicates, orgmembership.UserIDGTE(*i.UserIDGTE))
	}
	if i.UserIDLT != nil {
		predicates = append(predicates, orgmembership.UserIDLT(*i.UserIDLT))
	}
	if i.UserIDLTE != nil {
		predicates = append(predicates, orgmembership.UserIDLTE(*i.UserIDLTE))
	}
	if i.UserIDContains != nil {
		predicates = append(predicates, orgmembership.UserIDContains(*i.UserIDContains))
	}
	if i.UserIDHasPrefix != nil {
		predicates = append(predicates, orgmembership.UserIDHasPrefix(*i.UserIDHasPrefix))
	}
	if i.UserIDHasSuffix != nil {
		predicates = append(predicates, orgmembership.UserIDHasSuffix(*i.UserIDHasSuffix))
	}
	if i.UserIDEqualFold != nil {
		predicates = append(predicates, orgmembership.UserIDEqualFold(*i.UserIDEqualFold))
	}
	if i.UserIDContainsFold != nil {
		predicates = append(predicates, orgmembership.UserIDContainsFold(*i.UserIDContainsFold))
	}

	if i.HasOrganization != nil {
		p := orgmembership.HasOrganization()
		if !*i.HasOrganization {
			p = orgmembership.Not(p)
		}
		predicates = append(predicates, p)
	}
	if len(i.HasOrganizationWith) > 0 {
		with := make([]predicate.Organization, 0, len(i.HasOrganizationWith))
		for _, w := range i.HasOrganizationWith {
			p, err := w.P()
			if err != nil {
				return nil, fmt.Errorf("%w: field 'HasOrganizationWith'", err)
			}
			with = append(with, p)
		}
		predicates = append(predicates, orgmembership.HasOrganizationWith(with...))
	}
	switch len(predicates) {
	case 0:
		return nil, ErrEmptyOrgMembershipWhereInput
	case 1:
		return predicates[0], nil
	default:
		return orgmembership.And(predicates...), nil
	}
}

// OrganizationWhereInput represents a where input for filtering Organization queries.
type OrganizationWhereInput struct {
	Predicates []predicate.Organization  `json:"-"`
	Not        *OrganizationWhereInput   `json:"not,omitempty"`
	Or         []*OrganizationWhereInput `json:"or,omitempty"`
	And        []*OrganizationWhereInput `json:"and,omitempty"`

	// "id" field predicates.
	ID             *string  `json:"id,omitempty"`
	IDNEQ          *string  `json:"idNEQ,omitempty"`
	IDIn           []string `json:"idIn,omitempty"`
	IDNotIn        []string `json:"idNotIn,omitempty"`
	IDGT           *string  `json:"idGT,omitempty"`
	IDGTE          *string  `json:"idGTE,omitempty"`
	IDLT           *string  `json:"idLT,omitempty"`
	IDLTE          *string  `json:"idLTE,omitempty"`
	IDEqualFold    *string  `json:"idEqualFold,omitempty"`
	IDContainsFold *string  `json:"idContainsFold,omitempty"`

	// "display_id" field predicates.
	DisplayID             *string  `json:"displayID,omitempty"`
	DisplayIDNEQ          *string  `json:"displayIDNEQ,omitempty"`
	DisplayIDIn           []string `json:"displayIDIn,omitempty"`
	DisplayIDNotIn        []string `json:"displayIDNotIn,omitempty"`
	DisplayIDGT           *string  `json:"displayIDGT,omitempty"`
	DisplayIDGTE          *string  `json:"displayIDGTE,omitempty"`
	DisplayIDLT           *string  `json:"displayIDLT,omitempty"`
	DisplayIDLTE          *string  `json:"displayIDLTE,omitempty"`
	DisplayIDContains     *string  `json:"displayIDContains,omitempty"`
	DisplayIDHasPrefix    *string  `json:"displayIDHasPrefix,omitempty"`
	DisplayIDHasSuffix    *string  `json:"displayIDHasSuffix,omitempty"`
	DisplayIDEqualFold    *string  `json:"displayIDEqualFold,omitempty"`
	DisplayIDContainsFold *string  `json:"displayIDContainsFold,omitempty"`

	// "created_at" field predicates.
	CreatedAt       *time.Time  `json:"createdAt,omitempty"`
	CreatedAtNEQ    *time.Time  `json:"createdAtNEQ,omitempty"`
	CreatedAtIn     []time.Time `json:"createdAtIn,omitempty"`
	CreatedAtNotIn  []time.Time `json:"createdAtNotIn,omitempty"`
	CreatedAtGT     *time.Time  `json:"createdAtGT,omitempty"`
	CreatedAtGTE    *time.Time  `json:"createdAtGTE,omitempty"`
	CreatedAtLT     *time.Time  `json:"createdAtLT,omitempty"`
	CreatedAtLTE    *time.Time  `json:"createdAtLTE,omitempty"`
	CreatedAtIsNil  bool        `json:"createdAtIsNil,omitempty"`
	CreatedAtNotNil bool        `json:"createdAtNotNil,omitempty"`

	// "updated_at" field predicates.
	UpdatedAt       *time.Time  `json:"updatedAt,omitempty"`
	UpdatedAtNEQ    *time.Time  `json:"updatedAtNEQ,omitempty"`
	UpdatedAtIn     []time.Time `json:"updatedAtIn,omitempty"`
	UpdatedAtNotIn  []time.Time `json:"updatedAtNotIn,omitempty"`
	UpdatedAtGT     *time.Time  `json:"updatedAtGT,omitempty"`
	UpdatedAtGTE    *time.Time  `json:"updatedAtGTE,omitempty"`
	UpdatedAtLT     *time.Time  `json:"updatedAtLT,omitempty"`
	UpdatedAtLTE    *time.Time  `json:"updatedAtLTE,omitempty"`
	UpdatedAtIsNil  bool        `json:"updatedAtIsNil,omitempty"`
	UpdatedAtNotNil bool        `json:"updatedAtNotNil,omitempty"`

	// "created_by" field predicates.
	CreatedBy             *string  `json:"createdBy,omitempty"`
	CreatedByNEQ          *string  `json:"createdByNEQ,omitempty"`
	CreatedByIn           []string `json:"createdByIn,omitempty"`
	CreatedByNotIn        []string `json:"createdByNotIn,omitempty"`
	CreatedByGT           *string  `json:"createdByGT,omitempty"`
	CreatedByGTE          *string  `json:"createdByGTE,omitempty"`
	CreatedByLT           *string  `json:"createdByLT,omitempty"`
	CreatedByLTE          *string  `json:"createdByLTE,omitempty"`
	CreatedByContains     *string  `json:"createdByContains,omitempty"`
	CreatedByHasPrefix    *string  `json:"createdByHasPrefix,omitempty"`
	CreatedByHasSuffix    *string  `json:"createdByHasSuffix,omitempty"`
	CreatedByIsNil        bool     `json:"createdByIsNil,omitempty"`
	CreatedByNotNil       bool     `json:"createdByNotNil,omitempty"`
	CreatedByEqualFold    *string  `json:"createdByEqualFold,omitempty"`
	CreatedByContainsFold *string  `json:"createdByContainsFold,omitempty"`

	// "updated_by" field predicates.
	UpdatedBy             *string  `json:"updatedBy,omitempty"`
	UpdatedByNEQ          *string  `json:"updatedByNEQ,omitempty"`
	UpdatedByIn           []string `json:"updatedByIn,omitempty"`
	UpdatedByNotIn        []string `json:"updatedByNotIn,omitempty"`
	UpdatedByGT           *string  `json:"updatedByGT,omitempty"`
	UpdatedByGTE          *string  `json:"updatedByGTE,omitempty"`
	UpdatedByLT           *string  `json:"updatedByLT,omitempty"`
	UpdatedByLTE          *string  `json:"updatedByLTE,omitempty"`
	UpdatedByContains     *string  `json:"updatedByContains,omitempty"`
	UpdatedByHasPrefix    *string  `json:"updatedByHasPrefix,omitempty"`
	UpdatedByHasSuffix    *string  `json:"updatedByHasSuffix,omitempty"`
	UpdatedByIsNil        bool     `json:"updatedByIsNil,omitempty"`
	UpdatedByNotNil       bool     `json:"updatedByNotNil,omitempty"`
	UpdatedByEqualFold    *string  `json:"updatedByEqualFold,omitempty"`
	UpdatedByContainsFold *string  `json:"updatedByContainsFold,omitempty"`

	// "name" field predicates.
	Name             *string  `json:"name,omitempty"`
	NameNEQ          *string  `json:"nameNEQ,omitempty"`
	NameIn           []string `json:"nameIn,omitempty"`
	NameNotIn        []string `json:"nameNotIn,omitempty"`
	NameGT           *string  `json:"nameGT,omitempty"`
	NameGTE          *string  `json:"nameGTE,omitempty"`
	NameLT           *string  `json:"nameLT,omitempty"`
	NameLTE          *string  `json:"nameLTE,omitempty"`
	NameContains     *string  `json:"nameContains,omitempty"`
	NameHasPrefix    *string  `json:"nameHasPrefix,omitempty"`
	NameHasSuffix    *string  `json:"nameHasSuffix,omitempty"`
	NameEqualFold    *string  `json:"nameEqualFold,omitempty"`
	NameContainsFold *string  `json:"nameContainsFold,omitempty"`

	// "description" field predicates.
	Description             *string  `json:"description,omitempty"`
	DescriptionNEQ          *string  `json:"descriptionNEQ,omitempty"`
	DescriptionIn           []string `json:"descriptionIn,omitempty"`
	DescriptionNotIn        []string `json:"descriptionNotIn,omitempty"`
	DescriptionGT           *string  `json:"descriptionGT,omitempty"`
	DescriptionGTE          *string  `json:"descriptionGTE,omitempty"`
	DescriptionLT           *string  `json:"descriptionLT,omitempty"`
	DescriptionLTE          *string  `json:"descriptionLTE,omitempty"`
	DescriptionContains     *string  `json:"descriptionContains,omitempty"`
	DescriptionHasPrefix    *string  `json:"descriptionHasPrefix,omitempty"`
	DescriptionHasSuffix    *string  `json:"descriptionHasSuffix,omitempty"`
	DescriptionIsNil        bool     `json:"descriptionIsNil,omitempty"`
	DescriptionNotNil       bool     `json:"descriptionNotNil,omitempty"`
	DescriptionEqualFold    *string  `json:"descriptionEqualFold,omitempty"`
	DescriptionContainsFold *string  `json:"descriptionContainsFold,omitempty"`
}

// AddPredicates adds custom predicates to the where input to be used during the filtering phase.
func (i *OrganizationWhereInput) AddPredicates(predicates ...predicate.Organization) {
	i.Predicates = append(i.Predicates, predicates...)
}

// Filter applies the OrganizationWhereInput filter on the OrganizationQuery builder.
func (i *OrganizationWhereInput) Filter(q *OrganizationQuery) (*OrganizationQuery, error) {
	if i == nil {
		return q, nil
	}
	p, err := i.P()
	if err != nil {
		if err == ErrEmptyOrganizationWhereInput {
			return q, nil
		}
		return nil, err
	}
	return q.Where(p), nil
}

// ErrEmptyOrganizationWhereInput is returned in case the OrganizationWhereInput is empty.
var ErrEmptyOrganizationWhereInput = errors.New("ent: empty predicate OrganizationWhereInput")

// P returns a predicate for filtering organizations.
// An error is returned if the input is empty or invalid.
func (i *OrganizationWhereInput) P() (predicate.Organization, error) {
	var predicates []predicate.Organization
	if i.Not != nil {
		p, err := i.Not.P()
		if err != nil {
			return nil, fmt.Errorf("%w: field 'not'", err)
		}
		predicates = append(predicates, organization.Not(p))
	}
	switch n := len(i.Or); {
	case n == 1:
		p, err := i.Or[0].P()
		if err != nil {
			return nil, fmt.Errorf("%w: field 'or'", err)
		}
		predicates = append(predicates, p)
	case n > 1:
		or := make([]predicate.Organization, 0, n)
		for _, w := range i.Or {
			p, err := w.P()
			if err != nil {
				return nil, fmt.Errorf("%w: field 'or'", err)
			}
			or = append(or, p)
		}
		predicates = append(predicates, organization.Or(or...))
	}
	switch n := len(i.And); {
	case n == 1:
		p, err := i.And[0].P()
		if err != nil {
			return nil, fmt.Errorf("%w: field 'and'", err)
		}
		predicates = append(predicates, p)
	case n > 1:
		and := make([]predicate.Organization, 0, n)
		for _, w := range i.And {
			p, err := w.P()
			if err != nil {
				return nil, fmt.Errorf("%w: field 'and'", err)
			}
			and = append(and, p)
		}
		predicates = append(predicates, organization.And(and...))
	}
	predicates = append(predicates, i.Predicates...)
	if i.ID != nil {
		predicates = append(predicates, organization.IDEQ(*i.ID))
	}
	if i.IDNEQ != nil {
		predicates = append(predicates, organization.IDNEQ(*i.IDNEQ))
	}
	if len(i.IDIn) > 0 {
		predicates = append(predicates, organization.IDIn(i.IDIn...))
	}
	if len(i.IDNotIn) > 0 {
		predicates = append(predicates, organization.IDNotIn(i.IDNotIn...))
	}
	if i.IDGT != nil {
		predicates = append(predicates, organization.IDGT(*i.IDGT))
	}
	if i.IDGTE != nil {
		predicates = append(predicates, organization.IDGTE(*i.IDGTE))
	}
	if i.IDLT != nil {
		predicates = append(predicates, organization.IDLT(*i.IDLT))
	}
	if i.IDLTE != nil {
		predicates = append(predicates, organization.IDLTE(*i.IDLTE))
	}
	if i.IDEqualFold != nil {
		predicates = append(predicates, organization.IDEqualFold(*i.IDEqualFold))
	}
	if i.IDContainsFold != nil {
		predicates = append(predicates, organization.IDContainsFold(*i.IDContainsFold))
	}
	if i.DisplayID != nil {
		predicates = append(predicates, organization.DisplayIDEQ(*i.DisplayID))
	}
	if i.DisplayIDNEQ != nil {
		predicates = append(predicates, organization.DisplayIDNEQ(*i.DisplayIDNEQ))
	}
	if len(i.DisplayIDIn) > 0 {
		predicates = append(predicates, organization.DisplayIDIn(i.DisplayIDIn...))
	}
	if len(i.DisplayIDNotIn) > 0 {
		predicates = append(predicates, organization.DisplayIDNotIn(i.DisplayIDNotIn...))
	}
	if i.DisplayIDGT != nil {
		predicates = append(predicates, organization.DisplayIDGT(*i.DisplayIDGT))
	}
	if i.DisplayIDGTE != nil {
		predicates = append(predicates, organization.DisplayIDGTE(*i.DisplayIDGTE))
	}
	if i.DisplayIDLT != nil {
		predicates = append(predicates, organization.DisplayIDLT(*i.DisplayIDLT))
	}
	if i.DisplayIDLTE != nil {
		predicates = append(predicates, organization.DisplayIDLTE(*i.DisplayIDLTE))
	}
	if i.DisplayIDContains != nil {
		predicates = append(predicates, organization.DisplayIDContains(*i.DisplayIDContains))
	}
	if i.DisplayIDHasPrefix != nil {
		predicates = append(predicates, organization.DisplayIDHasPrefix(*i.DisplayIDHasPrefix))
	}
	if i.DisplayIDHasSuffix != nil {
		predicates = append(predicates, organization.DisplayIDHasSuffix(*i.DisplayIDHasSuffix))
	}
	if i.DisplayIDEqualFold != nil {
		predicates = append(predicates, organization.DisplayIDEqualFold(*i.DisplayIDEqualFold))
	}
	if i.DisplayIDContainsFold != nil {
		predicates = append(predicates, organization.DisplayIDContainsFold(*i.DisplayIDContainsFold))
	}
	if i.CreatedAt != nil {
		predicates = append(predicates, organization.CreatedAtEQ(*i.CreatedAt))
	}
	if i.CreatedAtNEQ != nil {
		predicates = append(predicates, organization.CreatedAtNEQ(*i.CreatedAtNEQ))
	}
	if len(i.CreatedAtIn) > 0 {
		predicates = append(predicates, organization.CreatedAtIn(i.CreatedAtIn...))
	}
	if len(i.CreatedAtNotIn) > 0 {
		predicates = append(predicates, organization.CreatedAtNotIn(i.CreatedAtNotIn...))
	}
	if i.CreatedAtGT != nil {
		predicates = append(predicates, organization.CreatedAtGT(*i.CreatedAtGT))
	}
	if i.CreatedAtGTE != nil {
		predicates = append(predicates, organization.CreatedAtGTE(*i.CreatedAtGTE))
	}
	if i.CreatedAtLT != nil {
		predicates = append(predicates, organization.CreatedAtLT(*i.CreatedAtLT))
	}
	if i.CreatedAtLTE != nil {
		predicates = append(predicates, organization.CreatedAtLTE(*i.CreatedAtLTE))
	}
	if i.CreatedAtIsNil {
		predicates = append(predicates, organization.CreatedAtIsNil())
	}
	if i.CreatedAtNotNil {
		predicates = append(predicates, organization.CreatedAtNotNil())
	}
	if i.UpdatedAt != nil {
		predicates = append(predicates, organization.UpdatedAtEQ(*i.UpdatedAt))
	}
	if i.UpdatedAtNEQ != nil {
		predicates = append(predicates, organization.UpdatedAtNEQ(*i.UpdatedAtNEQ))
	}
	if len(i.UpdatedAtIn) > 0 {
		predicates = append(predicates, organization.UpdatedAtIn(i.UpdatedAtIn...))
	}
	if len(i.UpdatedAtNotIn) > 0 {
		predicates = append(predicates, organization.UpdatedAtNotIn(i.UpdatedAtNotIn...))
	}
	if i.UpdatedAtGT != nil {
		predicates = append(predicates, organization.UpdatedAtGT(*i.UpdatedAtGT))
	}
	if i.UpdatedAtGTE != nil {
		predicates = append(predicates, organization.UpdatedAtGTE(*i.UpdatedAtGTE))
	}
	if i.UpdatedAtLT != nil {
		predicates = append(predicates, organization.UpdatedAtLT(*i.UpdatedAtLT))
	}
	if i.UpdatedAtLTE != nil {
		predicates = append(predicates, organization.UpdatedAtLTE(*i.UpdatedAtLTE))
	}
	if i.UpdatedAtIsNil {
		predicates = append(predicates, organization.UpdatedAtIsNil())
	}
	if i.UpdatedAtNotNil {
		predicates = append(predicates, organization.UpdatedAtNotNil())
	}
	if i.CreatedBy != nil {
		predicates = append(predicates, organization.CreatedByEQ(*i.CreatedBy))
	}
	if i.CreatedByNEQ != nil {
		predicates = append(predicates, organization.CreatedByNEQ(*i.CreatedByNEQ))
	}
	if len(i.CreatedByIn) > 0 {
		predicates = append(predicates, organization.CreatedByIn(i.CreatedByIn...))
	}
	if len(i.CreatedByNotIn) > 0 {
		predicates = append(predicates, organization.CreatedByNotIn(i.CreatedByNotIn...))
	}
	if i.CreatedByGT != nil {
		predicates = append(predicates, organization.CreatedByGT(*i.CreatedByGT))
	}
	if i.CreatedByGTE != nil {
		predicates = append(predicates, organization.CreatedByGTE(*i.CreatedByGTE))
	}
	if i.CreatedByLT != nil {
		predicates = append(predicates, organization.CreatedByLT(*i.CreatedByLT))
	}
	if i.CreatedByLTE != nil {
		predicates = append(predicates, organization.CreatedByLTE(*i.CreatedByLTE))
	}
	if i.CreatedByContains != nil {
		predicates = append(predicates, organization.CreatedByContains(*i.CreatedByContains))
	}
	if i.CreatedByHasPrefix != nil {
		predicates = append(predicates, organization.CreatedByHasPrefix(*i.CreatedByHasPrefix))
	}
	if i.CreatedByHasSuffix != nil {
		predicates = append(predicates, organization.CreatedByHasSuffix(*i.CreatedByHasSuffix))
	}
	if i.CreatedByIsNil {
		predicates = append(predicates, organization.CreatedByIsNil())
	}
	if i.CreatedByNotNil {
		predicates = append(predicates, organization.CreatedByNotNil())
	}
	if i.CreatedByEqualFold != nil {
		predicates = append(predicates, organization.CreatedByEqualFold(*i.CreatedByEqualFold))
	}
	if i.CreatedByContainsFold != nil {
		predicates = append(predicates, organization.CreatedByContainsFold(*i.CreatedByContainsFold))
	}
	if i.UpdatedBy != nil {
		predicates = append(predicates, organization.UpdatedByEQ(*i.UpdatedBy))
	}
	if i.UpdatedByNEQ != nil {
		predicates = append(predicates, organization.UpdatedByNEQ(*i.UpdatedByNEQ))
	}
	if len(i.UpdatedByIn) > 0 {
		predicates = append(predicates, organization.UpdatedByIn(i.UpdatedByIn...))
	}
	if len(i.UpdatedByNotIn) > 0 {
		predicates = append(predicates, organization.UpdatedByNotIn(i.UpdatedByNotIn...))
	}
	if i.UpdatedByGT != nil {
		predicates = append(predicates, organization.UpdatedByGT(*i.UpdatedByGT))
	}
	if i.UpdatedByGTE != nil {
		predicates = append(predicates, organization.UpdatedByGTE(*i.UpdatedByGTE))
	}
	if i.UpdatedByLT != nil {
		predicates = append(predicates, organization.UpdatedByLT(*i.UpdatedByLT))
	}
	if i.UpdatedByLTE != nil {
		predicates = append(predicates, organization.UpdatedByLTE(*i.UpdatedByLTE))
	}
	if i.UpdatedByContains != nil {
		predicates = append(predicates, organization.UpdatedByContains(*i.UpdatedByContains))
	}
	if i.UpdatedByHasPrefix != nil {
		predicates = append(predicates, organization.UpdatedByHasPrefix(*i.UpdatedByHasPrefix))
	}
	if i.UpdatedByHasSuffix != nil {
		predicates = append(predicates, organization.UpdatedByHasSuffix(*i.UpdatedByHasSuffix))
	}
	if i.UpdatedByIsNil {
		predicates = append(predicates, organization.UpdatedByIsNil())
	}
	if i.UpdatedByNotNil {
		predicates = append(predicates, organization.UpdatedByNotNil())
	}
	if i.UpdatedByEqualFold != nil {
		predicates = append(predicates, organization.UpdatedByEqualFold(*i.UpdatedByEqualFold))
	}
	if i.UpdatedByContainsFold != nil {
		predicates = append(predicates, organization.UpdatedByContainsFold(*i.UpdatedByContainsFold))
	}
	if i.Name != nil {
		predicates = append(predicates, organization.NameEQ(*i.Name))
	}
	if i.NameNEQ != nil {
		predicates = append(predicates, organization.NameNEQ(*i.NameNEQ))
	}
	if len(i.NameIn) > 0 {
		predicates = append(predicates, organization.NameIn(i.NameIn...))
	}
	if len(i.NameNotIn) > 0 {
		predicates = append(predicates, organization.NameNotIn(i.NameNotIn...))
	}
	if i.NameGT != nil {
		predicates = append(predicates, organization.NameGT(*i.NameGT))
	}
	if i.NameGTE != nil {
		predicates = append(predicates, organization.NameGTE(*i.NameGTE))
	}
	if i.NameLT != nil {
		predicates = append(predicates, organization.NameLT(*i.NameLT))
	}
	if i.NameLTE != nil {
		predicates = append(predicates, organization.NameLTE(*i.NameLTE))
	}
	if i.NameContains != nil {
		predicates = append(predicates, organization.NameContains(*i.NameContains))
	}
	if i.NameHasPrefix != nil {
		predicates = append(predicates, organization.NameHasPrefix(*i.NameHasPrefix))
	}
	if i.NameHasSuffix != nil {
		predicates = append(predicates, organization.NameHasSuffix(*i.NameHasSuffix))
	}
	if i.NameEqualFold != nil {
		predicates = append(predicates, organization.NameEqualFold(*i.NameEqualFold))
	}
	if i.NameContainsFold != nil {
		predicates = append(predicates, organization.NameContainsFold(*i.NameContainsFold))
	}
	if i.Description != nil {
		predicates = append(predicates, organization.DescriptionEQ(*i.Description))
	}
	if i.DescriptionNEQ != nil {
		predicates = append(predicates, organization.DescriptionNEQ(*i.DescriptionNEQ))
	}
	if len(i.DescriptionIn) > 0 {
		predicates = append(predicates, organization.DescriptionIn(i.DescriptionIn...))
	}
	if len(i.DescriptionNotIn) > 0 {
		predicates = append(predicates, organization.DescriptionNotIn(i.DescriptionNotIn...))
	}
	if i.DescriptionGT != nil {
		predicates = append(predicates, organization.DescriptionGT(*i.DescriptionGT))
	}
	if i.DescriptionGTE != nil {
		predicates = append(predicates, organization.DescriptionGTE(*i.DescriptionGTE))
	}
	if i.DescriptionLT != nil {
		predicates = append(predicates, organization.DescriptionLT(*i.DescriptionLT))
	}
	if i.DescriptionLTE != nil {
		predicates = append(predicates, organization.DescriptionLTE(*i.DescriptionLTE))
	}
	if i.DescriptionContains != nil {
		predicates = append(predicates, organization.DescriptionContains(*i.DescriptionContains))
	}
	if i.DescriptionHasPrefix != nil {
		predicates = append(predicates, organization.DescriptionHasPrefix(*i.DescriptionHasPrefix))
	}
	if i.DescriptionHasSuffix != nil {
		predicates = append(predicates, organization.DescriptionHasSuffix(*i.DescriptionHasSuffix))
	}
	if i.DescriptionIsNil {
		predicates = append(predicates, organization.DescriptionIsNil())
	}
	if i.DescriptionNotNil {
		predicates = append(predicates, organization.DescriptionNotNil())
	}
	if i.DescriptionEqualFold != nil {
		predicates = append(predicates, organization.DescriptionEqualFold(*i.DescriptionEqualFold))
	}
	if i.DescriptionContainsFold != nil {
		predicates = append(predicates, organization.DescriptionContainsFold(*i.DescriptionContainsFold))
	}

	switch len(predicates) {
	case 0:
		return nil, ErrEmptyOrganizationWhereInput
	case 1:
		return predicates[0], nil
	default:
		return organization.And(predicates...), nil
	}
}
