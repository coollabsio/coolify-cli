package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/coollabsio/coolify-cli/internal/api"
)

// describeValidationErrors adds the field errors of a 422 response to the error, so the user
// sees why the request failed instead of only "Validation failed.".
func describeValidationErrors(err error) error {
	var apiErr *api.Error
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 422 {
		return err
	}
	var body struct {
		Errors map[string][]string `json:"errors"`
	}
	if json.Unmarshal(apiErr.Body, &body) != nil || len(body.Errors) == 0 {
		return err
	}
	fields := make([]string, 0, len(body.Errors))
	for field := range body.Errors {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	details := make([]string, 0, len(fields))
	for _, field := range fields {
		details = append(details, fmt.Sprintf("%s: %s", field, strings.Join(body.Errors[field], " ")))
	}
	return fmt.Errorf("%w (%s)", err, strings.Join(details, "; "))
}
