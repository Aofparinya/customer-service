package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/saaof/order-platform/customer-service/internal/domain"
)

type ListFilter struct {
	Query        string
	CustomerType *domain.CustomerType
	Status       *domain.CustomerStatus
	Page         int
	PageSize     int
}

type Page[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"totalPages"`
}

type CreateCustomerInput struct {
	ID                 *uuid.UUID            `json:"id"`
	CustomerType       domain.CustomerType   `json:"customerType" validate:"required,oneof=INDIVIDUAL CORPORATE"`
	Status             domain.CustomerStatus `json:"status" validate:"omitempty,oneof=ACTIVE INACTIVE BLOCKED"`
	FirstName          *string               `json:"firstName"`
	LastName           *string               `json:"lastName"`
	CompanyName        *string               `json:"companyName"`
	RegistrationNumber *string               `json:"registrationNumber"`
	Note               *string               `json:"note"`
}

type UpdateCustomerInput struct {
	Status             *domain.CustomerStatus `json:"status" validate:"omitempty,oneof=ACTIVE INACTIVE BLOCKED"`
	FirstName          *string                `json:"firstName"`
	LastName           *string                `json:"lastName"`
	CompanyName        *string                `json:"companyName"`
	RegistrationNumber *string                `json:"registrationNumber"`
	Note               *string                `json:"note"`
}

type CreateAddressInput struct {
	AddressType domain.AddressType `json:"addressType" validate:"required,oneof=BILLING SHIPPING CONTACT"`
	Line1       string             `json:"line1" validate:"required,max=255"`
	Line2       *string            `json:"line2"`
	Subdistrict *string            `json:"subdistrict"`
	District    *string            `json:"district"`
	Province    string             `json:"province" validate:"required,max=150"`
	PostalCode  string             `json:"postalCode" validate:"required,max=20"`
	CountryCode string             `json:"countryCode" validate:"omitempty,len=2"`
	IsDefault   bool               `json:"isDefault"`
}

type UpdateAddressInput struct {
	AddressType *domain.AddressType `json:"addressType" validate:"omitempty,oneof=BILLING SHIPPING CONTACT"`
	Line1       *string             `json:"line1" validate:"omitempty,max=255"`
	Line2       *string             `json:"line2"`
	Subdistrict *string             `json:"subdistrict"`
	District    *string             `json:"district"`
	Province    *string             `json:"province" validate:"omitempty,max=150"`
	PostalCode  *string             `json:"postalCode" validate:"omitempty,max=20"`
	CountryCode *string             `json:"countryCode" validate:"omitempty,len=2"`
	IsDefault   *bool               `json:"isDefault"`
}

type CreateContactInput struct {
	FirstName string  `json:"firstName" validate:"required,max=150"`
	LastName  string  `json:"lastName" validate:"required,max=150"`
	Position  *string `json:"position"`
	Email     *string `json:"email" validate:"omitempty,email"`
	Phone     *string `json:"phone"`
	IsPrimary bool    `json:"isPrimary"`
}

type UpdateContactInput struct {
	FirstName *string `json:"firstName" validate:"omitempty,max=150"`
	LastName  *string `json:"lastName" validate:"omitempty,max=150"`
	Position  *string `json:"position"`
	Email     *string `json:"email" validate:"omitempty,email"`
	Phone     *string `json:"phone"`
	IsPrimary *bool   `json:"isPrimary"`
}

type CreateTaxProfileInput struct {
	TaxID        string            `json:"taxId" validate:"required,max=50"`
	BranchType   domain.BranchType `json:"branchType" validate:"required,oneof=HEAD_OFFICE BRANCH"`
	BranchCode   string            `json:"branchCode" validate:"omitempty,max=20"`
	BranchName   *string           `json:"branchName"`
	AddressLine1 string            `json:"addressLine1" validate:"required,max=255"`
	AddressLine2 *string           `json:"addressLine2"`
	Subdistrict  *string           `json:"subdistrict"`
	District     *string           `json:"district"`
	Province     string            `json:"province" validate:"required,max=150"`
	PostalCode   string            `json:"postalCode" validate:"required,max=20"`
	CountryCode  string            `json:"countryCode" validate:"omitempty,len=2"`
}

type UpdateTaxProfileInput struct {
	TaxID        *string            `json:"taxId" validate:"omitempty,max=50"`
	BranchType   *domain.BranchType `json:"branchType" validate:"omitempty,oneof=HEAD_OFFICE BRANCH"`
	BranchCode   *string            `json:"branchCode" validate:"omitempty,max=20"`
	BranchName   *string            `json:"branchName"`
	AddressLine1 *string            `json:"addressLine1" validate:"omitempty,max=255"`
	AddressLine2 *string            `json:"addressLine2"`
	Subdistrict  *string            `json:"subdistrict"`
	District     *string            `json:"district"`
	Province     *string            `json:"province" validate:"omitempty,max=150"`
	PostalCode   *string            `json:"postalCode" validate:"omitempty,max=20"`
	CountryCode  *string            `json:"countryCode" validate:"omitempty,len=2"`
}

type Repository interface {
	CreateCustomer(context.Context, domain.Customer) (domain.Customer, error)
	ListCustomers(context.Context, ListFilter) (Page[domain.Customer], error)
	GetCustomer(context.Context, uuid.UUID) (domain.Customer, error)
	UpdateCustomer(context.Context, uuid.UUID, UpdateCustomerInput, uuid.UUID) (domain.Customer, error)
	DeleteCustomer(context.Context, uuid.UUID, uuid.UUID) error

	CreateAddress(context.Context, domain.Address) (domain.Address, error)
	ListAddresses(context.Context, uuid.UUID) ([]domain.Address, error)
	UpdateAddress(context.Context, uuid.UUID, uuid.UUID, UpdateAddressInput) (domain.Address, error)
	DeleteAddress(context.Context, uuid.UUID, uuid.UUID) error

	CreateContact(context.Context, domain.Contact) (domain.Contact, error)
	ListContacts(context.Context, uuid.UUID) ([]domain.Contact, error)
	UpdateContact(context.Context, uuid.UUID, uuid.UUID, UpdateContactInput) (domain.Contact, error)
	DeleteContact(context.Context, uuid.UUID, uuid.UUID) error

	CreateTaxProfile(context.Context, domain.TaxProfile) (domain.TaxProfile, error)
	ListTaxProfiles(context.Context, uuid.UUID) ([]domain.TaxProfile, error)
	UpdateTaxProfile(context.Context, uuid.UUID, uuid.UUID, UpdateTaxProfileInput) (domain.TaxProfile, error)
	DeleteTaxProfile(context.Context, uuid.UUID, uuid.UUID) error
}
