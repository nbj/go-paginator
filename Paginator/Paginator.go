package Paginator

import (
	"encoding/json"
	"fmt"
	"github.com/nbj/go-collections/Collection"
	"gorm.io/gorm"
	"math"
	"reflect"
)

var DefaultBoundaries *Boundaries
var DefaultConnection *gorm.DB

type Paginator[T any] struct {
	connection *gorm.DB

	Page            int                       `json:"page"`
	PerPage         int                       `json:"per_page"`
	NextPage        int                       `json:"next_page"`
	PreviousPage    int                       `json:"previous_page"`
	LastPage        int                       `json:"last_page"`
	Total           int                       `json:"total"`
	FirstPageUrl    string                    `json:"first_page_url,omitempty"`
	LastPageUrl     string                    `json:"last_page_url,omitempty"`
	NextPageUrl     string                    `json:"next_page_url,omitempty"`
	PreviousPageUrl string                    `json:"previous_page_url,omitempty"`
	From            int                       `json:"from"`
	To              int                       `json:"to"`
	Path            string                    `json:"path"`
	Items           *Collection.Collection[T] `json:"items"`
}

type Boundaries struct {
	Page    int
	PerPage int
	Path    string
}

func SetDefaultBoundaries(boundaries *Boundaries) {
	DefaultBoundaries = boundaries
}

func SetDefaultConnection(connection *gorm.DB) {
	DefaultConnection = connection
}

// Paginate
// Gets a paginated database result of a specific type
// The function takes three types of arguments:
//   - connection  *gorm.DB
//   - boundaries  *Boundaries
//   - queries     []func(connection *gorm.DB) *gorm.DB)
func Paginate[T any](arguments ...any) *Paginator[T] {
	// Local variables to keep track of state
	var paginator Paginator[T]

	var boundaries *Boundaries
	var connection *gorm.DB
	var queries []func(connection *gorm.DB) *gorm.DB

	var output []T

	// If no arguments have been passed to the paginate function
	// and no default connection has been set, we simply return
	// nil, as we cannot give a result
	if len(arguments) == 0 && DefaultConnection == nil {
		return nil
	}

	// As the arguments passed to this function must comply with
	// a very specific scheme. The first argument being an
	// options struct or a query function, and the rest
	// being query functions. Query functions must be:
	// func(collection *gorm.DB) *gorm.DB
	if len(arguments) > 0 {
		// firstArgumentType := reflect.TypeOf(arguments[0]).String()
		for _, argument := range arguments {
			argumentType := reflect.TypeOf(argument).String()

			switch argumentType {
			case "*gorm.DB":
				connection = argument.(*gorm.DB)
			case "func(*gorm.DB) *gorm.DB":
				queries = append(queries, argument.(func(connection *gorm.DB) *gorm.DB))
			case "*Paginator.Boundaries":
				boundaries = argument.(*Boundaries)
			}
		}
	}

	// If no connection was passed to the function, and no default connection exists
	// we cannot continue and must return nil
	if connection == nil && DefaultConnection == nil {
		return nil
	}

	// If no connection was passed to the function, use the default connection
	if connection == nil {
		connection = DefaultConnection
	}

	// If no boundaries were passed to the function, use the defaults
	if boundaries == nil && DefaultBoundaries == nil {
		boundaries = &Boundaries{
			Page:    1,
			PerPage: 25,
			Path:    "",
		}
	}

	if boundaries == nil {
		boundaries = DefaultBoundaries
	}

	// If applicable, override empty boundary fields with values from the default boundaries
	if DefaultBoundaries != nil {
		if boundaries.Page == 0 {
			boundaries.Page = DefaultBoundaries.Page
		}

		if boundaries.PerPage == 0 {
			boundaries.PerPage = DefaultBoundaries.PerPage
		}

		if boundaries.Path == "" {
			boundaries.Path = DefaultBoundaries.Path
		}
	}

	// Instantiate a new paginator
	paginator = Paginator[T]{}

	// Configure the paginator
	paginator.connection = connection
	paginator.Page = boundaries.Page
	paginator.PerPage = boundaries.PerPage
	paginator.Path = boundaries.Path

	// Set some default values if none were present in the options
	paginator.assignDefaultPageAndPerPageValues()

	// Apply all the passed in query functions
	for _, query := range queries {
		paginator.connection = query(paginator.connection)
	}

	// Assign the total number of items
	paginator.assignTotal()

	// And the total number of pages
	paginator.assignLastPageIndex()

	// Bail if paginator is out of bounds
	if paginator.Page > paginator.LastPage {
		return nil
	}

	// Set next and previous page indices
	paginator.assignNextAndPreviousPageIndices()

	// Assign from and to values
	paginator.assignFromAndToValues()

	// Assign all paginator page urls
	paginator.assignPageUrls()

	// Lastly we want to add the actual data to the paginator
	paginator.connection.Offset(paginator.From - 1).Limit(paginator.PerPage).Find(&output)
	paginator.Items = Collection.Collect(output)

	return &paginator
}

