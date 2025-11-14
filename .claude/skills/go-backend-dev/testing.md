# Testing in Go

Comprehensive guide to testing patterns, TDD workflow, and best practices.

## Test-Driven Development (TDD)

### The Red-Green-Refactor Cycle

1. **Red**: Write a failing test
2. **Green**: Write minimal code to pass
3. **Refactor**: Improve code while keeping tests green

### TDD Workflow Example

**Feature**: Validate email addresses

#### Step 1: Red (Failing Test)

```go
// validator_test.go
package validator_test

import (
    "testing"
    "myapp/validator"
)

func TestValidateEmail_ValidFormat_ReturnsTrue(t *testing.T) {
    email := "user@example.com"

    result := validator.ValidateEmail(email)

    if !result {
        t.Errorf("ValidateEmail(%q) = false, want true", email)
    }
}
```

Run: `go test` → **FAIL** (function doesn't exist)

#### Step 2: Green (Minimal Implementation)

```go
// validator.go
package validator

// ValidateEmail checks if an email address is valid.
// Returns true if the email contains '@' symbol, false otherwise.
//
// Parameters:
//   - email: Email address string to validate
//
// Returns:
//   - bool: True if email is valid format
func ValidateEmail(email string) bool {
    // Minimal implementation to pass test
    return strings.Contains(email, "@")
}
```

Run: `go test` → **PASS**

#### Step 3: Red (More Tests)

```go
func TestValidateEmail_MissingAtSymbol_ReturnsFalse(t *testing.T) {
    email := "userexample.com"

    result := validator.ValidateEmail(email)

    if result {
        t.Errorf("ValidateEmail(%q) = true, want false", email)
    }
}

func TestValidateEmail_EmptyString_ReturnsFalse(t *testing.T) {
    email := ""

    result := validator.ValidateEmail(email)

    if result {
        t.Errorf("ValidateEmail(%q) = true, want false", email)
    }
}
```

#### Step 4: Green (Fix Implementation)

```go
func ValidateEmail(email string) bool {
    if email == "" {
        return false
    }
    return strings.Contains(email, "@") && strings.Contains(email, ".")
}
```

#### Step 5: Refactor (Improve Tests with Table-Driven)

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name  string
        email string
        want  bool
    }{
        {"valid email", "user@example.com", true},
        {"missing @", "userexample.com", false},
        {"empty string", "", false},
        {"missing domain", "user@", false},
        {"missing local", "@example.com", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := validator.ValidateEmail(tt.email)
            if got != tt.want {
                t.Errorf("ValidateEmail(%q) = %v, want %v",
                    tt.email, got, tt.want)
            }
        })
    }
}
```

## Table-Driven Tests

### Basic Structure

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a    int
        b    int
        want int
    }{
        {"positive numbers", 2, 3, 5},
        {"negative numbers", -2, -3, -5},
        {"mixed signs", -2, 3, 1},
        {"zero", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("Add(%d, %d) = %d, want %d",
                    tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

### With Error Handling

```go
func TestParseInt(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    int
        wantErr error
    }{
        {
            name:    "valid positive number",
            input:   "123",
            want:    123,
            wantErr: nil,
        },
        {
            name:    "valid negative number",
            input:   "-456",
            want:    -456,
            wantErr: nil,
        },
        {
            name:    "invalid input",
            input:   "abc",
            want:    0,
            wantErr: ErrInvalidFormat,
        },
        {
            name:    "empty string",
            input:   "",
            want:    0,
            wantErr: ErrEmptyInput,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseInt(tt.input)

            // Check error
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("ParseInt(%q) error = %v, wantErr %v",
                    tt.input, err, tt.wantErr)
                return
            }

            // Check result (only if no error expected)
            if err == nil && got != tt.want {
                t.Errorf("ParseInt(%q) = %d, want %d",
                    tt.input, got, tt.want)
            }
        })
    }
}
```

### Complex Scenarios

```go
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name      string
        user      *User
        mockSetup func(*MockDB)
        want      *User
        wantErr   error
    }{
        {
            name: "successful creation",
            user: &User{Email: "test@example.com"},
            mockSetup: func(db *MockDB) {
                db.On("Insert", mock.Anything).Return(nil)
            },
            want:    &User{ID: 1, Email: "test@example.com"},
            wantErr: nil,
        },
        {
            name: "duplicate email",
            user: &User{Email: "existing@example.com"},
            mockSetup: func(db *MockDB) {
                db.On("Insert", mock.Anything).Return(ErrDuplicateEmail)
            },
            want:    nil,
            wantErr: ErrDuplicateEmail,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mock
            mockDB := new(MockDB)
            tt.mockSetup(mockDB)

            svc := NewUserService(mockDB)

            // Execute
            got, err := svc.CreateUser(tt.user)

            // Assert error
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            // Assert result
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("CreateUser() = %v, want %v", got, tt.want)
            }

            mockDB.AssertExpectations(t)
        })
    }
}
```

## Test Helpers

### 1. Assertion Helpers

```go
// assertEqual compares two values and fails the test if they differ.
// It's a helper function to reduce repetitive assertion code.
//
// Parameters:
//   - t: Testing instance
//   - got: Actual value received
//   - want: Expected value
//   - msg: Custom error message
func assertEqual(t *testing.T, got, want interface{}, msg string) {
    t.Helper() // Marks this as helper, shows caller's line in errors

    if !reflect.DeepEqual(got, want) {
        t.Errorf("%s: got %v, want %v", msg, got, want)
    }
}

