package Feature

import (
	"github.com/nbj/go-paginator/Paginator"
	"github.com/nbj/go-paginator/Tests"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
)

func Test_it_returns_nil_if_no_options_are_available(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()

	// Act
	paginator := Paginator.Paginate[Tests.TestCaseModel]()

	// Assert
	assert.Nil(t, paginator)
	Tests.CleanUp(t)
}

func Test_it_can_have_default_options_set(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()
	assert.Nil(t, Paginator.DefaultOptions)

	options := Paginator.Options{
		Connection: Tests.DB,
	}

	// Act
	Paginator.SetDefaultOptions(&options)

	// Assert
	assert.NotNil(t, Paginator.DefaultOptions)
	Tests.CleanUp(t)
}

func Test_it_can_have_default_options_set_using_testcase(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()
	assert.Nil(t, Paginator.DefaultOptions)

	// Act
	Tests.SetupPaginator()

	// Assert
	assert.NotNil(t, Paginator.DefaultOptions)
	Tests.CleanUp(t)
}

func Test_it_returns_a_paginator_instance_if_valid_options_are_available(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()
	Tests.SetupPaginator()

	assert.Equal(t, 1, Paginator.DefaultOptions.Page)
	assert.Equal(t, 25, Paginator.DefaultOptions.PerPage)
	assert.Equal(t, "tests/default", Paginator.DefaultOptions.Path)

	// Act
	paginator := Paginator.Paginate[Tests.TestCaseModel]()

	// Assert
	assert.NotNil(t, paginator)
	assert.Equal(t, 1, paginator.Page)
	assert.Equal(t, 25, paginator.PerPage)
	assert.Equal(t, "tests/default", paginator.Path)
	Tests.CleanUp(t)
}

func Test_it_returns_nil_if_invalid_first_argument(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()
	Tests.SetupPaginator()
	// options := Paginator.Options{Connection: Tests.DB}

	// Act
	paginator := Paginator.Paginate[Tests.TestCaseModel]("this-is-invalid")

	// Assert
	assert.Nil(t, paginator)
	Tests.CleanUp(t)
}

func Test_options_can_be_passed_as_the_first_argument(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()

	assert.Nil(t, Paginator.DefaultOptions)

	options := Paginator.Options{
		Connection: Tests.DB,
		Page:       1,
		PerPage:    10,
		Path:       "tests/custom",
	}

	// Act
	paginator := Paginator.Paginate[Tests.TestCaseModel](options)

	// Assert
	assert.NotNil(t, paginator)
	assert.Equal(t, 1, paginator.Page)
	assert.Equal(t, 10, paginator.PerPage)
	assert.Equal(t, "tests/custom", paginator.Path)
	Tests.CleanUp(t)
}

func Test_default_options_can_be_overridden_by_the_first_argument(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()
	Tests.SetupPaginator()

	assert.NotNil(t, Paginator.DefaultOptions)
	assert.Equal(t, 1, Paginator.DefaultOptions.Page)
	assert.Equal(t, 25, Paginator.DefaultOptions.PerPage)
	assert.Equal(t, "tests/default", Paginator.DefaultOptions.Path)

	options := Paginator.Options{
		Connection: Tests.DB,
		Page:       1,
		PerPage:    10,
		Path:       "tests/custom",
	}

	// Act
	paginator := Paginator.Paginate[Tests.TestCaseModel](options)

	// Assert
	assert.NotNil(t, paginator)
	assert.Equal(t, 1, paginator.Page)
	assert.Equal(t, 10, paginator.PerPage)
	assert.Equal(t, "tests/custom", paginator.Path)
	Tests.CleanUp(t)
}

func Test_it_returns_nil_if_no_default_options_are_available_and_first_argument_is_not_options(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()

	// Act
	paginator := Paginator.Paginate[Tests.TestCaseModel](func(connection *gorm.DB) *gorm.DB {
		return connection
	})

	// Assert
	assert.Nil(t, paginator)
}

func Test_it_returns_paginated_result(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()
	Tests.SetupPaginator()

	// Act
	paginator := Paginator.Paginate[Tests.TestCaseModel]()

	// Assert
	assert.Equal(t, 5, paginator.Total)
	assert.NotEmpty(t, paginator.Items)
	assert.Equal(t, "Value [1]", paginator.Items[0].Value)
	assert.Equal(t, "Value [2]", paginator.Items[1].Value)
	assert.Equal(t, "Value [3]", paginator.Items[2].Value)
	assert.Equal(t, "Value [4]", paginator.Items[3].Value)
	assert.Equal(t, "Value [5]", paginator.Items[4].Value)
}

func Test_it_returns_paginated_result_based_on_query_functions_passed(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()
	Tests.SetupPaginator()

	// Act
	paginator := Paginator.Paginate[Tests.TestCaseModel](func(connection *gorm.DB) *gorm.DB {
		return connection.
			Where("value = ?", "Value [2]")
	})

	// Assert
	assert.Equal(t, 1, paginator.Total)
	assert.NotEmpty(t, paginator.Items)
	assert.Equal(t, "Value [2]", paginator.Items[0].Value)
}

func Test_it_returns_paginated_result_based_on_multiple_query_functions_passed(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()
	Tests.SetupPaginator()

	// Act
	paginator := Paginator.Paginate[Tests.TestCaseModel](
		func(connection *gorm.DB) *gorm.DB {
			return connection.Where("value in ?", []string{"Value [2]", "Value [4]"})
		},
		func(connection *gorm.DB) *gorm.DB {
			return connection.Order("value desc")
		},
	)

	// Assert
	assert.Equal(t, 2, paginator.Total)
	assert.NotEmpty(t, paginator.Items)
	assert.Equal(t, "Value [4]", paginator.Items[0].Value)
	assert.Equal(t, "Value [2]", paginator.Items[1].Value)
}
