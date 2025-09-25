package ui

import (
	"context"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stormlightlabs/noteleaf/internal/models"
)

type AssertionHelpers struct{}

// TUITestSuite provides comprehensive testing infrastructure for BubbleTea models
// with channel-based control and signal handling for interactive testing
type TUITestSuite struct {
	t         *testing.T
	model     tea.Model
	program   *tea.Program
	msgChan   chan tea.Msg
	doneChan  chan struct{}
	outputBuf *ControlledOutput
	inputBuf  *ControlledInput
	mu        sync.RWMutex
	updates   []tea.Model
	views     []string
	finished  bool
	ctx       context.Context
	cancel    context.CancelFunc
}

// ControlledOutput captures program output for verification
type ControlledOutput struct {
	buf    []byte
	mu     sync.RWMutex
	writes [][]byte
}

func (co *ControlledOutput) Write(p []byte) (n int, err error) {
	co.mu.Lock()
	defer co.mu.Unlock()
	co.buf = append(co.buf, p...)
	co.writes = append(co.writes, append([]byte(nil), p...))
	return len(p), nil
}

func (co *ControlledOutput) GetOutput() []byte {
	co.mu.RLock()
	defer co.mu.RUnlock()
	return append([]byte(nil), co.buf...)
}

func (co *ControlledOutput) GetWrites() [][]byte {
	co.mu.RLock()
	defer co.mu.RUnlock()
	writes := make([][]byte, len(co.writes))
	for i, w := range co.writes {
		writes[i] = append([]byte(nil), w...)
	}
	return writes
}

// ControlledInput provides controlled input simulation
type ControlledInput struct {
	sequences []tea.Msg
	mu        sync.RWMutex
}

// Read is primarily for compatibility - actual input comes through channels
func (ci *ControlledInput) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (ci *ControlledInput) QueueMessage(msg tea.Msg) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ci.sequences = append(ci.sequences, msg)
}

// NewTUITestSuite creates a new TUI test suite with controlled I/O and channels
func NewTUITestSuite(t *testing.T, model tea.Model, opts ...TUITestOption) *TUITestSuite {
	ctx, cancel := context.WithCancel(context.Background())

	suite := &TUITestSuite{
		t:         t,
		model:     model,
		msgChan:   make(chan tea.Msg, 100),
		doneChan:  make(chan struct{}),
		outputBuf: &ControlledOutput{},
		inputBuf:  &ControlledInput{},
		updates:   []tea.Model{},
		views:     []string{},
		ctx:       ctx,
		cancel:    cancel,
	}

	for _, opt := range opts {
		opt(suite)
	}

	suite.setupProgram()

	t.Cleanup(func() {
		suite.Close()
	})

	return suite
}

// TUITestOption configures the test suite
type TUITestOption func(*TUITestSuite)

// WithInitialSize sets the initial terminal size by storing size for program initialization
func WithInitialSize(width, height int) TUITestOption {
	return func(suite *TUITestSuite) {
		suite.msgChan <- tea.WindowSizeMsg{Width: width, Height: height}
	}
}

// WithTimeout sets a global timeout for operations
func WithTimeout(timeout time.Duration) TUITestOption {
	return func(suite *TUITestSuite) {
		ctx, cancel := context.WithTimeout(suite.ctx, timeout)
		suite.ctx = ctx
		suite.cancel = cancel
	}
}

// setupProgram creates a program with controlled I/O by...
//
//	Disabling signals for testing
//	Disabling renderer for testing
func (suite *TUITestSuite) setupProgram() {
	// For unit testing, we'll directly test the model instead of running a full program
	suite.program = nil
}

// Start begins the test program in a goroutine with time to initialize
func (suite *TUITestSuite) Start() {
	if cmd := suite.model.Init(); cmd != nil {
		suite.executeCmd(cmd)
	}

	suite.mu.Lock()
	suite.updates = append(suite.updates, suite.model)
	suite.views = append(suite.views, suite.model.View())
	suite.mu.Unlock()
}

// SendKey sends a key press message to the model
func (suite *TUITestSuite) SendKey(keyType tea.KeyType, runes ...rune) error {
	msg := tea.KeyMsg{Type: keyType}
	if len(runes) > 0 {
		msg.Type = tea.KeyRunes
		msg.Runes = runes
	}
	return suite.SendMessage(msg)
}

// SendKeyString sends a string as key runes
func (suite *TUITestSuite) SendKeyString(s string) error {
	return suite.SendKey(tea.KeyRunes, []rune(s)...)
}

// SendMessage sends an arbitrary message to the model
func (suite *TUITestSuite) SendMessage(msg tea.Msg) error {
	newModel, cmd := suite.model.Update(msg)
	suite.model = newModel

	if cmd != nil {
		suite.executeCmd(cmd)
	}

	suite.mu.Lock()
	suite.updates = append(suite.updates, suite.model)
	suite.views = append(suite.views, suite.model.View())
	suite.mu.Unlock()

	return nil
}

