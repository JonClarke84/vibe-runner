# Go Backend Code Examples

Annotated code examples showing idiomatic Go patterns, TDD workflow, and comprehensive documentation.

## Example 1: HTTP API with TDD

### Step 1: Write Test First

```go
// api/user_handler_test.go
package api_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "myapp/api"
)

func TestHandleCreateUser_ValidInput_ReturnsCreated(t *testing.T) {
    // Arrange: Create request body
    reqBody := map[string]string{
        "email": "test@example.com",
        "name":  "Test User",
    }
    body, _ := json.Marshal(reqBody)

    // Create HTTP request
    req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")

    // Create response recorder
    rr := httptest.NewRecorder()

    // Act: Call handler
    handler := api.HandleCreateUser()
    handler.ServeHTTP(rr, req)

    // Assert: Check response
    if rr.Code != http.StatusCreated {
        t.Errorf("got status %d, want %d", rr.Code, http.StatusCreated)
    }

    // Verify response body
    var response map[string]interface{}
    if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
        t.Fatalf("failed to parse response: %v", err)
    }

    if response["email"] != "test@example.com" {
        t.Errorf("got email %v, want test@example.com", response["email"])
    }
}
```

### Step 2: Implement Handler

```go
// api/user_handler.go
package api

import (
    "encoding/json"
    "fmt"
    "net/http"
)

// CreateUserRequest represents the request body for user creation.
// It contains the required fields for creating a new user account.
type CreateUserRequest struct {
    Email string `json:"email"` // User's email address (required, must be unique)
    Name  string `json:"name"`  // User's display name (required)
}

// CreateUserResponse represents the response body after user creation.
// It includes the newly created user's details.
type CreateUserResponse struct {
    ID    int    `json:"id"`    // Assigned user ID
    Email string `json:"email"` // Confirmed email address
    Name  string `json:"name"`  // Confirmed display name
}

// HandleCreateUser handles POST /users requests to create new users.
// It validates the request body, creates the user in the database,
// and returns the created user with a 201 status code.
//
// Request Body:
//   {
//     "email": "user@example.com",
//     "name": "User Name"
//   }
//
// Response Codes:
//   - 201 Created: User successfully created
//   - 400 Bad Request: Invalid JSON or validation error
//   - 409 Conflict: Email already exists
//   - 500 Internal Server Error: Database or server error
//
// Example:
//   curl -X POST http://localhost:8080/users \
//     -H "Content-Type: application/json" \
//     -d '{"email":"test@example.com","name":"Test User"}'
func HandleCreateUser() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Parse request body
        var req CreateUserRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "invalid request body", http.StatusBadRequest)
            return
        }

        // Validate input
        if req.Email == "" {
            http.Error(w, "email is required", http.StatusBadRequest)
            return
        }
        if req.Name == "" {
            http.Error(w, "name is required", http.StatusBadRequest)
            return
        }

        // Create user (simplified - would normally use service/repository)
        user := &User{
            ID:    1, // Would be assigned by database
            Email: req.Email,
            Name:  req.Name,
        }

        // Return response
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(CreateUserResponse{
            ID:    user.ID,
            Email: user.Email,
            Name:  user.Name,
        })
    }
}
```

## Example 2: Service Layer with Repository Pattern

### Repository Interface and Implementation