// assertError checks if an error matches the expected error.
func assertError(t *testing.T, got, want error) {
    t.Helper()

    if !errors.Is(got, want) {
        t.Errorf("got error %v, want %v", got, want)
    }
}

// Usage
func TestSomething(t *testing.T) {
    result := DoSomething()
    assertEqual(t, result, 42, "DoSomething result")
}
```

### 2. Test Fixtures

```go
// setupTestDB creates a test database with sample data.
// It returns the database instance and a cleanup function.
// Always call cleanup in defer to ensure proper teardown.
//
// Returns:
//   - *sql.DB: Test database instance
//   - func(): Cleanup function to call with defer
func setupTestDB(t *testing.T) (*sql.DB, func()) {
    t.Helper()

    // Create test database
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("failed to create test db: %v", err)
    }

    // Run migrations
    if err := runMigrations(db); err != nil {
        t.Fatalf("failed to run migrations: %v", err)
    }

    // Insert test data
    seedTestData(db)

    // Return cleanup function
    cleanup := func() {
        db.Close()
    }

    return db, cleanup
}

// Usage
func TestUserRepository(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()

    repo := NewUserRepository(db)
    // Run tests...
}
```

### 3. Table Setup Helper

```go
// newTestUser creates a user for testing with optional overrides.
func newTestUser(overrides ...func(*User)) *User {
    // Default test user
    user := &User{
        Email:     "test@example.com",
        FirstName: "Test",
        LastName:  "User",
        Age:       25,
    }

    // Apply overrides
    for _, override := range overrides {
        override(user)
    }

    return user
}

// Usage in tests
func TestValidateUser(t *testing.T) {
    tests := []struct {
        name    string
        user    *User
        wantErr error
    }{
        {
            name:    "valid user",
            user:    newTestUser(),
            wantErr: nil,
        },
        {
            name: "invalid email",
            user: newTestUser(func(u *User) {
                u.Email = "invalid-email"
            }),
            wantErr: ErrInvalidEmail,
        },
        {
            name: "underage user",
            user: newTestUser(func(u *User) {
                u.Age = 15
            }),
            wantErr: ErrUnderage,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateUser(tt.user)
            assertError(t, err, tt.wantErr)
        })
    }
}
```

## Mocking

### 1. Interface-Based Mocking

```go
// UserRepository defines database operations for users.
type UserRepository interface {
    FindByID(id int) (*User, error)
    Create(user *User) error
    Update(user *User) error
}

// MockUserRepository implements UserRepository for testing.
// It records method calls and allows configuring return values.
type MockUserRepository struct {
    FindByIDFunc func(id int) (*User, error)
    CreateFunc   func(user *User) error
    UpdateFunc   func(user *User) error
}

