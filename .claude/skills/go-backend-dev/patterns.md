# Idiomatic Go Patterns

This document covers standard Go patterns and conventions from Effective Go.

## Error Handling

### 1. Return Errors, Don't Panic

```go
// ❌ Bad: Using panic for expected errors
func GetUser(id int) *User {
    user, err := db.Query(id)
    if err != nil {
        panic(err) // Don't do this!
    }
    return user
}

// ✅ Good: Return errors explicitly
func GetUser(id int) (*User, error) {
    user, err := db.Query(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user %d: %w", id, err)
    }
    return user, nil
}
```

### 2. Wrap Errors with Context

```go
// ✅ Good: Wrap errors with fmt.Errorf and %w
func ProcessOrder(orderID int) error {
    order, err := fetchOrder(orderID)
    if err != nil {
        return fmt.Errorf("failed to fetch order: %w", err)
    }

    if err := validateOrder(order); err != nil {
        return fmt.Errorf("order validation failed: %w", err)
    }

    return nil
}

// Check wrapped errors
err := ProcessOrder(123)
if errors.Is(err, ErrOrderNotFound) {
    // Handle specific error
}
```

### 3. Sentinel Errors

```go
// Define package-level sentinel errors for common cases
var (
    ErrNotFound         = errors.New("resource not found")
    ErrInvalidInput     = errors.New("invalid input")
    ErrUnauthorized     = errors.New("unauthorized access")
    ErrDuplicateEntry   = errors.New("duplicate entry")
)

// Use in functions
func FindUser(id int) (*User, error) {
    user, err := db.Query(id)
    if err == sql.ErrNoRows {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("database error: %w", err)
    }
    return user, nil
}
```

### 4. Custom Error Types

```go
// ValidationError provides detailed validation failure information.
// It implements the error interface and includes field-specific errors.
type ValidationError struct {
    Field   string // Name of the invalid field
    Message string // Human-readable error message
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed on field '%s': %s", e.Field, e.Message)
}

// Usage
func ValidateUser(u *User) error {
    if u.Email == "" {
        return &ValidationError{
            Field:   "email",
            Message: "email is required",
        }
    }
    return nil
}
```

## Interfaces

### 1. Small, Focused Interfaces

```go
// ❌ Bad: Large interface with many methods
type UserService interface {
    Create(user *User) error
    Update(user *User) error
    Delete(id int) error
    FindByID(id int) (*User, error)
    FindByEmail(email string) (*User, error)
    List() ([]*User, error)
}

// ✅ Good: Small, single-purpose interfaces
type UserCreator interface {
    CreateUser(user *User) error
}

type UserFinder interface {
    FindUser(id int) (*User, error)
}

// Compose interfaces when needed
type UserRepository interface {
    UserCreator
    UserFinder
}
```

### 2. Accept Interfaces, Return Structs

```go
// ✅ Good: Function accepts interface (flexible)
func ProcessData(r io.Reader) error {
    data, err := io.ReadAll(r)
    // Process data...
    return nil
}

// Can be called with any io.Reader
file, _ := os.Open("data.txt")
ProcessData(file)

conn, _ := net.Dial("tcp", "example.com:80")
ProcessData(conn)
```

### 3. Empty Interface for Generic Types (Go < 1.18)

```go
// Store any type (use with caution)
func Store(key string, value interface{}) error {
    // Type assertion required when retrieving
    // Better to use generics in Go 1.18+
}

// ✅ Better with Go 1.18+ generics
func Store[T any](key string, value T) error {
    // Type-safe storage
}
```

## Struct Composition

### 1. Embedding Over Inheritance

```go
// Base type
type Logger struct {
    prefix string
}

func (l *Logger) Log(msg string) {
    fmt.Printf("[%s] %s\n", l.prefix, msg)
}

// ✅ Good: Embed Logger in Service
type UserService struct {
    Logger // Embedded field
    db     *sql.DB
}

// UserService automatically has Log method
func (s *UserService) CreateUser(u *User) error {
    s.Log("Creating user") // Inherited from Logger
    // Create user...
    return nil
}
```

### 2. Struct Tags for Metadata

