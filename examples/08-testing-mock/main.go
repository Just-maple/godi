package main

import (
	"fmt"

	"github.com/Just-maple/godi"
)

// Testing and Mock Example: Using dependency injection for testable code
// Demonstrates how to use interfaces and mocks for testing

type Database interface {
	Query(sql string) ([]map[string]interface{}, error)
}

type RealDatabase struct {
	DSN string
}

func (d *RealDatabase) Query(sql string) ([]map[string]interface{}, error) {
	// Actual database query
	return nil, nil
}

type MockDatabase struct {
	Data []map[string]interface{}
}

func (d *MockDatabase) Query(sql string) ([]map[string]interface{}, error) {
	return d.Data, nil
}

type UserService struct {
	DB Database
}

func (s *UserService) GetUser(id int) (map[string]interface{}, error) {
	results, err := s.DB.Query(fmt.Sprintf("SELECT * FROM users WHERE id = %d", id))
	if err != nil {
		return nil, err
	}
	if len(results) > 0 {
		return results[0], nil
	}
	return nil, nil
}

func main() {
	// Production configuration
	prodContainer := &godi.Container{}
	prodContainer.MustAdd(
		godi.Provide(&RealDatabase{DSN: "mysql://localhost/prod"}),
		godi.Provide(&UserService{DB: &RealDatabase{DSN: "mysql://localhost/prod"}}),
	)

	// Test configuration (using Mock)
	testContainer := &godi.Container{}
	mockDB := &MockDatabase{
		Data: []map[string]interface{}{
			{"id": 1, "name": "Test User", "email": "test@example.com"},
		},
	}
	testContainer.MustAdd(
		godi.Provide(mockDB),
		godi.Provide(&UserService{DB: mockDB}),
	)

	// Use test container
	testUserSvc, _ := godi.Inject[*UserService](testContainer)
	user, _ := testUserSvc.GetUser(1)
	fmt.Printf("Test user: %v\n", user)

	// Use production container
	prodUserSvc, _ := godi.Inject[*UserService](prodContainer)
	fmt.Printf("Production service ready: %v\n", prodUserSvc != nil)

	fmt.Println("\nDemo complete!")
	fmt.Println("- Test environment uses MockDatabase")
	fmt.Println("- Production environment uses RealDatabase")
}
