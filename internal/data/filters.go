package data

import (
	"fmt"
	"github.com/minhnghia2k3/greenlight/internal/validation"
	"math"
	"strings"
)

type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

func ValidateFilters(v *validation.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10000000, "page", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")
	v.Check(validation.PermittedValue(f.Sort, f.SortSafeList...), "sort", fmt.Sprintf("invalid sort value %q", f.SortSafeList))
}

// The calcualteMedata() function calculates the pagination metadata
// values given the total number of records, current page, and page size values.
func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}

// Check that the client-provided f.Sort field matches one of the entries in our safelist
// and if it does, extract the column name from the Sort field by stripping the leading
// hyphen character.
func (f *Filters) sortColumn() string {
	for _, safeValue := range f.SortSafeList {
		if f.Sort == safeValue {
			// Trim prefix
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	// Help prevent SQL injection.
	panic("unsafe sort parameter: " + f.Sort)
}

// Return the sort direction ("ASC" OR "DESC") depending on the prefix character of the
// sort field.
func (f *Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	// page_size=5&page=3
	// limit = 5, offset = 10.
	return (f.Page - 1) * f.PageSize
}
