CREATE SEQUENCE IF NOT EXISTS customer.customer_number_seq START WITH 1 INCREMENT BY 1;

CREATE TABLE IF NOT EXISTS customer.customers (
    id UUID PRIMARY KEY,
    customer_no VARCHAR(32) NOT NULL UNIQUE,
    customer_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    first_name VARCHAR(150),
    last_name VARCHAR(150),
    company_name VARCHAR(255),
    registration_number VARCHAR(100),
    note TEXT,
    created_by UUID,
    updated_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT customers_type_check CHECK (customer_type IN ('INDIVIDUAL', 'CORPORATE')),
    CONSTRAINT customers_status_check CHECK (status IN ('ACTIVE', 'INACTIVE', 'BLOCKED')),
    CONSTRAINT customers_name_check CHECK (
        (customer_type = 'INDIVIDUAL' AND first_name IS NOT NULL AND last_name IS NOT NULL)
        OR (customer_type = 'CORPORATE' AND company_name IS NOT NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS customers_registration_number_unique
    ON customer.customers(registration_number)
    WHERE registration_number IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS customers_search_idx
    ON customer.customers(
        customer_no,
        lower(COALESCE(first_name, '')),
        lower(COALESCE(last_name, '')),
        lower(COALESCE(company_name, ''))
    );

CREATE TABLE IF NOT EXISTS customer.customer_addresses (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES customer.customers(id),
    address_type VARCHAR(20) NOT NULL,
    line1 VARCHAR(255) NOT NULL,
    line2 VARCHAR(255),
    subdistrict VARCHAR(150),
    district VARCHAR(150),
    province VARCHAR(150) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    country_code CHAR(2) NOT NULL DEFAULT 'TH',
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT customer_addresses_type_check CHECK (address_type IN ('BILLING', 'SHIPPING', 'CONTACT'))
);

CREATE UNIQUE INDEX IF NOT EXISTS customer_addresses_one_default
    ON customer.customer_addresses(customer_id, address_type)
    WHERE is_default = TRUE AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS customer_addresses_customer_idx
    ON customer.customer_addresses(customer_id)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS customer.customer_contacts (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES customer.customers(id),
    first_name VARCHAR(150) NOT NULL,
    last_name VARCHAR(150) NOT NULL,
    position VARCHAR(150),
    email VARCHAR(255),
    phone VARCHAR(50),
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT customer_contacts_channel_check CHECK (email IS NOT NULL OR phone IS NOT NULL)
);

CREATE UNIQUE INDEX IF NOT EXISTS customer_contacts_one_primary
    ON customer.customer_contacts(customer_id)
    WHERE is_primary = TRUE AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS customer_contacts_customer_idx
    ON customer.customer_contacts(customer_id)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS customer.customer_tax_profiles (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES customer.customers(id),
    tax_id VARCHAR(50) NOT NULL,
    branch_type VARCHAR(20) NOT NULL,
    branch_code VARCHAR(20) NOT NULL,
    branch_name VARCHAR(255),
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    subdistrict VARCHAR(150),
    district VARCHAR(150),
    province VARCHAR(150) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    country_code CHAR(2) NOT NULL DEFAULT 'TH',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT customer_tax_profiles_branch_type_check CHECK (branch_type IN ('HEAD_OFFICE', 'BRANCH')),
    CONSTRAINT customer_tax_profiles_branch_check CHECK (
        (branch_type = 'HEAD_OFFICE' AND branch_code = '00000')
        OR (branch_type = 'BRANCH' AND branch_code <> '00000')
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS customer_tax_profiles_tax_branch_unique
    ON customer.customer_tax_profiles(tax_id, branch_code)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS customer_tax_profiles_customer_idx
    ON customer.customer_tax_profiles(customer_id)
    WHERE deleted_at IS NULL;
