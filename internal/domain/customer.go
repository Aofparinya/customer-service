package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type CustomerType string
type CustomerStatus string
type AddressType string
type BranchType string

const (
	CustomerTypeIndividual CustomerType = "INDIVIDUAL"
	CustomerTypeCorporate  CustomerType = "CORPORATE"

	CustomerStatusActive   CustomerStatus = "ACTIVE"
	CustomerStatusInactive CustomerStatus = "INACTIVE"
	CustomerStatusBlocked  CustomerStatus = "BLOCKED"

	AddressTypeBilling  AddressType = "BILLING"
	AddressTypeShipping AddressType = "SHIPPING"
	AddressTypeContact  AddressType = "CONTACT"

	BranchTypeHeadOffice BranchType = "HEAD_OFFICE"
	BranchTypeBranch     BranchType = "BRANCH"
)

type Customer struct {
	ID                 uuid.UUID      `json:"id"`
	CustomerNo         string         `json:"customerNo"`
	CustomerType       CustomerType   `json:"customerType"`
	Status             CustomerStatus `json:"status"`
	FirstName          *string        `json:"firstName,omitempty"`
	LastName           *string        `json:"lastName,omitempty"`
	CompanyName        *string        `json:"companyName,omitempty"`
	RegistrationNumber *string        `json:"registrationNumber,omitempty"`
	Note               *string        `json:"note,omitempty"`
	CreatedBy          *uuid.UUID     `json:"createdBy,omitempty"`
	UpdatedBy          *uuid.UUID     `json:"updatedBy,omitempty"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
}

func (customer Customer) DisplayName() string {
	if customer.CustomerType == CustomerTypeCorporate && customer.CompanyName != nil {
		return *customer.CompanyName
	}
	parts := make([]string, 0, 2)
	if customer.FirstName != nil {
		parts = append(parts, *customer.FirstName)
	}
	if customer.LastName != nil {
		parts = append(parts, *customer.LastName)
	}
	return strings.Join(parts, " ")
}

type Address struct {
	ID          uuid.UUID   `json:"id"`
	CustomerID  uuid.UUID   `json:"customerId"`
	AddressType AddressType `json:"addressType"`
	Line1       string      `json:"line1"`
	Line2       *string     `json:"line2,omitempty"`
	Subdistrict *string     `json:"subdistrict,omitempty"`
	District    *string     `json:"district,omitempty"`
	Province    string      `json:"province"`
	PostalCode  string      `json:"postalCode"`
	CountryCode string      `json:"countryCode"`
	IsDefault   bool        `json:"isDefault"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

type Contact struct {
	ID         uuid.UUID `json:"id"`
	CustomerID uuid.UUID `json:"customerId"`
	FirstName  string    `json:"firstName"`
	LastName   string    `json:"lastName"`
	Position   *string   `json:"position,omitempty"`
	Email      *string   `json:"email,omitempty"`
	Phone      *string   `json:"phone,omitempty"`
	IsPrimary  bool      `json:"isPrimary"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type TaxProfile struct {
	ID           uuid.UUID  `json:"id"`
	CustomerID   uuid.UUID  `json:"customerId"`
	TaxID        string     `json:"taxId"`
	BranchType   BranchType `json:"branchType"`
	BranchCode   string     `json:"branchCode"`
	BranchName   *string    `json:"branchName,omitempty"`
	AddressLine1 string     `json:"addressLine1"`
	AddressLine2 *string    `json:"addressLine2,omitempty"`
	Subdistrict  *string    `json:"subdistrict,omitempty"`
	District     *string    `json:"district,omitempty"`
	Province     string     `json:"province"`
	PostalCode   string     `json:"postalCode"`
	CountryCode  string     `json:"countryCode"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}