// assignDefaultPageAndPerPageValues
// Assigns the page and per page values for the paginator
func (paginator *Paginator[T]) assignDefaultPageAndPerPageValues() *Paginator[T] {
	if paginator.Page == 0 {
		paginator.Page = 1
	}

	if paginator.PerPage == 0 {
		paginator.PerPage = 25
	}

	return paginator
}

// assignTotal
// Assigns the total number of items
func (paginator *Paginator[T]) assignTotal() *Paginator[T] {
	var total int64
	var output *[]T

	paginator.connection.Model(output).Count(&total)
	paginator.Total = int(total)

	return paginator
}

// assignNextAndPreviousPageIndices
// Assigns the next and previous page indices for the paginator
func (paginator *Paginator[T]) assignNextAndPreviousPageIndices() *Paginator[T] {
	paginator.NextPage = paginator.Page + 1
	paginator.PreviousPage = paginator.Page - 1

	if paginator.NextPage > paginator.LastPage {
		paginator.NextPage = 1
	}

	if paginator.PreviousPage < 1 {
		paginator.PreviousPage = paginator.LastPage
	}

	return paginator
}

// assignLastPageIndex
// Assigns the last page index for the paginator
func (paginator *Paginator[T]) assignLastPageIndex() *Paginator[T] {
	paginator.LastPage = int(math.Ceil(float64(paginator.Total) / float64(paginator.PerPage)))

	return paginator
}

// assignFromAndToValues
// Assigns the from and to values for the paginator
func (paginator *Paginator[T]) assignFromAndToValues() *Paginator[T] {
	paginator.From = 1 + ((paginator.Page - 1) * paginator.PerPage)
	paginator.To = paginator.Page * paginator.PerPage

	if paginator.Page == paginator.LastPage {
		paginator.To = paginator.Total
	}

	return paginator
}

// assignPageUrls
// Assigns all the appropriate paginator urls
func (paginator *Paginator[T]) assignPageUrls() *Paginator[T] {
	if paginator.Page != 1 {
		paginator.FirstPageUrl = fmt.Sprintf("%s?page=%d&per_page=%d", paginator.Path, 1, paginator.PerPage)
	}

	if paginator.Page != paginator.LastPage {
		paginator.LastPageUrl = fmt.Sprintf("%s?page=%d&per_page=%d", paginator.Path, paginator.LastPage, paginator.PerPage)
	}

	if paginator.Page+1 <= paginator.LastPage {
		paginator.NextPageUrl = fmt.Sprintf("%s?page=%d&per_page=%d", paginator.Path, paginator.Page+1, paginator.PerPage)
	}

	if paginator.Page-1 > 0 {
		paginator.PreviousPageUrl = fmt.Sprintf("%s?page=%d&per_page=%d", paginator.Path, paginator.Page-1, paginator.PerPage)
	}

	return paginator
}

func (paginator *Paginator[T]) MarshalJSON() ([]byte, error) {
	type Alias Paginator[T]
	return json.Marshal(&struct {
		*Alias
		Items []T `json:"items"`
	}{
		Alias: (*Alias)(paginator),
		Items: paginator.Items.Items,
	})
}