func (m *MockUserRepository) FindByID(id int) (*User, error) {
    if m.FindByIDFunc != nil {
        return m.FindByIDFunc(id)
    }
    return nil, errors.New("not implemented")
}

func (m *MockUserRepository) Create(user *User) error {
    if m.CreateFunc != nil {
        return m.CreateFunc(user)
    }
    return errors.New("not implemented")
}

func (m *MockUserRepository) Update(user *User) error {
    if m.UpdateFunc != nil {
        return m.UpdateFunc(user)
    }
    return errors.New("not implemented")
}

// Usage in tests
func TestUserService_GetUser(t *testing.T) {
    // Setup mock
    mockRepo := &MockUserRepository{
        FindByIDFunc: func(id int) (*User, error) {
            if id == 1 {
                return &User{ID: 1, Email: "test@example.com"}, nil
            }
            return nil, ErrUserNotFound
        },
    }

    svc := NewUserService(mockRepo)

    // Test successful case
    user, err := svc.GetUser(1)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if user.ID != 1 {
        t.Errorf("got ID %d, want 1", user.ID)
    }

    // Test not found case
    _, err = svc.GetUser(999)
    if !errors.Is(err, ErrUserNotFound) {
        t.Errorf("got error %v, want ErrUserNotFound", err)
    }
}
```

### 2. Using testify/mock (Optional Library)

```go
import (
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/assert"
)

type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) FindByID(id int) (*User, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

// Usage
func TestWithTestify(t *testing.T) {
    mockRepo := new(MockRepository)

    // Configure mock behavior
    mockRepo.On("FindByID", 1).Return(&User{ID: 1}, nil)
    mockRepo.On("FindByID", 2).Return(nil, ErrUserNotFound)

    svc := NewUserService(mockRepo)

    // Test
    user, err := svc.GetUser(1)
    assert.NoError(t, err)
    assert.Equal(t, 1, user.ID)

    // Verify mock was called correctly
    mockRepo.AssertExpectations(t)
}
```

## Integration Tests

### 1. Database Integration Tests

```go
// +build integration

package repository_test

import (
    "database/sql"
    "testing"
    _ "github.com/lib/pq"
)

func TestUserRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Connect to test database
    db, err := sql.Open("postgres", "postgres://localhost/test?sslmode=disable")
    if err != nil {
        t.Fatalf("failed to connect to test db: %v", err)
    }
    defer db.Close()

    // Clean up before test
    cleanup(t, db)

    repo := NewUserRepository(db)

    t.Run("create and retrieve user", func(t *testing.T) {
        user := &User{
            Email: "integration@test.com",
            Name:  "Integration Test",
        }

        // Create
        err := repo.Create(user)
        if err != nil {
            t.Fatalf("failed to create user: %v", err)
        }

        // Retrieve
        retrieved, err := repo.FindByEmail("integration@test.com")
        if err != nil {
            t.Fatalf("failed to find user: %v", err)
        }

        // Verify
        if retrieved.Email != user.Email {
            t.Errorf("got email %s, want %s", retrieved.Email, user.Email)
        }
    })
}

