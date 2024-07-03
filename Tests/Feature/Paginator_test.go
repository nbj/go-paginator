package Feature

import (
	"github.com/nbj/go-paginator/Paginator"
	"github.com/nbj/go-paginator/Tests"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
)

func Test_it_returns_nil_if_no_connection_is_available(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()
	assert.Nil(t, Paginator.DefaultConnection)

	// Act
	paginator := Paginator.Paginate[Tests.TestCaseModel]()

	// Assert
	assert.Nil(t, paginator)
	Tests.CleanUp(t)
}

func Test_it_can_have_default_connection_set(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()
	assert.Nil(t, Paginator.DefaultConnection)

	// Act
	Paginator.SetDefaultConnection(Tests.DB)

	// Assert
	assert.NotNil(t, Paginator.DefaultConnection)
	Tests.CleanUp(t)
}

func Test_it_can_have_default_boundaries_set(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()
	assert.Nil(t, Paginator.DefaultBoundaries)

	boundaries := Paginator.Boundaries{
		Page:    1,
		PerPage: 50,
		Path:    "tests/custom",
	}

	// Act
	Paginator.SetDefaultBoundaries(&boundaries)

	// Assert
	assert.NotNil(t, Paginator.DefaultBoundaries)
	assert.Equal(t, 1, Paginator.DefaultBoundaries.Page)
	assert.Equal(t, 50, Paginator.DefaultBoundaries.PerPage)
	assert.Equal(t, "tests/custom", Paginator.DefaultBoundaries.Path)
	Tests.CleanUp(t)
}

func Test_it_can_have_default_boundaries_set_using_testcase(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()
	assert.Nil(t, Paginator.DefaultBoundaries)

	// Act
	Tests.SetupPaginator()

	// Assert
	assert.NotNil(t, Paginator.DefaultBoundaries)
	assert.Equal(t, 1, Paginator.DefaultBoundaries.Page)
	assert.Equal(t, 25, Paginator.DefaultBoundaries.PerPage)
	assert.Equal(t, "tests/default", Paginator.DefaultBoundaries.Path)
	Tests.CleanUp(t)
}

func Test_it_can_have_default_connection_set_using_testcase(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()
	assert.Nil(t, Paginator.DefaultConnection)

	// Act
	Tests.SetupPaginator()

	// Assert
	assert.NotNil(t, Paginator.DefaultConnection)
	Tests.CleanUp(t)
}

func Test_it_returns_a_paginator_instance_if_valid_connection_is_available(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()
	Tests.SetupPaginator()

	// Act
	paginator := Paginator.Paginate[Tests.TestCaseModel]()

	// Assert
	assert.NotNil(t, paginator)
	Tests.CleanUp(t)
}

func Test_it_returns_a_paginator_instance_with_boundaries_set_eventhough_only_connection_was_set(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()
	Paginator.SetDefaultConnection(Tests.DB)

	// Act
	paginator := Paginator.Paginate[Tests.TestCaseModel]()

	// Assert
	assert.NotNil(t, paginator)
	assert.Equal(t, 1, paginator.Page)
	assert.Equal(t, 25, paginator.PerPage)
	assert.Equal(t, "", paginator.Path)
	Tests.CleanUp(t)
}

func Test_it_returns_nil_if_only_invalid_arguments(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()

	// Act
	paginator := Paginator.Paginate[Tests.TestCaseModel]("this-is-invalid", 666, false)

	// Assert
	assert.Nil(t, paginator)
	Tests.CleanUp(t)
}

func Test_it_returns_nil_if_no_default_connection_exists_even_though_arguments_are_valid_as_long_as_no_connection_is_passed(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()

	// Act
	paginator := Paginator.Paginate[Tests.TestCaseModel](&Paginator.Boundaries{
		Page:    2,
		PerPage: 45,
		Path:    "tests/custom",
	})

	// Assert
	assert.Nil(t, paginator)
	Tests.CleanUp(t)
}

func Test_it_returns_a_paginator_instance_if_connection_is_the_only_argument_passed(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()

	// Act
	paginator := Paginator.Paginate[Tests.TestCaseModel](Tests.DB)

	// Assert
	assert.NotNil(t, paginator)
	Tests.CleanUp(t)
}

func Test_boundaries_can_be_passed_as_arguments(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()
	Paginator.SetDefaultConnection(Tests.DB)
	assert.Nil(t, Paginator.DefaultBoundaries)

	boundaries := &Paginator.Boundaries{
		Page:    1,
		PerPage: 10,
		Path:    "tests/custom",
	}

	// Act
	paginator := Paginator.Paginate[Tests.TestCaseModel](boundaries)

	// Assert
	assert.Nil(t, Paginator.DefaultBoundaries)
	assert.NotNil(t, paginator)
	assert.Equal(t, 1, paginator.Page)
	assert.Equal(t, 10, paginator.PerPage)
	assert.Equal(t, "tests/custom", paginator.Path)
	Tests.CleanUp(t)
}

func Test_last_boundaries_passed_takes_presidency(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()
	Paginator.SetDefaultConnection(Tests.DB)
	assert.Nil(t, Paginator.DefaultBoundaries)

	boundariesA := &Paginator.Boundaries{
		Page:    1,
		PerPage: 10,
		Path:    "tests/customA",
	}

	boundariesB := &Paginator.Boundaries{
		Page:    1,
		PerPage: 20,
		Path:    "tests/customB",
	}

	// Act
	paginator := Paginator.Paginate[Tests.TestCaseModel](boundariesA, boundariesB)

	// Assert
	assert.Nil(t, Paginator.DefaultBoundaries)
	assert.NotNil(t, paginator)
	assert.Equal(t, 1, paginator.Page)
	assert.Equal(t, 20, paginator.PerPage)
	assert.Equal(t, "tests/customB", paginator.Path)
	Tests.CleanUp(t)
}

func Test_nil_is_returned_if_page_is_out_of_bounds(t *testing.T) {
	// Arrange
	Tests.SetupEnvironment()
	Paginator.SetDefaultConnection(Tests.DB)

	// Act
	paginator := Paginator.Paginate[Tests.TestCaseModel](&Paginator.Boundaries{
		Page:    100,
		PerPage: 10,
		Path:    "tests/custom",
	})

	// Assert
	assert.Nil(t, paginator)
	Tests.CleanUp(t)
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
