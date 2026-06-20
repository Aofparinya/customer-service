package httpadapter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/saaof/order-platform/customer-service/internal/application"
	"github.com/saaof/order-platform/customer-service/internal/domain"
)

type Handler struct {
	service  *application.Service
	validate *validator.Validate
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{
		service:  service,
		validate: validator.New(),
	}
}

func (handler *Handler) Health(context echo.Context) error {
	return context.JSON(http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "customer-service",
	})
}

func (handler *Handler) CreateCustomer(context echo.Context) error {
	var input application.CreateCustomerInput
	if err := handler.decodeAndValidate(context, &input); err != nil {
		return err
	}
	actorID, err := ActorID(context)
	if err != nil {
		return err
	}
	customer, err := handler.service.CreateCustomer(context.Request().Context(), input, actorID)
	if err != nil {
		return err
	}
	return context.JSON(http.StatusCreated, customer)
}

func (handler *Handler) ListCustomers(context echo.Context) error {
	filter, err := listFilter(context)
	if err != nil {
		return err
	}
	result, err := handler.service.ListCustomers(context.Request().Context(), filter)
	if err != nil {
		return err
	}
	return context.JSON(http.StatusOK, result)
}

func (handler *Handler) GetCustomer(context echo.Context) error {
	customerID, err := pathUUID(context, "id")
	if err != nil {
		return err
	}
	customer, err := handler.service.GetCustomer(context.Request().Context(), customerID)
	if err != nil {
		return err
	}
	return context.JSON(http.StatusOK, customer)
}

func (handler *Handler) UpdateCustomer(context echo.Context) error {
	customerID, err := pathUUID(context, "id")
	if err != nil {
		return err
	}
	var input application.UpdateCustomerInput
	if err := handler.decodeAndValidate(context, &input); err != nil {
		return err
	}
	actorID, err := ActorID(context)
	if err != nil {
		return err
	}
	customer, err := handler.service.UpdateCustomer(
		context.Request().Context(),
		customerID,
		input,
		actorID,
	)
	if err != nil {
		return err
	}
	return context.JSON(http.StatusOK, customer)
}

func (handler *Handler) DeleteCustomer(context echo.Context) error {
	customerID, err := pathUUID(context, "id")
	if err != nil {
		return err
	}
	actorID, err := ActorID(context)
	if err != nil {
		return err
	}
	if err := handler.service.DeleteCustomer(context.Request().Context(), customerID, actorID); err != nil {
		return err
	}
	return context.NoContent(http.StatusNoContent)
}

func (handler *Handler) CreateAddress(context echo.Context) error {
	customerID, err := pathUUID(context, "id")
	if err != nil {
		return err
	}
	var input application.CreateAddressInput
	if err := handler.decodeAndValidate(context, &input); err != nil {
		return err
	}
	address, err := handler.service.CreateAddress(context.Request().Context(), customerID, input)
	if err != nil {
		return err
	}
	return context.JSON(http.StatusCreated, address)
}

func (handler *Handler) ListAddresses(context echo.Context) error {
	customerID, err := pathUUID(context, "id")
	if err != nil {
		return err
	}
	addresses, err := handler.service.ListAddresses(context.Request().Context(), customerID)
	if err != nil {
		return err
	}
	return context.JSON(http.StatusOK, addresses)
}

func (handler *Handler) UpdateAddress(context echo.Context) error {
	customerID, childID, err := childIDs(context, "addressId")
	if err != nil {
		return err
	}
	var input application.UpdateAddressInput
	if err := handler.decodeAndValidate(context, &input); err != nil {
		return err
	}
	address, err := handler.service.UpdateAddress(
		context.Request().Context(),
		customerID,
		childID,
		input,
	)
	if err != nil {
		return err
	}
	return context.JSON(http.StatusOK, address)
}

func (handler *Handler) DeleteAddress(context echo.Context) error {
	customerID, childID, err := childIDs(context, "addressId")
	if err != nil {
		return err
	}
	if err := handler.service.DeleteAddress(context.Request().Context(), customerID, childID); err != nil {
		return err
	}
	return context.NoContent(http.StatusNoContent)
}

func (handler *Handler) CreateContact(context echo.Context) error {
	customerID, err := pathUUID(context, "id")
	if err != nil {
		return err
	}
	var input application.CreateContactInput
	if err := handler.decodeAndValidate(context, &input); err != nil {
		return err
	}
	contact, err := handler.service.CreateContact(context.Request().Context(), customerID, input)
	if err != nil {
		return err
	}
	return context.JSON(http.StatusCreated, contact)
}

func (handler *Handler) ListContacts(context echo.Context) error {
	customerID, err := pathUUID(context, "id")
	if err != nil {
		return err
	}
	contacts, err := handler.service.ListContacts(context.Request().Context(), customerID)
	if err != nil {
		return err
	}
	return context.JSON(http.StatusOK, contacts)
}