func cleanup(t *testing.T, db *sql.DB) {
    t.Helper()
    _, err := db.Exec("DELETE FROM users WHERE email LIKE '%@test.com'")
    if err != nil {
        t.Fatalf("cleanup failed: %v", err)
    }
}
```

### 2. HTTP Handler Tests

```go
func TestHandleGetUser(t *testing.T) {
    // Setup mock repository
    mockRepo := &MockUserRepository{
        FindByIDFunc: func(id int) (*User, error) {
            if id == 1 {
                return &User{ID: 1, Email: "test@example.com"}, nil
            }
            return nil, ErrUserNotFound
        },
    }

    // Create handler
    handler := HandleGetUser(mockRepo)

    tests := []struct {
        name           string
        userID         string
        wantStatus     int
        wantBodyContains string
    }{
        {
            name:             "existing user",
            userID:           "1",
            wantStatus:       http.StatusOK,
            wantBodyContains: "test@example.com",
        },
        {
            name:       "user not found",
            userID:     "999",
            wantStatus: http.StatusNotFound,
        },
        {
            name:       "invalid user id",
            userID:     "abc",
            wantStatus: http.StatusBadRequest,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Create request
            req := httptest.NewRequest("GET", "/users/"+tt.userID, nil)
            req = req.WithContext(context.WithValue(req.Context(), "userID", tt.userID))

            // Create response recorder
            rr := httptest.NewRecorder()

            // Execute handler
            handler.ServeHTTP(rr, req)

            // Check status code
            if rr.Code != tt.wantStatus {
                t.Errorf("got status %d, want %d", rr.Code, tt.wantStatus)
            }

            // Check response body
            if tt.wantBodyContains != "" {
                if !strings.Contains(rr.Body.String(), tt.wantBodyContains) {
                    t.Errorf("response body doesn't contain %q", tt.wantBodyContains)
                }
            }
        })
    }
}
```

## Benchmark Tests

```go
func BenchmarkValidateEmail(b *testing.B) {
    email := "test@example.com"

    // Run validation b.N times
    for i := 0; i < b.N; i++ {
        ValidateEmail(email)
    }
}

func BenchmarkProcessData(b *testing.B) {
    data := generateTestData(1000)

    // Reset timer to exclude setup time
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        ProcessData(data)
    }
}

// Run benchmarks
// go test -bench=. -benchmem
```

## Test Coverage

```bash
# Run tests with coverage
go test -cover ./...

# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# Check coverage threshold (example: require 80%)
go test -cover ./... | grep -E "coverage: [0-9]+\.[0-9]+%" | \
    awk '{if (substr($2, 1, length($2)-1) < 80) exit 1}'
```

## Test Organization

### File Structure

```
package/
├── user.go           # Implementation
├── user_test.go      # Unit tests
└── integration_test.go  # Integration tests (with build tag)
```

### Naming Conventions

**Test Functions:**
```go
func Test<Function>_<Scenario>_<ExpectedBehavior>(t *testing.T)

// Examples:
func TestCreateUser_ValidInput_ReturnsUser(t *testing.T)
func TestCreateUser_DuplicateEmail_ReturnsError(t *testing.T)
func TestCreateUser_EmptyEmail_ReturnsValidationError(t *testing.T)
```

**Benchmark Functions:**
```go
func Benchmark<Operation>(b *testing.B)

// Examples:
func BenchmarkHashPassword(b *testing.B)
func BenchmarkValidateEmail(b *testing.B)
```

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -run TestCreateUser

# Run specific test pattern
go test -run "TestCreate.*"

# Run tests in specific package
go test ./internal/user/...

# Skip integration tests
go test -short ./...

# Run with race detector
go test -race ./...

# Run with coverage
go test -cover ./...

# Run benchmarks
go test -bench=.

# Run benchmarks with memory stats
go test -bench=. -benchmem
```

## Best Practices

1. **Write tests first** (TDD red-green-refactor)
2. **Use table-driven tests** for multiple scenarios
3. **Test behavior, not implementation**
4. **Keep tests simple and readable**
5. **Use t.Helper()** in helper functions
6. **Mock external dependencies** (database, APIs)
7. **Test error cases** thoroughly
8. **Use meaningful test names** (TestX_Scenario_Expected)
9. **Avoid test interdependence** (each test should be isolated)
10. **Clean up resources** (use defer for cleanup)

## Common Pitfalls

❌ **Don't test private functions directly** (test through public API)
❌ **Don't skip error checking in tests**
❌ **Don't use time.Sleep in tests** (use channels/mocks for synchronization)
❌ **Don't share state between tests**
❌ **Don't make tests depend on execution order**
❌ **Don't test third-party libraries** (test your usage of them)

## Summary Checklist

- [ ] Write test before implementation (TDD)
- [ ] Use table-driven tests for multiple cases
- [ ] Test happy path and error cases
- [ ] Mock external dependencies
- [ ] Add integration tests for critical paths
- [ ] Document test purpose in comments
- [ ] Run tests before committing
- [ ] Check test coverage regularly
- [ ] Use benchmarks for performance-critical code
- [ ] Keep tests fast and isolated
