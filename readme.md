# Ludwig: AI Task Orchestrator

Ludwig is an AI-powered task orchestrator that automates project work through integrated AI clients (currently Gemini). It manages task execution, git workflows, and human review cycles through a command-line interface.

## Project Structure

```
ludwig/
├── cmd/                              # Command-line entry point
│   └── main.go                       # Application initialization
├── internal/
│   ├── cli/                          # CLI interface and display
│   │   ├── cli.go                    # Main CLI loop
│   │   ├── commandPallete.go         # Command definitions
│   │   └── kanban.go                 # Kanban board display
│   ├── config/                       # Configuration management
│   │   └── config.json               # User config location: ~/.ai-orchestrator/config.json
│   ├── mcp/                          # Model context protocol (future enhancement)
│   ├── orchestrator/                 # Core orchestration logic
│   │   ├── orchestrator.go           # Main orchestrator loop
│   │   ├── prompts.go                # System prompts for AI agents
│   │   ├── git.go                    # Git operations
│   │   └── clients/
│   │       └── gemini.go             # Gemini AI client
│   ├── storage/                      # Data persistence
│   │   ├── taskStorage.go            # Task file storage
│   │   ├── responseStorage.go        # AI response streaming
│   │   └── streamingWriter.go        # Stream writing utilities
│   ├── types/                        # Core data types
│   │   └── task.go                   # Task definition
│   └── utils/                        # Utility functions
│       ├── cliUtils.go               # CLI input/output utilities
│       └── mapUtils.go               # Data transformation utilities
├── test/                             # Test suite (136+ tests)
│   ├── cli/
│   ├── config/
│   ├── orchestrator/
│   ├── storage/
│   ├── types/
│   └── utils/
├── go.mod                            # Go module definition
├── go.sum                            # Dependency checksums
└── readme.md                         # This file
```

## Development Setup

### Prerequisites

- **Go 1.25.5 or later**
- **Git** (for version control and branch operations)

### Build

```bash
# Build the Ludwig binary
go build -o ludwig ./cmd/main.go

# Or use the standard Go build
go build ./cmd/main.go
```

### Testing

Ludwig has comprehensive test coverage with **136+ tests** across all modules.

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run specific test package
go test ./test/storage -v
go test ./test/orchestrator -v
go test ./test/types -v

# Run a specific test
go test ./test/types -run TestStatusString -v

# Run tests with coverage
go test ./... -cover
```

### Running

```bash
# Start the CLI application
./ludwig

# Or directly with go run
go run ./cmd/main.go
```

## Development Workflow

### Before Making Changes

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Check the test suite status**
   ```bash
   go test ./...
   ```

### During Development

1. **Make incremental changes** with focused commits
   ```bash
   git add <modified files>
   git commit -m "Clear, descriptive commit message"
   ```

2. **Run tests frequently** to catch issues early
   ```bash
   go test ./...
   ```

### Before Committing to Main

1. **Run full test suite**
   ```bash
   go test ./...
   ```

2. **Ensure all tests pass** - if tests fail:
   - **First**: Understand the root cause
   - **Investigate**: Does the source code need fixing, or do the tests need updating?
   - **Fix source code first** if the implementation is wrong
   - **Only update tests** if the behavior change is intentional and correct

3. **Add tests for new features** (if test files already exist in the codebase)
   - Add tests alongside your implementation
   - Use the existing test patterns as reference
   - Ensure new tests pass before committing

4. **Build before pushing**
   ```bash
   go build ./cmd/main.go
   ```

## Data Types

### Task States

- **Pending**: Waiting to be processed by the orchestrator
- **In Progress**: Currently being processed by an AI agent
- **Needs Review**: Waiting for human feedback on a design decision
- **Completed**: Task finished successfully

### Task Structure

```go
type Task struct {
    ID             string           // Unique identifier
    Name           string           // Task description
    Status         Status           // Current status
    BranchName     string           // Associated git branch
    WorkInProgress string           // Intermediate work progress
    Review         *ReviewRequest   // Design decision request
    ReviewResponse *ReviewResponse  // Human response to review
    ResponseFile   string           // Path to AI response file
}
```

## CLI Commands

| Command | Usage | Description |
|---------|-------|-------------|
| `add` | `add <task description>` | Add a new task (multiple words, no quotes needed) |
| `start` | `start` | Start the AI orchestrator to process tasks |
| `stop` | `stop` | Stop the orchestrator |
| `clear` | `clear` | Clear the screen |
| `help` | `help` | Show available commands |
| `exit` | `exit` | Exit the application |

## Orchestrator Workflow

1. **Initialization**: Loads tasks from storage and creates task branches
2. **Polling**: Checks for pending tasks and processes them in order
3. **AI Processing**: Sends tasks to Gemini with system prompt and task description
4. **Review Detection**: Parses responses for `---NEEDS_REVIEW---` markers
5. **Review Handling**: If review needed, waits for human decision
6. **Completion**: Marks tasks complete and checks out to main branch

### Task Processing Flow

```
Pending Task
    ↓
