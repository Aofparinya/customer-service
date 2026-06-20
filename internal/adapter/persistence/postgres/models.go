package postgres

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type customerModel struct {
	ID                 uuid.UUID      `gorm:"column:id;type:uuid;primaryKey"`
	CustomerNo         string         `gorm:"column:customer_no"`
	CustomerType       string         `gorm:"column:customer_type"`
	Status             string         `gorm:"column:status"`
	FirstName          *string        `gorm:"column:first_name"`
	LastName           *string        `gorm:"column:last_name"`
	CompanyName        *string        `gorm:"column:company_name"`
	RegistrationNumber *string        `gorm:"column:registration_number"`
	Note               *string        `gorm:"column:note"`
	CreatedBy          *uuid.UUID     `gorm:"column:created_by;type:uuid"`
	UpdatedBy          *uuid.UUID     `gorm:"column:updated_by;type:uuid"`
	CreatedAt          time.Time      `gorm:"column:created_at"`
	UpdatedAt          time.Time      `gorm:"column:updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (customerModel) TableName() string { return "customer.customers" }

type addressModel struct {
	ID          uuid.UUID      `gorm:"column:id;type:uuid;primaryKey"`
	CustomerID  uuid.UUID      `gorm:"column:customer_id;type:uuid"`
	AddressType string         `gorm:"column:address_type"`
	Line1       string         `gorm:"column:line1"`
	Line2       *string        `gorm:"column:line2"`
	Subdistrict *string        `gorm:"column:subdistrict"`
	District    *string        `gorm:"column:district"`
	Province    string         `gorm:"column:province"`
	PostalCode  string         `gorm:"column:postal_code"`
	CountryCode string         `gorm:"column:country_code"`
	IsDefault   bool           `gorm:"column:is_default"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (addressModel) TableName() string { return "customer.customer_addresses" }

type contactModel struct {
	ID         uuid.UUID      `gorm:"column:id;type:uuid;primaryKey"`
	CustomerID uuid.UUID      `gorm:"column:customer_id;type:uuid"`
	FirstName  string         `gorm:"column:first_name"`
	LastName   string         `gorm:"column:last_name"`
	Position   *string        `gorm:"column:position"`
	Email      *string        `gorm:"column:email"`
	Phone      *string        `gorm:"column:phone"`
	IsPrimary  bool           `gorm:"column:is_primary"`
	CreatedAt  time.Time      `gorm:"column:created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (contactModel) TableName() string { return "customer.customer_contacts" }

type taxProfileModel struct {
	ID           uuid.UUID      `gorm:"column:id;type:uuid;primaryKey"`
	CustomerID   uuid.UUID      `gorm:"column:customer_id;type:uuid"`
	TaxID        string         `gorm:"column:tax_id"`
	BranchType   string         `gorm:"column:branch_type"`
	BranchCode   string         `gorm:"column:branch_code"`
	BranchName   *string        `gorm:"column:branch_name"`
	AddressLine1 string         `gorm:"column:address_line1"`
	AddressLine2 *string        `gorm:"column:address_line2"`
	Subdistrict  *string        `gorm:"column:subdistrict"`
	District     *string        `gorm:"column:district"`
	Province     string         `gorm:"column:province"`
	PostalCode   string         `gorm:"column:postal_code"`
	CountryCode  string         `gorm:"column:country_code"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (taxProfileModel) TableName() string { return "customer.customer_tax_profiles" }
