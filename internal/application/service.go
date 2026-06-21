package application

import (
	"context"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/saaof/order-platform/customer-service/internal/domain"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (service *Service) CreateCustomer(
	ctx context.Context,
	input CreateCustomerInput,
	actorID uuid.UUID,
) (domain.Customer, error) {
	normalizeCustomerInput(&input)
	if input.Status == "" {
		input.Status = domain.CustomerStatusActive
	}
	if err := validateCustomerNames(input.CustomerType, input.FirstName, input.LastName, input.CompanyName); err != nil {
		return domain.Customer{}, err
	}
	customerID := uuid.New()
	if input.ID != nil && *input.ID != uuid.Nil {
		customerID = *input.ID
	}
	customer := domain.Customer{
		ID:                 customerID,
		CustomerType:       input.CustomerType,
		Status:             input.Status,
		FirstName:          input.FirstName,
		LastName:           input.LastName,
		CompanyName:        input.CompanyName,
		RegistrationNumber: input.RegistrationNumber,
		Note:               input.Note,
		CreatedBy:          &actorID,
		UpdatedBy:          &actorID,
	}
	return service.repository.CreateCustomer(ctx, customer)
}

func (service *Service) ListCustomers(
	ctx context.Context,
	filter ListFilter,
) (Page[domain.Customer], error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	filter.Query = strings.TrimSpace(filter.Query)
	return service.repository.ListCustomers(ctx, filter)
}

func (service *Service) GetCustomer(
	ctx context.Context,
	customerID uuid.UUID,
) (domain.Customer, error) {
	return service.repository.GetCustomer(ctx, customerID)
}

func (service *Service) UpdateCustomer(
	ctx context.Context,
	customerID uuid.UUID,
	input UpdateCustomerInput,
	actorID uuid.UUID,
) (domain.Customer, error) {
	existing, err := service.repository.GetCustomer(ctx, customerID)
	if err != nil {
		return domain.Customer{}, err
	}
	normalizeUpdateCustomerInput(&input)
	firstName := chooseString(input.FirstName, existing.FirstName)
	lastName := chooseString(input.LastName, existing.LastName)
	companyName := chooseString(input.CompanyName, existing.CompanyName)
	if err := validateCustomerNames(existing.CustomerType, firstName, lastName, companyName); err != nil {
		return domain.Customer{}, err
	}
	return service.repository.UpdateCustomer(ctx, customerID, input, actorID)
}

func (service *Service) DeleteCustomer(
	ctx context.Context,
	customerID uuid.UUID,
	actorID uuid.UUID,
) error {
	return service.repository.DeleteCustomer(ctx, customerID, actorID)
}

func (service *Service) CreateAddress(
	ctx context.Context,
	customerID uuid.UUID,
	input CreateAddressInput,
) (domain.Address, error) {
	if _, err := service.repository.GetCustomer(ctx, customerID); err != nil {
		return domain.Address{}, err
	}
	input.Line1 = strings.TrimSpace(input.Line1)
	input.Province = strings.TrimSpace(input.Province)
	input.PostalCode = strings.TrimSpace(input.PostalCode)
	input.CountryCode = normalizeCountryCode(input.CountryCode)
	return service.repository.CreateAddress(ctx, domain.Address{
		ID:          uuid.New(),
		CustomerID:  customerID,
		AddressType: input.AddressType,
		Line1:       input.Line1,
		Line2:       trimPointer(input.Line2),
		Subdistrict: trimPointer(input.Subdistrict),
		District:    trimPointer(input.District),
		Province:    input.Province,
		PostalCode:  input.PostalCode,
		CountryCode: input.CountryCode,
		IsDefault:   input.IsDefault,
	})
}

func (service *Service) ListAddresses(
	ctx context.Context,
	customerID uuid.UUID,
) ([]domain.Address, error) {
	if _, err := service.repository.GetCustomer(ctx, customerID); err != nil {
		return nil, err
	}
	return service.repository.ListAddresses(ctx, customerID)
}

func (service *Service) UpdateAddress(
	ctx context.Context,
	customerID uuid.UUID,
	addressID uuid.UUID,
	input UpdateAddressInput,
) (domain.Address, error) {
	if input.CountryCode != nil {
		value := normalizeCountryCode(*input.CountryCode)
		input.CountryCode = &value
	}
	return service.repository.UpdateAddress(ctx, customerID, addressID, input)
}

func (service *Service) DeleteAddress(
	ctx context.Context,
	customerID uuid.UUID,
	addressID uuid.UUID,
) error {
	return service.repository.DeleteAddress(ctx, customerID, addressID)
}

func (service *Service) CreateContact(
	ctx context.Context,
	customerID uuid.UUID,
	input CreateContactInput,
) (domain.Contact, error) {
	if _, err := service.repository.GetCustomer(ctx, customerID); err != nil {
		return domain.Contact{}, err
	}
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)
	input.Email = lowerPointer(input.Email)
	input.Phone = trimPointer(input.Phone)
	if input.Email == nil && input.Phone == nil {
		return domain.Contact{}, domain.ValidationError{Message: "email or phone is required"}
	}
	return service.repository.CreateContact(ctx, domain.Contact{
		ID:         uuid.New(),
		CustomerID: customerID,
		FirstName:  input.FirstName,
		LastName:   input.LastName,
		Position:   trimPointer(input.Position),
		Email:      input.Email,
		Phone:      input.Phone,
		IsPrimary:  input.IsPrimary,
	})
}

func (service *Service) ListContacts(
	ctx context.Context,
	customerID uuid.UUID,
) ([]domain.Contact, error) {
	if _, err := service.repository.GetCustomer(ctx, customerID); err != nil {
		return nil, err
	}
	return service.repository.ListContacts(ctx, customerID)
}

