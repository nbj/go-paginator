package Tests

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/nbj/go-paginator/Paginator"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
	"time"
)

var DB *gorm.DB

func SetupEnvironment() {
	// Instantiate database with test data
	connection := getSqliteDatabaseConnection()
	seedTestData(connection)

	DB = connection
}

func SetupPaginator() {
	// Configure paginator
	boundaries := Paginator.Boundaries{
		Page:    1,
		PerPage: 25,
		Path:    "tests/default",
	}

	Paginator.SetDefaultBoundaries(&boundaries)
	Paginator.SetDefaultConnection(DB)
}

func CleanUp(t *testing.T) {
	t.Cleanup(func() {
		Paginator.SetDefaultBoundaries(nil)
		Paginator.SetDefaultConnection(nil)
	})
}

func getSqliteDatabaseConnection() *gorm.DB {
	var connection *gorm.DB
	var err error

	if connection, err = gorm.Open(sqlite.Open("file::memory:?cache=private"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}); nil != err {
		panic("failed to connect database: " + err.Error())
	}

	modelsToMigrate := []any{
		TestCaseModel{},
	}

	if err = connection.AutoMigrate(modelsToMigrate...); nil != err {
		panic("failed to auto migrate database: " + err.Error())
	}

	return connection
}

func seedTestData(connection *gorm.DB) {
	numberOfEntries := 5

	for entry := 1; entry <= numberOfEntries; entry++ {
		uniqueIdentifier, _ := uuid.NewV7()

		instance := TestCaseModel{
			Id:    uniqueIdentifier,
			Value: fmt.Sprintf("Value [%d]", entry),
		}

		connection.Create(&instance)
	}
}

type TestCaseModel struct {
	Id        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;uniqueIndex"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at" gorm:"index;not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
}
