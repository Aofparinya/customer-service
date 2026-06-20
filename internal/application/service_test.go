package application

import (
	"testing"

	"github.com/saaof/order-platform/customer-service/internal/domain"
)

func TestValidateCustomerNames(t *testing.T) {
	firstName := "Parinya"
	lastName := "Sakulsantitham"
	company := "Order Platform Co., Ltd."

	tests := []struct {
		name         string
		customerType domain.CustomerType
		firstName    *string
		lastName     *string
		companyName  *string
		wantError    bool
	}{
		{"individual valid", domain.CustomerTypeIndividual, &firstName, &lastName, nil, false},
		{"individual missing name", domain.CustomerTypeIndividual, nil, &lastName, nil, true},
		{"corporate valid", domain.CustomerTypeCorporate, nil, nil, &company, false},
		{"corporate missing company", domain.CustomerTypeCorporate, nil, nil, nil, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateCustomerNames(
				test.customerType,
				test.firstName,
				test.lastName,
				test.companyName,
			)
			if (err != nil) != test.wantError {
				t.Fatalf("validateCustomerNames() error = %v, wantError %v", err, test.wantError)
			}
		})
	}
}

func TestValidateBranch(t *testing.T) {
	if err := validateBranch(domain.BranchTypeHeadOffice, "00000"); err != nil {
		t.Fatalf("expected valid head office: %v", err)
	}
	if err := validateBranch(domain.BranchTypeBranch, "00001"); err != nil {
		t.Fatalf("expected valid branch: %v", err)
	}
	if err := validateBranch(domain.BranchTypeBranch, "00000"); err == nil {
		t.Fatal("expected branch validation error")
	}
}
