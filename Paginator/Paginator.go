package Paginator

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"math"
	"reflect"
)

var DefaultOptions *Options

type Paginator[T any] struct {
	Page            int    `json:"page"`
	PerPage         int    `json:"per_page"`
	NextPage        int    `json:"next_page"`
	PreviousPage    int    `json:"previous_page"`
	LastPage        int    `json:"last_page"`
	Total           int    `json:"total"`
	FirstPageUrl    string `json:"first_page_url,omitempty"`
	LastPageUrl     string `json:"last_page_url,omitempty"`
	NextPageUrl     string `json:"next_page_url,omitempty"`
	PreviousPageUrl string `json:"previous_page_url,omitempty"`
	From            int    `json:"from"`
	To              int    `json:"to"`
	Path            string `json:"path"`
	Items           []T    `json:"items"`
}

type Options struct {
	Connection *gorm.DB
	Page       int
	PerPage    int
	Path       string
}

func SetDefaultOptions(options *Options) {
	DefaultOptions = options
}

func Paginate[T any](arguments ...any) *Paginator[T] {
	// Local variables to keep track of state
	var paginator *Paginator[T]
	var options *Options
	var queries []func(connection *gorm.DB) *gorm.DB
	var total int64
	var output *[]T

	// Defer function to recover from type conversion errors
	// and make sure return nil as expected on error
	defer func() {
		if err := recover(); err != nil {
			paginator = nil
			log.Println(err)
			return
		}
	}()

	// If no options have been passed to the paginate function
	// and no default options has been set, we simply return
	// nil, as we cannot give a result
	if len(arguments) == 0 && DefaultOptions == nil {
		return nil
	}

	// As the arguments passed to this function must comply with
	// a very specific scheme. The first argument being an
	// options struct or a query function, and the rest
	// being query functions. Query functions must be:
	// func(collection *gorm.DB) *gorm.DB
	if len(arguments) > 0 {
		firstArgumentType := reflect.TypeOf(arguments[0]).String()

		switch firstArgumentType {
		case "func(*gorm.DB) *gorm.DB":
			if DefaultOptions == nil {
				return nil
			}

			convertedArgument := arguments[0].(func(connection *gorm.DB) *gorm.DB)
			queries = append(queries, convertedArgument)
		default:
			optionsArgument := arguments[0].(Options)
			options = &optionsArgument
		}
	}

	// Now that we have the first argument in place, we need to unpack the query functions
	if len(arguments) > 1 {
		arguments = arguments[1:]

		for _, argument := range arguments {
			convertedArgument := argument.(func(*gorm.DB) *gorm.DB)
			queries = append(queries, convertedArgument)
		}
	}

	// If no options was passed to the function, use the default options
	if options == nil {
		options = DefaultOptions
	}

	// Instantiate a new paginator
	paginator = &Paginator[T]{}

	// Configure the paginator
	paginator.Page = options.Page
	paginator.PerPage = options.PerPage
	paginator.Path = options.Path

	// Set some default values if none were present in the options
	if paginator.Page == 0 {
		paginator.Page = 1
	}

	if paginator.PerPage == 0 {
		paginator.PerPage = 25
	}

	// Apply all the passed in query functions
	for _, query := range queries {
		options.Connection = query(options.Connection)
	}

	// Calculate the total number of entries
	options.Connection.Model(output).Count(&total)
	paginator.Total = int(total)

	// And the total number of pages
	paginator.LastPage = int(math.Ceil(float64(total) / float64(paginator.PerPage)))

	// Bail if paginator is out of bounds
	if paginator.Page > paginator.LastPage {
		return nil
	}

	// Set next and previous page indices
	paginator.NextPage = paginator.Page + 1
	paginator.PreviousPage = paginator.Page - 1

	if paginator.NextPage > paginator.LastPage {
		paginator.NextPage = 1
	}

	if paginator.PreviousPage < 1 {
		paginator.PreviousPage = paginator.LastPage
	}

	// Calculate from and to values
	paginator.From = 1 + ((paginator.Page - 1) * paginator.PerPage)
	paginator.To = paginator.Page * paginator.PerPage

	if paginator.Page == paginator.LastPage {
		paginator.To = paginator.Total
	}

	// Set all paginator links
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

	// Lastly we want to add the actual data to the paginator
	options.Connection.Offset(paginator.From - 1).Limit(paginator.PerPage).Find(&output)
	paginator.Items = *output

	return paginator
}
