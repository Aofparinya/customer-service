package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/saaof/order-platform/customer-service/internal/application"
	"github.com/saaof/order-platform/customer-service/internal/domain"
	"gorm.io/gorm"
)

type NumberIssuer interface {
	NextNumber(context.Context, string) (string, error)
}

type Repository struct {
	db      *gorm.DB
	numbers NumberIssuer
}

func NewRepository(db *gorm.DB, issuers ...NumberIssuer) *Repository {
	var numbers NumberIssuer
	if len(issuers) > 0 {
		numbers = issuers[0]
	}
	return &Repository{db: db, numbers: numbers}
}

func (repository *Repository) CreateCustomer(
	ctx context.Context,
	customer domain.Customer,
) (domain.Customer, error) {
	var result customerModel
	err := repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var customerNo string
		var err error
		if repository.numbers != nil {
			customerNo, err = repository.numbers.NextNumber(ctx, "CUS")
		} else {
			err = tx.Raw(`SELECT 'CUS-' || to_char(CURRENT_DATE, 'YYYYMMDD') || '-' || lpad(nextval('customer.customer_number_seq')::text, 6, '0')`).Scan(&customerNo).Error
		}
		if err != nil {
			return err
		}
		result = customerModelFromDomain(customer)
		result.CustomerNo = customerNo
		return tx.Create(&result).Error
	})
	if err != nil {
		return domain.Customer{}, translateError(err)
	}
	return customerFromModel(result), nil
}

func (repository *Repository) ListCustomers(
	ctx context.Context,
	filter application.ListFilter,
) (application.Page[domain.Customer], error) {
	query := repository.db.WithContext(ctx).Model(&customerModel{})
	if filter.Query != "" {
		search := "%" + strings.ToLower(filter.Query) + "%"
		query = query.Where(`
			lower(customer_no) LIKE ? OR
			lower(COALESCE(first_name, '')) LIKE ? OR
			lower(COALESCE(last_name, '')) LIKE ? OR
			lower(COALESCE(company_name, '')) LIKE ? OR
			lower(COALESCE(registration_number, '')) LIKE ?
		`, search, search, search, search, search)
	}
	if filter.CustomerType != nil {
		query = query.Where("customer_type = ?", *filter.CustomerType)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return application.Page[domain.Customer]{}, translateError(err)
	}
	var models []customerModel
	if err := query.
		Order("created_at DESC").
		Limit(filter.PageSize).
		Offset((filter.Page - 1) * filter.PageSize).
		Find(&models).Error; err != nil {
		return application.Page[domain.Customer]{}, translateError(err)
	}
	customers := make([]domain.Customer, 0, len(models))
	for _, model := range models {
		customers = append(customers, customerFromModel(model))
	}
	return application.NewPage(customers, filter.Page, filter.PageSize, total), nil
}

func (repository *Repository) GetCustomer(
	ctx context.Context,
	customerID uuid.UUID,
) (domain.Customer, error) {
	var model customerModel
	if err := repository.db.WithContext(ctx).First(&model, "id = ?", customerID).Error; err != nil {
		return domain.Customer{}, translateError(err)
	}
	return customerFromModel(model), nil
}