Create Git Branch
    ↓
Send to AI (with SystemPrompt + Task)
    ↓
Does response contain ---NEEDS_REVIEW---?
    ├─ Yes: Needs Review
    │   ↓
    │   Human provides decision
    │   ↓
    │   Resume with user feedback
    │   ↓
    │   Completed
    └─ No: Completed immediately
```

## Git Integration

- Each task gets its own feature branch: `ludwig/<task-name>`
- AI agents are instructed to make regular commits
- Branches are checked out automatically for task processing
- After task completion, orchestrator checks out back to `main`

## Testing Guidelines

### Adding New Tests

When adding new functionality:

1. Check if tests already exist in the codebase (they do - 136+ tests)
2. Add tests alongside your implementation
3. Use existing test patterns from the `test/` directory
4. Follow Go testing conventions (function names start with `Test`)
5. Include both happy path and error cases

### Test Organization

Tests mirror the package structure:

```
internal/storage/taskStorage.go      → test/storage/taskStorage_test.go
internal/types/task.go              → test/types/task_test.go
internal/orchestrator/prompts.go    → test/orchestrator/prompts_test.go
```

### Test Patterns

**Table-driven tests** (for multiple scenarios):
```go
tests := []struct {
    name     string
    input    string
    expected string
}{
    {"case 1", "input1", "expected1"},
    {"case 2", "input2", "expected2"},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test code
    })
}
```

**Cleanup/Setup**:
```go
func TestSomething(t *testing.T) {
    defer cleanupTestStorage(t)
    // test code
}
```

## Current Test Coverage

- **Total Tests**: 136+
- **Modules Covered**:
  - `storage/`: 45 tests
  - `orchestrator/`: 38 tests
  - `types/`: 28 tests
  - `utils/`: 15 tests
  - `cli/`: 14 tests
  - `config/`: 3 tests

For detailed coverage information, see [TEST_COVERAGE_IMPROVEMENTS.md](./TEST_COVERAGE_IMPROVEMENTS.md).

## Common Development Tasks

### Add a new CLI command
1. Edit `internal/cli/commandPallete.go`
2. Add new command struct to `PalleteCommands()`
3. Add tests in `test/cli/`

### Add storage functionality
1. Extend `internal/storage/taskStorage.go`
2. Add tests in `test/storage/`
3. Run tests: `go test ./test/storage -v`

### Modify task types
1. Update `internal/types/task.go`
2. Add/update tests in `test/types/`
3. Update related storage and orchestrator logic

### Change AI prompt behavior
1. Edit `internal/orchestrator/prompts.go`
2. Update tests in `test/orchestrator/`
3. Test with actual tasks to verify AI understanding

## Troubleshooting

### Tests Failing

1. Run tests with verbose output: `go test ./... -v`
2. Check if failure is in source code or test expectations
3. For storage tests, ensure `~/.ai-orchestrator/` directory is writable
4. Clean up: `rm -rf ~/.ai-orchestrator/tasks.json`

### Build Issues

1. Check Go version: `go version` (need 1.25.5+)
2. Get dependencies: `go mod download`
3. Clean build: `go clean && go build ./cmd/main.go`

### Orchestrator Not Starting

1. Check if Gemini API key is set in environment
2. Verify git repository is initialized: `git status`
3. Check task storage is accessible: `ls ~/.ai-orchestrator/`

## Dependencies

```
golang.org/x/term v0.38.0    # Terminal control
github.com/google/uuid v1.6.0 # UUID generation
golang.org/x/sys v0.39.0      # System calls
```

## Contributing

When contributing:

1. Follow the existing code structure and naming conventions
2. Write tests for new functionality (tests exist in codebase)
3. Run full test suite before committing
4. Use clear, descriptive commit messages
5. Create git branches for features
6. Ensure all tests pass

## License

MIT (adjust if needed)

## Next Steps / Future Enhancements

- [ ] Support additional AI clients beyond Gemini
- [ ] Model Context Protocol (MCP) integration
- [ ] Advanced task scheduling and prioritization
- [ ] Web UI for task management
- [ ] Webhook integration for automated task triggers
- [ ] Task templates and presets
- [ ] Performance metrics and analytics
