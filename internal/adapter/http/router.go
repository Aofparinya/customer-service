package httpadapter

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/saaof/order-platform/customer-service/internal/adapter/auth"
	"github.com/saaof/order-platform/customer-service/internal/domain"
)

func NewRouter(handler *Handler, authClient *auth.Client, corsOrigins []string) *echo.Echo {
	server := echo.New()
	server.HideBanner = true
	server.HTTPErrorHandler = errorHandler
	server.Use(middleware.Recover())
	server.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogMethod:   true,
		LogLatency:  true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(_ echo.Context, values middleware.RequestLoggerValues) error {
			slog.Info("http request",
				"method", values.Method,
				"uri", values.URI,
				"status", values.Status,
				"latency", values.Latency,
				"error", values.Error,
			)
			return nil
		},
	}))
	server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: corsOrigins,
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	api := server.Group("/api/v1")
	api.GET("/health", handler.Health)

	customers := api.Group("/customers", Authenticate(authClient))
	read := RequirePermission("customers.read")
	write := RequirePermission("customers.write")

	customers.POST("", handler.CreateCustomer, write)
	customers.GET("", handler.ListCustomers, read)
	customers.GET("/:id", handler.GetCustomer, read)
	customers.PATCH("/:id", handler.UpdateCustomer, write)
	customers.DELETE("/:id", handler.DeleteCustomer, write)

	customers.POST("/:id/addresses", handler.CreateAddress, write)
	customers.GET("/:id/addresses", handler.ListAddresses, read)
	customers.PATCH("/:id/addresses/:addressId", handler.UpdateAddress, write)
	customers.DELETE("/:id/addresses/:addressId", handler.DeleteAddress, write)

	customers.POST("/:id/contacts", handler.CreateContact, write)
	customers.GET("/:id/contacts", handler.ListContacts, read)
	customers.PATCH("/:id/contacts/:contactId", handler.UpdateContact, write)
	customers.DELETE("/:id/contacts/:contactId", handler.DeleteContact, write)

	customers.POST("/:id/tax-profiles", handler.CreateTaxProfile, write)
	customers.GET("/:id/tax-profiles", handler.ListTaxProfiles, read)
	customers.PATCH("/:id/tax-profiles/:taxProfileId", handler.UpdateTaxProfile, write)
	customers.DELETE("/:id/tax-profiles/:taxProfileId", handler.DeleteTaxProfile, write)

	return server
}

func errorHandler(err error, context echo.Context) {
	if context.Response().Committed {
		return
	}
	status := http.StatusInternalServerError
	message := "Internal server error"
	errorName := "Internal Server Error"

	var httpError *echo.HTTPError
	switch {
	case errors.As(err, &httpError):
		status = httpError.Code
		message = messageFromHTTPError(httpError)
		errorName = http.StatusText(status)
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		message = "Resource not found"
		errorName = "Not Found"
	case errors.Is(err, domain.ErrConflict):
		status = http.StatusConflict
		message = err.Error()
		errorName = "Conflict"
	case errors.Is(err, domain.ErrInvalid):
		status = http.StatusBadRequest
		message = err.Error()
		errorName = "Bad Request"
	default:
		slog.Error("unhandled request error", "error", err)
	}

	_ = context.JSON(status, map[string]any{
		"message":    message,
		"error":      errorName,
		"statusCode": status,
	})
}

func messageFromHTTPError(err *echo.HTTPError) string {
	if message, ok := err.Message.(string); ok {
		return message
	}
	return http.StatusText(err.Code)
}