func (repository *Repository) UpdateCustomer(
	ctx context.Context,
	customerID uuid.UUID,
	input application.UpdateCustomerInput,
	actorID uuid.UUID,
) (domain.Customer, error) {
	updates := map[string]any{
		"updated_at": time.Now().UTC(),
		"updated_by": actorID,
	}
	setIfNotNil(updates, "status", input.Status)
	setIfNotNil(updates, "first_name", input.FirstName)
	setIfNotNil(updates, "last_name", input.LastName)
	setIfNotNil(updates, "company_name", input.CompanyName)
	setIfNotNil(updates, "registration_number", input.RegistrationNumber)
	setIfNotNil(updates, "note", input.Note)

	result := repository.db.WithContext(ctx).
		Model(&customerModel{}).
		Where("id = ?", customerID).
		Updates(updates)
	if result.Error != nil {
		return domain.Customer{}, translateError(result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.Customer{}, domain.ErrNotFound
	}
	return repository.GetCustomer(ctx, customerID)
}

func (repository *Repository) DeleteCustomer(
	ctx context.Context,
	customerID uuid.UUID,
	actorID uuid.UUID,
) error {
	now := time.Now().UTC()
	err := repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&customerModel{}).
			Where("id = ?", customerID).
			Updates(map[string]any{
				"deleted_at": now,
				"updated_at": now,
				"updated_by": actorID,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
		}
		for _, model := range []any{&addressModel{}, &contactModel{}, &taxProfileModel{}} {
			if err := tx.Model(model).
				Where("customer_id = ? AND deleted_at IS NULL", customerID).
				Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return translateError(err)
}

func (repository *Repository) CreateAddress(
	ctx context.Context,
	address domain.Address,
) (domain.Address, error) {
	model := addressModelFromDomain(address)
	err := repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if model.IsDefault {
			if err := tx.Model(&addressModel{}).
				Where("customer_id = ? AND address_type = ? AND deleted_at IS NULL", model.CustomerID, model.AddressType).
				Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Create(&model).Error
	})
	if err != nil {
		return domain.Address{}, translateError(err)
	}
	return addressFromModel(model), nil
}

func (repository *Repository) ListAddresses(
	ctx context.Context,
	customerID uuid.UUID,
) ([]domain.Address, error) {
	var models []addressModel
	if err := repository.db.WithContext(ctx).
		Where("customer_id = ?", customerID).
		Order("address_type, is_default DESC, created_at").
		Find(&models).Error; err != nil {
		return nil, translateError(err)
	}
	result := make([]domain.Address, 0, len(models))
	for _, model := range models {
		result = append(result, addressFromModel(model))
	}
	return result, nil
}

func (repository *Repository) UpdateAddress(
	ctx context.Context,
	customerID uuid.UUID,
	addressID uuid.UUID,
	input application.UpdateAddressInput,
) (domain.Address, error) {
	var model addressModel
	err := repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND customer_id = ?", addressID, customerID).First(&model).Error; err != nil {
			return err
		}
		finalType := model.AddressType
		if input.AddressType != nil {
			finalType = string(*input.AddressType)
		}
		finalDefault := model.IsDefault
		if input.IsDefault != nil {
			finalDefault = *input.IsDefault
		}
		if finalDefault {
			if err := tx.Model(&addressModel{}).
				Where("customer_id = ? AND address_type = ? AND id <> ? AND deleted_at IS NULL", customerID, finalType, addressID).
				Update("is_default", false).Error; err != nil {
				return err
			}
		}
		updates := map[string]any{"updated_at": time.Now().UTC()}
		setIfNotNil(updates, "address_type", input.AddressType)
		setIfNotNil(updates, "line1", input.Line1)
		setIfNotNil(updates, "line2", input.Line2)
		setIfNotNil(updates, "subdistrict", input.Subdistrict)
		setIfNotNil(updates, "district", input.District)
		setIfNotNil(updates, "province", input.Province)
		setIfNotNil(updates, "postal_code", input.PostalCode)
		setIfNotNil(updates, "country_code", input.CountryCode)
		setIfNotNil(updates, "is_default", input.IsDefault)
		if err := tx.Model(&model).Updates(updates).Error; err != nil {
			return err
		}
		return tx.First(&model, "id = ?", addressID).Error
	})
	if err != nil {
		return domain.Address{}, translateError(err)
	}
	return addressFromModel(model), nil
}

func (repository *Repository) DeleteAddress(
	ctx context.Context,
	customerID uuid.UUID,
	addressID uuid.UUID,
) error {
	return repository.softDeleteChild(ctx, &addressModel{}, customerID, addressID)
}

func (repository *Repository) CreateContact(
	ctx context.Context,
	contact domain.Contact,
) (domain.Contact, error) {
	model := contactModelFromDomain(contact)
	err := repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if model.IsPrimary {
			if err := tx.Model(&contactModel{}).
				Where("customer_id = ? AND deleted_at IS NULL", model.CustomerID).
				Update("is_primary", false).Error; err != nil {
				return err
			}
		}
		return tx.Create(&model).Error
	})
	if err != nil {
		return domain.Contact{}, translateError(err)
	}
	return contactFromModel(model), nil
}

func (repository *Repository) ListContacts(
	ctx context.Context,
	customerID uuid.UUID,
) ([]domain.Contact, error) {
	var models []contactModel
	if err := repository.db.WithContext(ctx).
		Where("customer_id = ?", customerID).
		Order("is_primary DESC, created_at").
		Find(&models).Error; err != nil {
		return nil, translateError(err)
	}
	result := make([]domain.Contact, 0, len(models))
	for _, model := range models {
		result = append(result, contactFromModel(model))
	}
	return result, nil
}