func (handler *Handler) UpdateContact(context echo.Context) error {
	customerID, childID, err := childIDs(context, "contactId")
	if err != nil {
		return err
	}
	var input application.UpdateContactInput
	if err := handler.decodeAndValidate(context, &input); err != nil {
		return err
	}
	contact, err := handler.service.UpdateContact(
		context.Request().Context(),
		customerID,
		childID,
		input,
	)
	if err != nil {
		return err
	}
	return context.JSON(http.StatusOK, contact)
}

func (handler *Handler) DeleteContact(context echo.Context) error {
	customerID, childID, err := childIDs(context, "contactId")
	if err != nil {
		return err
	}
	if err := handler.service.DeleteContact(context.Request().Context(), customerID, childID); err != nil {
		return err
	}
	return context.NoContent(http.StatusNoContent)
}

func (handler *Handler) CreateTaxProfile(context echo.Context) error {
	customerID, err := pathUUID(context, "id")
	if err != nil {
		return err
	}
	var input application.CreateTaxProfileInput
	if err := handler.decodeAndValidate(context, &input); err != nil {
		return err
	}
	profile, err := handler.service.CreateTaxProfile(context.Request().Context(), customerID, input)
	if err != nil {
		return err
	}
	return context.JSON(http.StatusCreated, profile)
}

func (handler *Handler) ListTaxProfiles(context echo.Context) error {
	customerID, err := pathUUID(context, "id")
	if err != nil {
		return err
	}
	profiles, err := handler.service.ListTaxProfiles(context.Request().Context(), customerID)
	if err != nil {
		return err
	}
	return context.JSON(http.StatusOK, profiles)
}

func (handler *Handler) UpdateTaxProfile(context echo.Context) error {
	customerID, childID, err := childIDs(context, "taxProfileId")
	if err != nil {
		return err
	}
	var input application.UpdateTaxProfileInput
	if err := handler.decodeAndValidate(context, &input); err != nil {
		return err
	}
	profile, err := handler.service.UpdateTaxProfile(
		context.Request().Context(),
		customerID,
		childID,
		input,
	)
	if err != nil {
		return err
	}
	return context.JSON(http.StatusOK, profile)
}

func (handler *Handler) DeleteTaxProfile(context echo.Context) error {
	customerID, childID, err := childIDs(context, "taxProfileId")
	if err != nil {
		return err
	}
	if err := handler.service.DeleteTaxProfile(context.Request().Context(), customerID, childID); err != nil {
		return err
	}
	return context.NoContent(http.StatusNoContent)
}

func (handler *Handler) decodeAndValidate(context echo.Context, target any) error {
	decoder := json.NewDecoder(context.Request().Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		if errors.Is(err, io.EOF) {
			return echo.NewHTTPError(http.StatusBadRequest, "Request body is required")
		}
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
	}
	if err := handler.validate.Struct(target); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, validationMessage(err))
	}
	return nil
}

func listFilter(context echo.Context) (application.ListFilter, error) {
	page, err := positiveQueryInt(context.QueryParam("page"), 1)
	if err != nil {
		return application.ListFilter{}, echo.NewHTTPError(http.StatusBadRequest, "page must be a positive integer")
	}
	pageSize, err := positiveQueryInt(context.QueryParam("pageSize"), 20)
	if err != nil {
		return application.ListFilter{}, echo.NewHTTPError(http.StatusBadRequest, "pageSize must be a positive integer")
	}
	filter := application.ListFilter{
		Query:    context.QueryParam("q"),
		Page:     page,
		PageSize: pageSize,
	}
	if value := strings.TrimSpace(context.QueryParam("customerType")); value != "" {
		customerType := domain.CustomerType(strings.ToUpper(value))
		if customerType != domain.CustomerTypeIndividual && customerType != domain.CustomerTypeCorporate {
			return application.ListFilter{}, echo.NewHTTPError(http.StatusBadRequest, "invalid customerType")
		}
		filter.CustomerType = &customerType
	}
	if value := strings.TrimSpace(context.QueryParam("status")); value != "" {
		status := domain.CustomerStatus(strings.ToUpper(value))
		if status != domain.CustomerStatusActive &&
			status != domain.CustomerStatusInactive &&
			status != domain.CustomerStatusBlocked {
			return application.ListFilter{}, echo.NewHTTPError(http.StatusBadRequest, "invalid status")
		}
		filter.Status = &status
	}
	return filter, nil
}

func positiveQueryInt(value string, fallback int) (int, error) {
	if strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	result, err := strconv.Atoi(value)
	if err != nil || result <= 0 {
		return 0, errors.New("invalid positive integer")
	}
	return result, nil
}

func pathUUID(context echo.Context, name string) (uuid.UUID, error) {
	value, err := uuid.Parse(context.Param(name))
	if err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, name+" must be a valid UUID")
	}
	return value, nil
}

func childIDs(context echo.Context, childName string) (uuid.UUID, uuid.UUID, error) {
	customerID, err := pathUUID(context, "id")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	childID, err := pathUUID(context, childName)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return customerID, childID, nil
}

func validationMessage(err error) string {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) || len(validationErrors) == 0 {
		return "Request validation failed"
	}
	item := validationErrors[0]
	return fmt.Sprintf("%s failed validation rule %s", item.Field(), item.Tag())
}