```go
// repository/user_repository.go
package repository

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
)

var (
    // ErrUserNotFound is returned when a user query returns no results.
    ErrUserNotFound = errors.New("user not found")

    // ErrDuplicateEmail is returned when attempting to create a user
    // with an email that already exists in the database.
    ErrDuplicateEmail = errors.New("email already exists")
)

// User represents a user account in the system.
type User struct {
    ID    int    // Unique user identifier
    Email string // User's email address (unique)
    Name  string // User's display name
}

// UserRepository defines database operations for user management.
// Implementations should be thread-safe and handle database errors appropriately.
type UserRepository interface {
    // Create inserts a new user into the database.
    // Returns ErrDuplicateEmail if the email already exists.
    Create(ctx context.Context, user *User) error

    // FindByID retrieves a user by their unique ID.
    // Returns ErrUserNotFound if no user exists with the given ID.
    FindByID(ctx context.Context, id int) (*User, error)

    // FindByEmail retrieves a user by their email address.
    // Returns ErrUserNotFound if no user exists with the given email.
    FindByEmail(ctx context.Context, email string) (*User, error)
}

// PostgresUserRepository implements UserRepository using PostgreSQL.
// It provides thread-safe database operations for user management.
type PostgresUserRepository struct {
    db *sql.DB
}

// NewPostgresUserRepository creates a new PostgreSQL-backed user repository.
//
// Parameters:
//   - db: Database connection pool (must be non-nil and connected)
//
// Returns:
//   - *PostgresUserRepository: Configured repository instance
func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
    return &PostgresUserRepository{db: db}
}

// Create inserts a new user into the database and assigns an ID.
// The user's ID field will be updated with the database-assigned value.
//
// Parameters:
//   - ctx: Context for query timeout and cancellation
//   - user: User to create (ID will be assigned, Email must be unique)
//
// Returns:
//   - error: Returns ErrDuplicateEmail if email exists,
//            or wrapped database error for other failures
func (r *PostgresUserRepository) Create(ctx context.Context, user *User) error {
    query := `
        INSERT INTO users (email, name)
        VALUES ($1, $2)
        RETURNING id
    `

    // Execute query and scan returned ID
    err := r.db.QueryRowContext(ctx, query, user.Email, user.Name).Scan(&user.ID)
    if err != nil {
        // Check for unique constraint violation (PostgreSQL error code 23505)
        if isUniqueViolation(err) {
            return ErrDuplicateEmail
        }
        return fmt.Errorf("failed to create user: %w", err)
    }

    return nil
}

// FindByID retrieves a user by their unique ID.
//
// Parameters:
//   - ctx: Context for query timeout and cancellation
//   - id: User ID to look up
//
// Returns:
//   - *User: The found user (nil if not found)
//   - error: Returns ErrUserNotFound if user doesn't exist,
//            or wrapped database error for other failures
func (r *PostgresUserRepository) FindByID(ctx context.Context, id int) (*User, error) {
    query := `SELECT id, email, name FROM users WHERE id = $1`

    user := &User{}
    err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Email, &user.Name)
    if err == sql.ErrNoRows {
        return nil, ErrUserNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("failed to find user by id: %w", err)
    }

    return user, nil
}

// FindByEmail retrieves a user by their email address.
// Email lookup is case-insensitive.
//
// Parameters:
//   - ctx: Context for query timeout and cancellation
//   - email: Email address to look up (case-insensitive)
//
// Returns:
//   - *User: The found user (nil if not found)
//   - error: Returns ErrUserNotFound if user doesn't exist,
//            or wrapped database error for other failures
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
    query := `SELECT id, email, name FROM users WHERE LOWER(email) = LOWER($1)`

    user := &User{}
    err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Email, &user.Name)
    if err == sql.ErrNoRows {
        return nil, ErrUserNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("failed to find user by email: %w", err)
    }

    return user, nil
}

// isUniqueViolation checks if a database error is a unique constraint violation.
// This is PostgreSQL-specific error checking.
func isUniqueViolation(err error) bool {
    // Implementation would check for PostgreSQL error code 23505
    // Using pq library: pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505"
    return false // Simplified for example
}
```

### Service Layer with Business Logic