```go
// User represents a user account with JSON serialization.
type User struct {
    ID        int       `json:"id" db:"id"`
    Email     string    `json:"email" db:"email" validate:"required,email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    // Use omitempty to exclude zero values
    UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
}
```

## Concurrency Patterns

### 1. Goroutines and WaitGroups

```go
// ExecuteParallel runs multiple tasks concurrently and waits for completion.
// Each task runs in its own goroutine. The function blocks until all
// tasks complete.
//
// Parameters:
//   - tasks: Slice of functions to execute concurrently
//
// The function does not return errors. Tasks should handle their own errors.
func ExecuteParallel(tasks []func()) {
    var wg sync.WaitGroup

    for _, task := range tasks {
        wg.Add(1)
        go func(t func()) {
            defer wg.Done()
            t()
        }(task)
    }

    wg.Wait()
}
```

### 2. Channels for Communication

```go
// Pipeline demonstrates channel-based data pipeline.
// Data flows: generator -> processor -> consumer
func Pipeline() {
    // Generator produces values
    gen := func() <-chan int {
        out := make(chan int)
        go func() {
            defer close(out)
            for i := 0; i < 10; i++ {
                out <- i
            }
        }()
        return out
    }

    // Processor transforms values
    proc := func(in <-chan int) <-chan int {
        out := make(chan int)
        go func() {
            defer close(out)
            for val := range in {
                out <- val * 2
            }
        }()
        return out
    }

    // Consumer receives processed values
    consumer := func(in <-chan int) {
        for val := range in {
            fmt.Println(val)
        }
    }

    // Connect pipeline
    nums := gen()
    processed := proc(nums)
    consumer(processed)
}
```

### 3. Select for Multiple Channels

```go
// Merge combines multiple channels into one output channel.
// It reads from all input channels concurrently and forwards
// values to the output channel. The output channel is closed
// when all input channels are closed.
//
// Parameters:
//   - channels: Variadic list of input channels to merge
//
// Returns:
//   - <-chan int: Output channel containing values from all inputs
func Merge(channels ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup

    // Start goroutine for each input channel
    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for val := range c {
                out <- val
            }
        }(ch)
    }

    // Close output when all inputs are done
    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}
```

### 4. Context for Cancellation

```go
// Worker performs work until context is cancelled.
// It checks ctx.Done() periodically to respect cancellation signals.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//
// Returns:
//   - error: Returns ctx.Err() if cancelled, nil on completion
func Worker(ctx context.Context) error {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            // Context cancelled, return immediately
            return ctx.Err()
        case <-ticker.C:
            // Do periodic work
            if err := doWork(); err != nil {
                return err
            }
        }
    }
}

// Usage with timeout
func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := Worker(ctx); err != nil {
        log.Printf("Worker stopped: %v", err)
    }
}
```

### 5. Mutex for Shared State

```go
// Counter provides thread-safe increment operations.
// It uses a mutex to protect the internal count from race conditions.
type Counter struct {
    mu    sync.Mutex
    count int
}

// Increment increases the counter by 1 in a thread-safe manner.
// Multiple goroutines can safely call this method concurrently.
func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

// Value returns the current counter value.
// It acquires a read lock to ensure consistent reads.
func (c *Counter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.count
}
```

### 6. RWMutex for Read-Heavy Workloads

```go
// Cache provides thread-safe in-memory caching with read/write locks.
// It optimizes for read-heavy workloads by using sync.RWMutex.
type Cache struct {
    mu    sync.RWMutex
    items map[string]interface{}
}

// Get retrieves a value from cache using a read lock.
// Multiple goroutines can read concurrently.
//
// Parameters:
//   - key: Cache key to look up
//
// Returns:
//   - interface{}: The cached value, or nil if not found
//   - bool: True if key exists, false otherwise
func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    val, ok := c.items[key]
    return val, ok
}

// Set stores a value in cache using a write lock.
// This blocks all other reads and writes.
//
// Parameters:
//   - key: Cache key
//   - value: Value to store
func (c *Cache) Set(key string, value interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[key] = value
}
```

## Defer for Cleanup

### 1. Resource Cleanup

```go
// ProcessFile opens, reads, and processes a file with automatic cleanup.
// The file is guaranteed to close even if an error occurs.
//
// Parameters:
//   - filename: Path to the file to process
//
// Returns:
//   - error: File opening, reading, or processing error
func ProcessFile(filename string) error {
    // Open file
    file, err := os.Open(filename)
    if err != nil {
        return fmt.Errorf("failed to open file: %w", err)
    }
    // Defer close - runs even if errors occur below
    defer file.Close()

    // Read and process
    data, err := io.ReadAll(file)
    if err != nil {
        return fmt.Errorf("failed to read file: %w", err)
    }

    return processData(data)
}
```

### 2. Unlock Mutex

```go
func (s *Service) UpdateState(key string, value interface{}) {
    s.mu.Lock()
    defer s.mu.Unlock() // Ensures unlock even if panic occurs

    // Update state
    s.state[key] = value
}
```

### 3. Rollback Transactions

```go
// CreateUser creates a user within a database transaction.
// The transaction is automatically rolled back if any error occurs.
//
// Parameters:
//   - ctx: Context for query timeout
//   - user: User data to insert
//
// Returns:
//   - error: Database or validation error
func (r *Repository) CreateUser(ctx context.Context, user *User) error {
    // Begin transaction
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }

    // Defer rollback - will be no-op if commit succeeds
    defer func() {
        if err != nil {
            tx.Rollback()
        }
    }()

    // Insert user
    _, err = tx.ExecContext(ctx, "INSERT INTO users ...", user.Email)
    if err != nil {
        return fmt.Errorf("failed to insert user: %w", err)
    }

    // Commit transaction
    if err = tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}