func (repository *Repository) UpdateContact(
	ctx context.Context,
	customerID uuid.UUID,
	contactID uuid.UUID,
	input application.UpdateContactInput,
) (domain.Contact, error) {
	var model contactModel
	err := repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND customer_id = ?", contactID, customerID).First(&model).Error; err != nil {
			return err
		}
		if input.IsPrimary != nil && *input.IsPrimary {
			if err := tx.Model(&contactModel{}).
				Where("customer_id = ? AND id <> ? AND deleted_at IS NULL", customerID, contactID).
				Update("is_primary", false).Error; err != nil {
				return err
			}
		}
		updates := map[string]any{"updated_at": time.Now().UTC()}
		setIfNotNil(updates, "first_name", input.FirstName)
		setIfNotNil(updates, "last_name", input.LastName)
		setIfNotNil(updates, "position", input.Position)
		setIfNotNil(updates, "email", input.Email)
		setIfNotNil(updates, "phone", input.Phone)
		setIfNotNil(updates, "is_primary", input.IsPrimary)
		if err := tx.Model(&model).Updates(updates).Error; err != nil {
			return err
		}
		return tx.First(&model, "id = ?", contactID).Error
	})
	if err != nil {
		return domain.Contact{}, translateError(err)
	}
	return contactFromModel(model), nil
}

func (repository *Repository) DeleteContact(
	ctx context.Context,
	customerID uuid.UUID,
	contactID uuid.UUID,
) error {
	return repository.softDeleteChild(ctx, &contactModel{}, customerID, contactID)
}

func (repository *Repository) CreateTaxProfile(
	ctx context.Context,
	profile domain.TaxProfile,
) (domain.TaxProfile, error) {
	model := taxProfileModelFromDomain(profile)
	if err := repository.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.TaxProfile{}, translateError(err)
	}
	return taxProfileFromModel(model), nil
}

func (repository *Repository) ListTaxProfiles(
	ctx context.Context,
	customerID uuid.UUID,
) ([]domain.TaxProfile, error) {
	var models []taxProfileModel
	if err := repository.db.WithContext(ctx).
		Where("customer_id = ?", customerID).
		Order("branch_type, branch_code").
		Find(&models).Error; err != nil {
		return nil, translateError(err)
	}
	result := make([]domain.TaxProfile, 0, len(models))
	for _, model := range models {
		result = append(result, taxProfileFromModel(model))
	}
	return result, nil
}