```go
// service/user_service.go
package service

import (
    "context"
    "errors"
    "fmt"
    "myapp/repository"
    "regexp"
)

var (
    // ErrInvalidEmail is returned when an email address is not valid format.
    ErrInvalidEmail = errors.New("invalid email format")

    // ErrEmptyName is returned when a user's name is empty.
    ErrEmptyName = errors.New("name cannot be empty")
)

// UserService handles business logic for user operations.
// It validates inputs, enforces business rules, and coordinates
// with the repository layer for data persistence.
type UserService struct {
    repo repository.UserRepository
}

// NewUserService creates a new user service.
//
// Parameters:
//   - repo: User repository for data access (must be non-nil)
//
// Returns:
//   - *UserService: Configured service instance
func NewUserService(repo repository.UserRepository) *UserService {
    return &UserService{repo: repo}
}

// CreateUser creates a new user account after validation.
// It validates the email format and name, then persists the user.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - email: User's email address (must be valid format and unique)
//   - name: User's display name (must be non-empty)
//
// Returns:
//   - *repository.User: The created user with assigned ID
//   - error: Returns ErrInvalidEmail for invalid email format,
//            ErrEmptyName if name is empty,
//            repository.ErrDuplicateEmail if email already exists,
//            or other repository errors
//
// Example:
//   user, err := svc.CreateUser(ctx, "test@example.com", "Test User")
//   if errors.Is(err, ErrInvalidEmail) {
//       // Handle validation error
//   }
func (s *UserService) CreateUser(ctx context.Context, email, name string) (*repository.User, error) {
    // Validate email format
    if !isValidEmail(email) {
        return nil, ErrInvalidEmail
    }

    // Validate name
    if name == "" {
        return nil, ErrEmptyName
    }

    // Create user entity
    user := &repository.User{
        Email: email,
        Name:  name,
    }

    // Persist to database
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    return user, nil
}

// GetUser retrieves a user by ID.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - id: User ID to retrieve
//
// Returns:
//   - *repository.User: The found user
//   - error: Returns repository.ErrUserNotFound if user doesn't exist
func (s *UserService) GetUser(ctx context.Context, id int) (*repository.User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return user, nil
}

// isValidEmail checks if an email address is in valid format.
// It uses a simple regex pattern for basic validation.
//
// Parameters:
//   - email: Email address to validate
//
// Returns:
//   - bool: True if email format is valid
func isValidEmail(email string) bool {
    // Simple email regex (production would use more robust validation)
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return emailRegex.MatchString(email)
}
```

### Service Tests with Mocks

```go
// service/user_service_test.go
package service_test

import (
    "context"
    "errors"
    "testing"
    "myapp/repository"
    "myapp/service"
)

// MockUserRepository implements repository.UserRepository for testing.
type MockUserRepository struct {
    CreateFunc      func(ctx context.Context, user *repository.User) error
    FindByIDFunc    func(ctx context.Context, id int) (*repository.User, error)
    FindByEmailFunc func(ctx context.Context, email string) (*repository.User, error)
}

func (m *MockUserRepository) Create(ctx context.Context, user *repository.User) error {
    if m.CreateFunc != nil {
        return m.CreateFunc(ctx, user)
    }
    return errors.New("not implemented")
}

func (m *MockUserRepository) FindByID(ctx context.Context, id int) (*repository.User, error) {
    if m.FindByIDFunc != nil {
        return m.FindByIDFunc(ctx, id)
    }
    return nil, errors.New("not implemented")
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*repository.User, error) {
    if m.FindByEmailFunc != nil {
        return m.FindByEmailFunc(ctx, email)
    }
    return nil, errors.New("not implemented")
}

func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name      string
        email     string
        userName  string
        mockSetup func(*MockUserRepository)
        wantErr   error
    }{
        {
            name:     "valid input creates user",
            email:    "test@example.com",
            userName: "Test User",
            mockSetup: func(m *MockUserRepository) {
                m.CreateFunc = func(ctx context.Context, user *repository.User) error {
                    user.ID = 1 // Simulate database assigning ID
                    return nil
                }
            },
            wantErr: nil,
        },
        {
            name:     "invalid email returns error",
            email:    "invalid-email",
            userName: "Test User",
            mockSetup: func(m *MockUserRepository) {
                // No mock setup needed - validation happens before repository call
            },
            wantErr: service.ErrInvalidEmail,
        },
        {
            name:     "empty name returns error",
            email:    "test@example.com",
            userName: "",
            mockSetup: func(m *MockUserRepository) {
                // No mock setup needed
            },
            wantErr: service.ErrEmptyName,
        },
        {
            name:     "duplicate email returns error",
            email:    "existing@example.com",
            userName: "Test User",
            mockSetup: func(m *MockUserRepository) {
                m.CreateFunc = func(ctx context.Context, user *repository.User) error {
                    return repository.ErrDuplicateEmail
                }
            },
            wantErr: repository.ErrDuplicateEmail,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mock
            mockRepo := &MockUserRepository{}
            tt.mockSetup(mockRepo)

            // Create service
            svc := service.NewUserService(mockRepo)

            // Execute
            ctx := context.Background()
            user, err := svc.CreateUser(ctx, tt.email, tt.userName)

            // Assert error
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            // If no error expected, verify user was created
            if err == nil {
                if user == nil {
                    t.Error("CreateUser() returned nil user with no error")
                }
                if user.Email != tt.email {
                    t.Errorf("user email = %s, want %s", user.Email, tt.email)
                }
                if user.ID == 0 {
                    t.Error("user ID was not assigned")
                }
            }
        })
    }
}
```