func (service *Service) UpdateContact(
	ctx context.Context,
	customerID uuid.UUID,
	contactID uuid.UUID,
	input UpdateContactInput,
) (domain.Contact, error) {
	input.Email = lowerPointer(input.Email)
	input.Phone = trimPointer(input.Phone)
	return service.repository.UpdateContact(ctx, customerID, contactID, input)
}

func (service *Service) DeleteContact(
	ctx context.Context,
	customerID uuid.UUID,
	contactID uuid.UUID,
) error {
	return service.repository.DeleteContact(ctx, customerID, contactID)
}

func (service *Service) CreateTaxProfile(
	ctx context.Context,
	customerID uuid.UUID,
	input CreateTaxProfileInput,
) (domain.TaxProfile, error) {
	if _, err := service.repository.GetCustomer(ctx, customerID); err != nil {
		return domain.TaxProfile{}, err
	}
	normalizeCreateTaxProfile(&input)
	if err := validateBranch(input.BranchType, input.BranchCode); err != nil {
		return domain.TaxProfile{}, err
	}
	return service.repository.CreateTaxProfile(ctx, domain.TaxProfile{
		ID:           uuid.New(),
		CustomerID:   customerID,
		TaxID:        input.TaxID,
		BranchType:   input.BranchType,
		BranchCode:   input.BranchCode,
		BranchName:   trimPointer(input.BranchName),
		AddressLine1: input.AddressLine1,
		AddressLine2: trimPointer(input.AddressLine2),
		Subdistrict:  trimPointer(input.Subdistrict),
		District:     trimPointer(input.District),
		Province:     input.Province,
		PostalCode:   input.PostalCode,
		CountryCode:  input.CountryCode,
	})
}

func (service *Service) ListTaxProfiles(
	ctx context.Context,
	customerID uuid.UUID,
) ([]domain.TaxProfile, error) {
	if _, err := service.repository.GetCustomer(ctx, customerID); err != nil {
		return nil, err
	}
	return service.repository.ListTaxProfiles(ctx, customerID)
}

func (service *Service) UpdateTaxProfile(
	ctx context.Context,
	customerID uuid.UUID,
	taxProfileID uuid.UUID,
	input UpdateTaxProfileInput,
) (domain.TaxProfile, error) {
	if input.CountryCode != nil {
		value := normalizeCountryCode(*input.CountryCode)
		input.CountryCode = &value
	}
	return service.repository.UpdateTaxProfile(ctx, customerID, taxProfileID, input)
}

func (service *Service) DeleteTaxProfile(
	ctx context.Context,
	customerID uuid.UUID,
	taxProfileID uuid.UUID,
) error {
	return service.repository.DeleteTaxProfile(ctx, customerID, taxProfileID)
}

func NewPage[T any](data []T, page, pageSize int, total int64) Page[T] {
	totalPages := int64(0)
	if total > 0 {
		totalPages = int64(math.Ceil(float64(total) / float64(pageSize)))
	}
	return Page[T]{
		Data: data,
		Pagination: Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}

func normalizeCustomerInput(input *CreateCustomerInput) {
	input.FirstName = trimPointer(input.FirstName)
	input.LastName = trimPointer(input.LastName)
	input.CompanyName = trimPointer(input.CompanyName)
	input.RegistrationNumber = trimPointer(input.RegistrationNumber)
	input.Note = trimPointer(input.Note)
}

func normalizeUpdateCustomerInput(input *UpdateCustomerInput) {
	input.FirstName = trimPointer(input.FirstName)
	input.LastName = trimPointer(input.LastName)
	input.CompanyName = trimPointer(input.CompanyName)
	input.RegistrationNumber = trimPointer(input.RegistrationNumber)
	input.Note = trimPointer(input.Note)
}

func validateCustomerNames(
	customerType domain.CustomerType,
	firstName *string,
	lastName *string,
	companyName *string,
) error {
	switch customerType {
	case domain.CustomerTypeIndividual:
		if firstName == nil || lastName == nil {
			return domain.ValidationError{Message: "firstName and lastName are required for INDIVIDUAL"}
		}
	case domain.CustomerTypeCorporate:
		if companyName == nil {
			return domain.ValidationError{Message: "companyName is required for CORPORATE"}
		}
	default:
		return domain.ValidationError{Message: "invalid customerType"}
	}
	return nil
}

func normalizeCreateTaxProfile(input *CreateTaxProfileInput) {
	input.TaxID = strings.TrimSpace(input.TaxID)
	input.BranchCode = strings.TrimSpace(input.BranchCode)
	if input.BranchType == domain.BranchTypeHeadOffice {
		input.BranchCode = "00000"
	}
	input.AddressLine1 = strings.TrimSpace(input.AddressLine1)
	input.Province = strings.TrimSpace(input.Province)
	input.PostalCode = strings.TrimSpace(input.PostalCode)
	input.CountryCode = normalizeCountryCode(input.CountryCode)
}

func validateBranch(branchType domain.BranchType, branchCode string) error {
	if branchType == domain.BranchTypeHeadOffice && branchCode != "00000" {
		return domain.ValidationError{Message: "HEAD_OFFICE branchCode must be 00000"}
	}
	if branchType == domain.BranchTypeBranch && (branchCode == "" || branchCode == "00000") {
		return domain.ValidationError{Message: "BRANCH requires a branchCode other than 00000"}
	}
	return nil
}

func normalizeCountryCode(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return "TH"
	}
	return value
}

func trimPointer(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func lowerPointer(value *string) *string {
	trimmed := trimPointer(value)
	if trimmed == nil {
		return nil
	}
	lower := strings.ToLower(*trimmed)
	return &lower
}

func chooseString(update *string, existing *string) *string {
	if update != nil {
		return update
	}
	return existing
}