func (repository *Repository) UpdateTaxProfile(
	ctx context.Context,
	customerID uuid.UUID,
	taxProfileID uuid.UUID,
	input application.UpdateTaxProfileInput,
) (domain.TaxProfile, error) {
	updates := map[string]any{"updated_at": time.Now().UTC()}
	setIfNotNil(updates, "tax_id", input.TaxID)
	setIfNotNil(updates, "branch_type", input.BranchType)
	setIfNotNil(updates, "branch_code", input.BranchCode)
	setIfNotNil(updates, "branch_name", input.BranchName)
	setIfNotNil(updates, "address_line1", input.AddressLine1)
	setIfNotNil(updates, "address_line2", input.AddressLine2)
	setIfNotNil(updates, "subdistrict", input.Subdistrict)
	setIfNotNil(updates, "district", input.District)
	setIfNotNil(updates, "province", input.Province)
	setIfNotNil(updates, "postal_code", input.PostalCode)
	setIfNotNil(updates, "country_code", input.CountryCode)

	result := repository.db.WithContext(ctx).
		Model(&taxProfileModel{}).
		Where("id = ? AND customer_id = ?", taxProfileID, customerID).
		Updates(updates)
	if result.Error != nil {
		return domain.TaxProfile{}, translateError(result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.TaxProfile{}, domain.ErrNotFound
	}
	var model taxProfileModel
	if err := repository.db.WithContext(ctx).First(&model, "id = ?", taxProfileID).Error; err != nil {
		return domain.TaxProfile{}, translateError(err)
	}
	return taxProfileFromModel(model), nil
}

func (repository *Repository) DeleteTaxProfile(
	ctx context.Context,
	customerID uuid.UUID,
	taxProfileID uuid.UUID,
) error {
	return repository.softDeleteChild(ctx, &taxProfileModel{}, customerID, taxProfileID)
}

func (repository *Repository) softDeleteChild(
	ctx context.Context,
	model any,
	customerID uuid.UUID,
	id uuid.UUID,
) error {
	result := repository.db.WithContext(ctx).
		Model(model).
		Where("id = ? AND customer_id = ?", id, customerID).
		Updates(map[string]any{
			"deleted_at": time.Now().UTC(),
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return translateError(result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func translateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrInvalid) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrNotFound
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23505":
			return fmt.Errorf("%w: duplicate value", domain.ErrConflict)
		case "23514", "23502":
			return domain.ValidationError{Message: postgresError.Message}
		}
	}
	return err
}

func setIfNotNil[T any](updates map[string]any, key string, value *T) {
	if value != nil {
		updates[key] = *value
	}
}

func customerModelFromDomain(customer domain.Customer) customerModel {
	return customerModel{
		ID:                 customer.ID,
		CustomerNo:         customer.CustomerNo,
		CustomerType:       string(customer.CustomerType),
		Status:             string(customer.Status),
		FirstName:          customer.FirstName,
		LastName:           customer.LastName,
		CompanyName:        customer.CompanyName,
		RegistrationNumber: customer.RegistrationNumber,
		Note:               customer.Note,
		CreatedBy:          customer.CreatedBy,
		UpdatedBy:          customer.UpdatedBy,
	}
}

func customerFromModel(model customerModel) domain.Customer {
	return domain.Customer{
		ID:                 model.ID,
		CustomerNo:         model.CustomerNo,
		CustomerType:       domain.CustomerType(model.CustomerType),
		Status:             domain.CustomerStatus(model.Status),
		FirstName:          model.FirstName,
		LastName:           model.LastName,
		CompanyName:        model.CompanyName,
		RegistrationNumber: model.RegistrationNumber,
		Note:               model.Note,
		CreatedBy:          model.CreatedBy,
		UpdatedBy:          model.UpdatedBy,
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
	}
}

func addressModelFromDomain(address domain.Address) addressModel {
	return addressModel{
		ID:          address.ID,
		CustomerID:  address.CustomerID,
		AddressType: string(address.AddressType),
		Line1:       address.Line1,
		Line2:       address.Line2,
		Subdistrict: address.Subdistrict,
		District:    address.District,
		Province:    address.Province,
		PostalCode:  address.PostalCode,
		CountryCode: address.CountryCode,
		IsDefault:   address.IsDefault,
	}
}

func addressFromModel(model addressModel) domain.Address {
	return domain.Address{
		ID:          model.ID,
		CustomerID:  model.CustomerID,
		AddressType: domain.AddressType(model.AddressType),
		Line1:       model.Line1,
		Line2:       model.Line2,
		Subdistrict: model.Subdistrict,
		District:    model.District,
		Province:    model.Province,
		PostalCode:  model.PostalCode,
		CountryCode: model.CountryCode,
		IsDefault:   model.IsDefault,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

func contactModelFromDomain(contact domain.Contact) contactModel {
	return contactModel{
		ID:         contact.ID,
		CustomerID: contact.CustomerID,
		FirstName:  contact.FirstName,
		LastName:   contact.LastName,
		Position:   contact.Position,
		Email:      contact.Email,
		Phone:      contact.Phone,
		IsPrimary:  contact.IsPrimary,
	}
}

func contactFromModel(model contactModel) domain.Contact {
	return domain.Contact{
		ID:         model.ID,
		CustomerID: model.CustomerID,
		FirstName:  model.FirstName,
		LastName:   model.LastName,
		Position:   model.Position,
		Email:      model.Email,
		Phone:      model.Phone,
		IsPrimary:  model.IsPrimary,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}
}

func taxProfileModelFromDomain(profile domain.TaxProfile) taxProfileModel {
	return taxProfileModel{
		ID:           profile.ID,
		CustomerID:   profile.CustomerID,
		TaxID:        profile.TaxID,
		BranchType:   string(profile.BranchType),
		BranchCode:   profile.BranchCode,
		BranchName:   profile.BranchName,
		AddressLine1: profile.AddressLine1,
		AddressLine2: profile.AddressLine2,
		Subdistrict:  profile.Subdistrict,
		District:     profile.District,
		Province:     profile.Province,
		PostalCode:   profile.PostalCode,
		CountryCode:  profile.CountryCode,
	}
}

func taxProfileFromModel(model taxProfileModel) domain.TaxProfile {
	return domain.TaxProfile{
		ID:           model.ID,
		CustomerID:   model.CustomerID,
		TaxID:        model.TaxID,
		BranchType:   domain.BranchType(model.BranchType),
		BranchCode:   model.BranchCode,
		BranchName:   model.BranchName,
		AddressLine1: model.AddressLine1,
		AddressLine2: model.AddressLine2,
		Subdistrict:  model.Subdistrict,
		District:     model.District,
		Province:     model.Province,
		PostalCode:   model.PostalCode,
		CountryCode:  model.CountryCode,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}