## Example 3: WebSocket Server with Concurrency

```go
// websocket/server.go
package websocket

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "sync"
    "time"

    "github.com/gorilla/websocket"
)

// Message represents a WebSocket message with event type and data.
type Message struct {
    Event string          `json:"e"` // Event type (e.g., "join", "message")
    Data  json.RawMessage `json:"d"` // Event-specific data
}

// Client represents a connected WebSocket client.
// Each client has its own goroutines for reading and writing messages.
type Client struct {
    id       int                // Unique client identifier
    conn     *websocket.Conn    // WebSocket connection
    send     chan Message       // Outbound message queue
    server   *Server            // Reference to server
    ctx      context.Context    // Context for cancellation
    cancel   context.CancelFunc // Cancel function
}

// Server manages WebSocket connections and message broadcasting.
// It runs a central hub goroutine that coordinates message distribution.
type Server struct {
    clients    map[int]*Client // Connected clients (protected by mutex)
    register   chan *Client    // Channel for new client registration
    unregister chan *Client    // Channel for client disconnection
    broadcast  chan Message    // Channel for broadcasting to all clients
    mu         sync.RWMutex    // Protects clients map
    nextID     int             // Next client ID (atomic)
}

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    // Allow all origins for development (restrict in production)
    CheckOrigin: func(r *http.Request) bool { return true },
}

// NewServer creates a new WebSocket server.
// It initializes channels and the client map but doesn't start goroutines yet.
//
// Returns:
//   - *Server: Configured server instance ready to run
func NewServer() *Server {
    return &Server{
        clients:    make(map[int]*Client),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        broadcast:  make(chan Message, 256), // Buffered for performance
        nextID:     1,
    }
}

// Run starts the server's hub goroutine.
// It handles client registration, unregistration, and message broadcasting.
// This function blocks until the provided context is cancelled.
//
// The hub goroutine coordinates all client connections and ensures
// thread-safe access to the clients map.
//
// Parameters:
//   - ctx: Context for server shutdown
func (s *Server) Run(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            // Server shutdown - disconnect all clients
            s.mu.Lock()
            for _, client := range s.clients {
                close(client.send)
                client.conn.Close()
            }
            s.mu.Unlock()
            return

        case client := <-s.register:
            // New client connected
            s.mu.Lock()
            s.clients[client.id] = client
            s.mu.Unlock()
            log.Printf("Client %d connected (total: %d)", client.id, len(s.clients))

        case client := <-s.unregister:
            // Client disconnected
            s.mu.Lock()
            if _, ok := s.clients[client.id]; ok {
                delete(s.clients, client.id)
                close(client.send)
            }
            s.mu.Unlock()
            log.Printf("Client %d disconnected (total: %d)", client.id, len(s.clients))

        case message := <-s.broadcast:
            // Broadcast message to all clients
            s.mu.RLock()
            for _, client := range s.clients {
                select {
                case client.send <- message:
                    // Message queued successfully
                default:
                    // Client's send buffer is full, disconnect slow client
                    log.Printf("Client %d send buffer full, disconnecting", client.id)
                    s.unregister <- client
                }
            }
            s.mu.RUnlock()
        }
    }
}

// HandleWebSocket upgrades HTTP connections to WebSocket and manages clients.
// It creates a new client, starts read/write goroutines, and registers the client.
//
// This handler should be mounted at the WebSocket endpoint:
//   http.HandleFunc("/ws", server.HandleWebSocket)
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    // Upgrade HTTP connection to WebSocket
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket upgrade failed: %v", err)
        return
    }

    // Create client with cancellable context
    ctx, cancel := context.WithCancel(context.Background())
    client := &Client{
        id:     s.nextID,
        conn:   conn,
        send:   make(chan Message, 256),
        server: s,
        ctx:    ctx,
        cancel: cancel,
    }
    s.nextID++

    // Register client
    s.register <- client

    // Start client goroutines
    go client.readPump()
    go client.writePump()
}

// readPump reads messages from the WebSocket connection.
// It runs in its own goroutine per client and handles incoming messages.
// The goroutine exits when the connection is closed or an error occurs.
//
// This method sets read deadlines to detect disconnected clients.
func (c *Client) readPump() {
    defer func() {
        c.server.unregister <- c
        c.conn.Close()
        c.cancel()
    }()

    // Configure WebSocket connection
    c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        return nil
    })

    // Read messages loop
    for {
        var msg Message
        err := c.conn.ReadJSON(&msg)
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("Client %d read error: %v", c.id, err)
            }
            break
        }

        // Handle message (simplified - would route to handlers)
        log.Printf("Client %d sent: %s", c.id, msg.Event)

        // Example: Echo message back to all clients
        c.server.broadcast <- msg
    }
}

// writePump sends messages from the send channel to the WebSocket connection.
// It runs in its own goroutine per client and handles outbound messages.
// The goroutine exits when the send channel is closed or an error occurs.
//
// This method implements a ping ticker to keep connections alive.
func (c *Client) writePump() {
    // Ticker for sending ping messages (keepalive)
    ticker := time.NewTicker(54 * time.Second)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()

    for {
        select {
        case <-c.ctx.Done():
            // Client context cancelled
            return

        case message, ok := <-c.send:
            // Set write deadline
            c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

            if !ok {
                // Send channel closed, close connection
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            // Write message as JSON
            if err := c.conn.WriteJSON(message); err != nil {
                log.Printf("Client %d write error: %v", c.id, err)
                return
            }

        case <-ticker.C:
            // Send ping message for keepalive
            c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}

// Broadcast sends a message to all connected clients.
// This is a convenience method for external code to broadcast messages.
//
// Parameters:
//   - event: Event type identifier
//   - data: Event payload (will be marshaled to JSON)
func (s *Server) Broadcast(event string, data interface{}) error {
    dataJSON, err := json.Marshal(data)
    if err != nil {
        return err
    }

    msg := Message{
        Event: event,
        Data:  dataJSON,
    }

    s.broadcast <- msg
    return nil
}
```

These examples demonstrate:
- ✅ TDD workflow (write tests first)
- ✅ Comprehensive function documentation
- ✅ Error handling with wrapped context
- ✅ Interface-based design
- ✅ Proper use of goroutines and channels
- ✅ Context for cancellation
- ✅ Thread-safe concurrent access
- ✅ Table-driven tests
- ✅ Mocking for unit tests