// WaitFor waits for a condition to be met within the timeout
func (suite *TUITestSuite) WaitFor(condition func(tea.Model) bool, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(suite.ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("condition not met within timeout: %w", ctx.Err())
		case <-ticker.C:
			suite.mu.RLock()
			if len(suite.updates) > 0 {
				currentModel := suite.updates[len(suite.updates)-1]
				if condition(currentModel) {
					suite.mu.RUnlock()
					return nil
				}
			}
			suite.mu.RUnlock()
		}
	}
}

// WaitForView waits for a view to contain specific content
func (suite *TUITestSuite) WaitForView(contains string, timeout time.Duration) error {
	return suite.WaitFor(func(model tea.Model) bool {
		view := model.View()
		return len(view) > 0 && containsString(view, contains)
	}, timeout)
}

// GetCurrentModel returns the latest model state (thread-safe)
func (suite *TUITestSuite) GetCurrentModel() tea.Model {
	suite.mu.RLock()
	defer suite.mu.RUnlock()

	if len(suite.updates) == 0 {
		return suite.model
	}
	return suite.updates[len(suite.updates)-1]
}

// GetCurrentView returns the latest view output
func (suite *TUITestSuite) GetCurrentView() string {
	model := suite.GetCurrentModel()
	return model.View()
}

// GetOutput returns all captured output
func (suite *TUITestSuite) GetOutput() []byte {
	return suite.outputBuf.GetOutput()
}

// executeCmd executes any commands returned by model updates
//
//	For unit testing, we ignore commands or handle specific ones we care about
//	This could be extended to handle specific command types if needed
func (suite *TUITestSuite) executeCmd(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
}

// Close properly shuts down the test suite
func (suite *TUITestSuite) Close() {
	if !suite.finished {
		suite.finished = true
		suite.cancel()
	}
}

// SimulateKeySequence sends a sequence of keys with timing
func (suite *TUITestSuite) SimulateKeySequence(keys []KeyWithTiming) error {
	for _, key := range keys {
		if err := suite.SendKey(key.KeyType, key.Runes...); err != nil {
			return fmt.Errorf("failed to send key %v: %w", key.KeyType, err)
		}
		if key.Delay > 0 {
			time.Sleep(key.Delay)
		}
	}
	return nil
}

// KeyWithTiming represents a key press with optional delay
type KeyWithTiming struct {
	KeyType tea.KeyType
	Runes   []rune
	Delay   time.Duration
}

// MockTaskRepository provides a mock implementation for testing
type MockTaskRepository struct {
	tasks   map[int64]*models.Task
	updated []*models.Task
	mu      sync.RWMutex
}

func NewMockTaskRepository() *MockTaskRepository {
	return &MockTaskRepository{
		tasks: make(map[int64]*models.Task),
	}
}

func (m *MockTaskRepository) AddTask(task *models.Task) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks[task.ID] = task
}

func (m *MockTaskRepository) GetUpdatedTasks() []*models.Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*models.Task, len(m.updated))
	copy(result, m.updated)
	return result
}

func (ah *AssertionHelpers) AssertModelState(t *testing.T, suite *TUITestSuite, checker func(tea.Model) bool, msg string) {
	t.Helper()
	model := suite.GetCurrentModel()
	if !checker(model) {
		t.Errorf("Model state assertion failed: %s", msg)
	}
}

func (ah *AssertionHelpers) AssertViewContains(t *testing.T, suite *TUITestSuite, expected string, msg string) {
	t.Helper()
	view := suite.GetCurrentView()
	if !containsString(view, expected) {
		t.Errorf("View assertion failed: %s\nView content: %s\nExpected to contain: %s", msg, view, expected)
	}
}

func (ah *AssertionHelpers) AssertViewNotContains(t *testing.T, suite *TUITestSuite, unexpected string, msg string) {
	t.Helper()
	view := suite.GetCurrentView()
	if containsString(view, unexpected) {
		t.Errorf("View assertion failed: %s\nView content: %s\nShould not contain: %s", msg, view, unexpected)
	}
}

var Expect = AssertionHelpers{}

// Helper function to check if string contains substring
func containsString(haystack, needle string) bool {
	if needle == "" {
		return true
	}

	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

// Test generators for switch case coverage
type SwitchCaseTest struct {
	Name        string
	Input       any
	Expected    any
	ShouldError bool
	Setup       func(*TUITestSuite)
	Verify      func(*testing.T, *TUITestSuite)
}

// CreateTestSuiteWithModel is a helper to create a test suite with a specific model
//
//	This should be used in individual test files where the model type is known
func CreateTestSuiteWithModel(t *testing.T, model tea.Model, opts ...TUITestOption) *TUITestSuite {
	return NewTUITestSuite(t, model, opts...)
}