```

## Package Organization

### 1. Internal vs Pkg

```
project/
├── internal/          # Private packages (not importable by other projects)
│   ├── auth/         # Authentication logic
│   ├── database/     # Database clients
│   └── config/       # Configuration
├── pkg/              # Public packages (can be imported by others)
│   ├── models/       # Shared data models
│   └── utils/        # Utility functions
```

### 2. Package Naming

```go
// ❌ Bad: Stuttering names
package user
type UserService struct {} // user.UserService

// ✅ Good: Simple names
package user
type Service struct {} // user.Service

// ✅ Good: Clear usage
svc := user.NewService()
```

### 3. Init Functions (Use Sparingly)

```go
package database

var db *sql.DB

// init runs automatically when package is imported.
// Use sparingly - prefer explicit initialization.
func init() {
    var err error
    db, err = sql.Open("postgres", "...")
    if err != nil {
        panic(err) // Only panic in init if truly unrecoverable
    }
}
```

## Context Propagation

### 1. Pass Context as First Parameter

```go
// ✅ Good: Context is always first parameter
func FetchUser(ctx context.Context, userID int) (*User, error) {
    // Use ctx for database query timeout
    return db.QueryRowContext(ctx, "SELECT ...", userID)
}

// ❌ Bad: Context not first or missing
func FetchUser(userID int, ctx context.Context) (*User, error) {}
```

### 2. Context Values (Use Sparingly)

```go
type contextKey string

const requestIDKey contextKey = "requestID"

// SetRequestID adds request ID to context.
func SetRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey, id)
}

// GetRequestID extracts request ID from context.
func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(requestIDKey).(string); ok {
        return id
    }
    return ""
}
```

## Functional Options Pattern

### 1. Configure Structs Flexibly

```go
// Server represents an HTTP server with configurable options.
type Server struct {
    host    string
    port    int
    timeout time.Duration
}

// Option configures a Server.
type Option func(*Server)

// WithHost sets the server host.
func WithHost(host string) Option {
    return func(s *Server) {
        s.host = host
    }
}

// WithPort sets the server port.
func WithPort(port int) Option {
    return func(s *Server) {
        s.port = port
    }
}

// WithTimeout sets the server timeout.
func WithTimeout(timeout time.Duration) Option {
    return func(s *Server) {
        s.timeout = timeout
    }
}

// NewServer creates a server with default values and applies options.
func NewServer(opts ...Option) *Server {
    // Default values
    s := &Server{
        host:    "localhost",
        port:    8080,
        timeout: 30 * time.Second,
    }

    // Apply options
    for _, opt := range opts {
        opt(s)
    }

    return s
}

// Usage
server := NewServer(
    WithHost("0.0.0.0"),
    WithPort(3000),
    WithTimeout(60 * time.Second),
)
```

## Best Practices Summary

1. **Error Handling**: Always return errors, wrap with context, use sentinel errors
2. **Interfaces**: Keep small and focused, accept interfaces, return structs
3. **Composition**: Embed structs instead of inheritance
4. **Concurrency**: Use goroutines + channels, protect shared state with mutexes
5. **Defer**: Clean up resources (files, locks, transactions)
6. **Context**: Pass as first parameter, use for cancellation and timeouts
7. **Naming**: Avoid stuttering, use clear package names
8. **Documentation**: Comment all exported types and functions

## Anti-Patterns to Avoid

❌ **Don't use panic for normal errors**
❌ **Don't ignore errors** (`_ = someFunc()`)
❌ **Don't pass pointers to sync.Mutex**
❌ **Don't close channels from receiver side**
❌ **Don't use `interface{}` when you can use generics (Go 1.18+)**
❌ **Don't copy mutexes** (pass pointers to structs with mutexes)
❌ **Don't use `init()` for complex initialization**
