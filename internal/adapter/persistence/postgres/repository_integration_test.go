package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/saaof/order-platform/customer-service/internal/application"
	"github.com/saaof/order-platform/customer-service/internal/domain"
	"github.com/saaof/order-platform/customer-service/internal/infrastructure/database"
)

func TestRepositoryCustomerLifecycle(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	db, err := database.Open(databaseURL, false)
	if err != nil {
		t.Fatalf("database.Open() error = %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("database.Migrate() error = %v", err)
	}
	repository := NewRepository(db)
	actorID := uuid.New()
	firstName := "Integration"
	lastName := "Customer"

	customer, err := repository.CreateCustomer(context.Background(), domain.Customer{
		ID:           uuid.New(),
		CustomerType: domain.CustomerTypeIndividual,
		Status:       domain.CustomerStatusActive,
		FirstName:    &firstName,
		LastName:     &lastName,
		CreatedBy:    &actorID,
		UpdatedBy:    &actorID,
	})
	if err != nil {
		t.Fatalf("CreateCustomer() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Exec("DELETE FROM customer.customer_addresses WHERE customer_id = ?", customer.ID).Error
		_ = db.Exec("DELETE FROM customer.customer_contacts WHERE customer_id = ?", customer.ID).Error
		_ = db.Exec("DELETE FROM customer.customer_tax_profiles WHERE customer_id = ?", customer.ID).Error
		_ = db.Exec("DELETE FROM customer.customers WHERE id = ?", customer.ID).Error
	})
	if customer.CustomerNo == "" {
		t.Fatal("expected generated customer number")
	}

	page, err := repository.ListCustomers(context.Background(), application.ListFilter{
		Query:    customer.CustomerNo,
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("ListCustomers() error = %v", err)
	}
	if page.Pagination.Total < 1 {
		t.Fatal("expected customer in search result")
	}

	firstAddress, err := repository.CreateAddress(context.Background(), domain.Address{
		ID:          uuid.New(),
		CustomerID:  customer.ID,
		AddressType: domain.AddressTypeBilling,
		Line1:       "First address",
		Province:    "Bangkok",
		PostalCode:  "10110",
		CountryCode: "TH",
		IsDefault:   true,
	})
	if err != nil {
		t.Fatalf("CreateAddress(first) error = %v", err)
	}
	secondAddress, err := repository.CreateAddress(context.Background(), domain.Address{
		ID:          uuid.New(),
		CustomerID:  customer.ID,
		AddressType: domain.AddressTypeBilling,
		Line1:       "Second address",
		Province:    "Bangkok",
		PostalCode:  "10110",
		CountryCode: "TH",
		IsDefault:   true,
	})
	if err != nil {
		t.Fatalf("CreateAddress(second) error = %v", err)
	}
	addresses, err := repository.ListAddresses(context.Background(), customer.ID)
	if err != nil {
		t.Fatalf("ListAddresses() error = %v", err)
	}
	defaults := 0
	for _, address := range addresses {
		if address.IsDefault {
			defaults++
			if address.ID != secondAddress.ID {
				t.Fatalf("expected second address to be default, got %s", address.ID)
			}
		}
	}
	if defaults != 1 {
		t.Fatalf("expected one default address, got %d; first address %s", defaults, firstAddress.ID)
	}
}
