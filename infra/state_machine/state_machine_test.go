package state_machine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pixie-sh/core-go/pkg/types/slices"

	perrors "github.com/pixie-sh/errors-go"
	"github.com/stretchr/testify/assert"

	mapper "github.com/pixie-sh/core-go/pkg/models/serializer"
)

// MockStorage implements the StateMachineStorage interface for testing
type MockStorage struct {
	storedModel MachineModel
	t           *testing.T
}

func (ms *MockStorage) Store(model MachineModel) error {
	ms.storedModel = model
	blob, err := mapper.ToJSONB(model)
	assert.Nil(ms.t, err)
	assert.NotNil(ms.t, blob)

	return nil
}

func (ms *MockStorage) Get(machineID string) (MachineModel, error) {
	if ms.storedModel.ID == machineID {
		return ms.storedModel, nil
	}
	return MachineModel{}, errors.New("machine not found")
}

func TestNewMachine(t *testing.T) {
	storage := &MockStorage{t: t}
	m := NewMachine(context.Background(), "test-machine", storage)

	assert.Equal(t, "test-machine", m.id)
	assert.Equal(t, storage, m.storage)
	assert.Empty(t, m.states)
	assert.Empty(t, m.visitedStates)
	assert.Empty(t, m.transitions)
	assert.Empty(t, m.conditionalsTransitions)
	assert.NotEmpty(t, m.guid)
}

func TestAddState(t *testing.T) {
	m := NewMachine(context.Background(), "test-machine", &MockStorage{t: t})
	m.AddState("state1")
	m.AddState("state2")

	assert.Len(t, m.states, 2)
	assert.Contains(t, m.states, State("state1"))
	assert.Contains(t, m.states, State("state2"))
}

func TestSetInitialState(t *testing.T) {
	m := NewMachine(context.Background(), "test-machine", &MockStorage{t: t})
	m.AddState("state1")

	err := m.SetInitialState("state1")
	assert.NoError(t, err)
	assert.Equal(t, State("state1"), m.currentState)
	assert.NotZero(t, m.currentStateAt)

	err = m.SetInitialState("non-existent")
	assert.Error(t, err)
}

func TestAddTransition(t *testing.T) {
	m := NewMachine(context.Background(), "test-machine", &MockStorage{t: t})
	m.AddState("state1")
	m.AddState("state2")

	err := m.AddTransition("state1", "event1", "state2")
	assert.NoError(t, err)
	assert.Equal(t, State("state2"), m.transitions["state1"]["event1"])

	err = m.AddTransition("non-existent", "event1", "state2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'from' state does not exist")

	err = m.AddTransition("state1", "event1", "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'to' state does not exist")
}

func TestTrigger(t *testing.T) {
	m := NewMachine(context.Background(), "test-machine", &MockStorage{t: t})
	m.AddState("state1")
	m.AddState("state2")
	m.SetInitialState("state1")
	m.AddTransition("state1", "event1", "state2")

	ctx := context.TODO()
	newState, err := m.Trigger(ctx, "event1")
	assert.NoError(t, err)
	assert.Equal(t, State("state2"), newState)
	assert.Equal(t, State("state2"), m.CurrentState())

	st := slices.Find(m.visitedStates, func(s VisitedState) bool {
		return s.State == "state1"
	})
	assert.NotEmpty(t, st)

	_, err = m.Trigger(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid transition")
}

func TestSaveAndRestore(t *testing.T) {
	storage := &MockStorage{t: t}
	m := NewMachine(context.Background(), "test-machine", storage)
	m.AddState("state1")
	m.AddState("state2")
	m.SetInitialState("state1")
	m.AddTransition("state1", "event1", "state2")

	err := m.Save()
	assert.NoError(t, err)

	newMachine := NewMachine(context.Background(), "test-machine", storage)
	err = newMachine.Restore()
	assert.NoError(t, err)

	assert.Equal(t, m.states, newMachine.states)
	assert.Equal(t, m.transitions, newMachine.transitions)
	assert.Equal(t, m.currentState, newMachine.currentState)
	assert.Equal(t, m.id, newMachine.id)
	assert.Equal(t, m.guid, newMachine.guid)
}

func TestID(t *testing.T) {
	m := NewMachine(context.Background(), "test-machine", &MockStorage{t: t})
	assert.Equal(t, "test-machine", m.ID())
}

func TestCurrentOrVisited(t *testing.T) {
	m := NewMachine(context.Background(), "test-machine", &MockStorage{t: t})
	m.AddState("state1")
	m.AddState("state2")
	m.SetInitialState("state1")

	err := m.CurrentOrVisited("state1")
	assert.NoError(t, err)

	err = m.CurrentOrVisited("state2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Missing steps in flow")

	ctx := context.TODO()
	m.AddTransition("state1", "event1", "state2")
	_, _ = m.Trigger(ctx, "event1")
	err = m.CurrentOrVisited("state1")
	assert.NoError(t, err)
}

func TestMachine_VisitedOneOf(t *testing.T) {
	tests := []struct {
		name           string
		visitedStates  []State
		statesToCheck  []State
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:          "One state visited",
			visitedStates: []State{"A"},
			statesToCheck: []State{"A", "B", "C"},
			expectedError: false,
		},
		{
			name:          "Multiple states visited, one matches",
			visitedStates: []State{"A", "B"},
			statesToCheck: []State{"B", "C", "D"},
			expectedError: false,
		},
		{
			name:           "No states visited",
			visitedStates:  []State{},
			statesToCheck:  []State{"A", "B", "C"},
			expectedError:  true,
			expectedErrMsg: "no requested states were visited",
		},
		{
			name:           "States visited, but none match",
			visitedStates:  []State{"X", "Y", "Z"},
			statesToCheck:  []State{"A", "B", "C"},
			expectedError:  true,
			expectedErrMsg: "no requested states were visited",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Machine{
				visitedStates: []VisitedState{},
			}
			for _, state := range tt.visitedStates {
				m.visitedStates = append(m.visitedStates, VisitedState{State: state, At: time.Now()})
			}

			err := m.VisitedOneOf(tt.statesToCheck...)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)

				perr, ok := perrors.Has(err, perrors.InvalidFormDataCode)
				assert.True(t, ok)
				assert.Len(t, perr.FieldErrors, 1)
				assert.Equal(t, "state", perr.FieldErrors[0].Field)
				assert.Equal(t, "stateNotVisited", perr.FieldErrors[0].Rule)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMachine_VisitedAll(t *testing.T) {
	tests := []struct {
		name           string
		visitedStates  []State
		statesToCheck  []State
		expectedError  bool
		expectedErrMsg string
		expectedFields int
	}{
		{
			name:          "All states visited",
			visitedStates: []State{"A", "B", "C"},
			statesToCheck: []State{"A", "B", "C"},
			expectedError: false,
		},
		{
			name:           "Some states not visited",
			visitedStates:  []State{"A", "B"},
			statesToCheck:  []State{"A", "B", "C", "D"},
			expectedError:  true,
			expectedErrMsg: "some states are not visited",
			expectedFields: 2,
		},
		{
			name:           "No states visited",
			visitedStates:  []State{},
			statesToCheck:  []State{"A", "B", "C"},
			expectedError:  true,
			expectedErrMsg: "some states are not visited",
			expectedFields: 3,
		},
		{
			name:          "More states visited than checked",
			visitedStates: []State{"A", "B", "C", "D", "E"},
			statesToCheck: []State{"B", "D"},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Machine{
				visitedStates: []VisitedState{},
			}
			for _, state := range tt.visitedStates {
				m.visitedStates = append(m.visitedStates, VisitedState{State: state, At: time.Now()})
			}

			err := m.VisitedAll(tt.statesToCheck...)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)

				perr, ok := perrors.Has(err, perrors.StateMachineStateNotVisitedErrorCode)
				assert.True(t, ok)
				assert.Len(t, perr.FieldErrors, tt.expectedFields)
				for _, field := range perr.FieldErrors {
					assert.Equal(t, "state", field.Field)
					assert.Equal(t, "stateNotVisited", field.Rule)
					assert.Contains(t, field.Message, "is not visited")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTriggerWithConditionalTransitions(t *testing.T) {
	m := NewMachine(context.Background(), "test-machine", &MockStorage{t: t})
	m.AddState("state1")
	m.AddState("state2")
	m.AddState("state3")

	err := m.SetInitialState("state1")
	assert.NoError(t, err)

	// Add a regular transition
	err = m.AddTransition("state1", "regular-event", "state2")
	assert.NoError(t, err)

	// Add conditional transitions
	err = m.AddConditionalTransition("state1", "cond-event", "state3", ConditionNeverVisitedOneOf, "state2")
	assert.NoError(t, err)

	err = m.AddConditionalTransition("state2", "cond-event", "state3", ConditionIfVisitedOneOf, "state1")
	assert.NoError(t, err)

	// Test regular transition first
	ctx := context.TODO()
	newState, err := m.Trigger(ctx, "regular-event")
	assert.NoError(t, err)
	assert.Equal(t, State("state2"), newState)
	assert.Equal(t, State("state2"), m.CurrentState())

	// Test condition that should be met (we've visited state1)
	newState, err = m.Trigger(ctx, "cond-event")
	assert.NoError(t, err)
	assert.Equal(t, State("state3"), newState)
	assert.Equal(t, State("state3"), m.CurrentState())

	// Reset for testing ConditionNeverVisitedOneOf
	m = NewMachine(context.Background(), "test-machine2", &MockStorage{t: t})
	m.AddState("state1")
	m.AddState("state2")
	m.AddState("state3")

	err = m.SetInitialState("state1")
	assert.NoError(t, err)

	err = m.AddConditionalTransition("state1", "never-visited-event", "state3", ConditionNeverVisitedOneOf, "state2")
	assert.NoError(t, err)

	// This should work since we've never visited state2
	newState, err = m.Trigger(ctx, "never-visited-event")
	assert.NoError(t, err)
	assert.Equal(t, State("state3"), newState)

	// Test ConditionIfVisitedAll
	m = NewMachine(context.Background(), "test-machine3", &MockStorage{t: t})
	m.AddState("start")
	m.AddState("step1")
	m.AddState("step2")
	m.AddState("final")

	err = m.SetInitialState("start")
	assert.NoError(t, err)

	err = m.AddTransition("start", "go-step1", "step1")
	assert.NoError(t, err)

	err = m.AddTransition("step1", "go-step2", "step2")
	assert.NoError(t, err)

	err = m.AddConditionalTransition("step2", "go-final", "final", ConditionIfVisitedAll, "start", "step1")
	assert.NoError(t, err)

	// Navigate through the machine
	_, err = m.Trigger(ctx, "go-step1")
	assert.NoError(t, err)

	_, err = m.Trigger(ctx, "go-step2")
	assert.NoError(t, err)

	// This should work because we've visited all required states
	newState, err = m.Trigger(ctx, "go-final")
	assert.NoError(t, err)
	assert.Equal(t, State("final"), newState)

	// Test invalid event
	_, err = m.Trigger(ctx, "non-existent-event")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid transition")
}

func TestTriggerWithConditionalPriorityOverRegular(t *testing.T) {
	m := NewMachine(context.Background(), "test-machine", &MockStorage{t: t})
	m.AddState("state1")
	m.AddState("state2")
	m.AddState("state3")

	err := m.SetInitialState("state1")
	assert.NoError(t, err)

	err = m.AddTransition("state1", "dual-event", "state2")
	assert.NoError(t, err)

	err = m.AddConditionalTransition("state1", "dual-event", "state3", ConditionNeverVisitedOneOf, "state2")
	assert.NoError(t, err)

	ctx := context.TODO()
	newState, err := m.Trigger(ctx, "dual-event")
	assert.NoError(t, err)
	assert.Equal(t, State("state3"), newState, "Conditional transition should take priority when condition is met")

	m = NewMachine(context.Background(), "test-machine2", &MockStorage{t: t})
	m.AddState("state1")
	m.AddState("state2")
	m.AddState("state3")

	err = m.SetInitialState("state1")
	assert.NoError(t, err)

	// Mark state2 as visited
	m.visitedStates = append(m.visitedStates, VisitedState{State: "state2", At: time.Now()})

	// Add both transitions again
	err = m.AddTransition("state1", "dual-event", "state2")
	assert.NoError(t, err)

	err = m.AddConditionalTransition("state1", "dual-event", "state3", ConditionNeverVisitedOneOf, "state2")
	assert.NoError(t, err)

	// Now condition isn't met, so regular transition should be used
	newState, err = m.Trigger(ctx, "dual-event")
	assert.NoError(t, err)
	assert.Equal(t, State("state2"), newState, "Regular transition should be used when condition is not met")
}

func TestAddConditionalTransition(t *testing.T) {
	m := NewMachine(context.Background(), "test-machine", &MockStorage{t: t})
	m.AddState("state1")
	m.AddState("state2")
	m.AddState("state3")

	// Test valid conditional transition with ConditionNeverVisitedOneOf
	err := m.AddConditionalTransition("state1", "event1", "state2", ConditionNeverVisitedOneOf, "state3")
	assert.NoError(t, err)
	conditions, exists := m.conditionalsTransitions["state1"]["event1"]
	assert.True(t, exists)
	assert.Len(t, conditions, 1)
	assert.Equal(t, State("state2"), conditions[0].To)
	assert.Equal(t, ConditionNeverVisitedOneOf, conditions[0].If)
	assert.ElementsMatch(t, []State{"state3"}, conditions[0].IfStates)

	// Test adding a second condition to the same from-state and event
	err = m.AddConditionalTransition("state1", "event1", "state3", ConditionIfVisitedOneOf, "state2")
	assert.NoError(t, err)
	conditions, exists = m.conditionalsTransitions["state1"]["event1"]
	assert.True(t, exists)
	assert.Len(t, conditions, 2)
	assert.Equal(t, State("state3"), conditions[1].To)
	assert.Equal(t, ConditionIfVisitedOneOf, conditions[1].If)
	assert.ElementsMatch(t, []State{"state2"}, conditions[1].IfStates)

	// Test non-existent 'from' state
	err = m.AddConditionalTransition("non-existent", "event1", "state2", ConditionNeverVisitedAll, "state3")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'from' state does not exist")

	// Test non-existent 'to' state
	err = m.AddConditionalTransition("state1", "event3", "non-existent", ConditionNeverVisitedAll, "state3")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'to' state does not exist")

	// Test non-existent condition state
	err = m.AddConditionalTransition("state1", "event4", "state2", ConditionNeverVisitedAll, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid condition states")

	// Test with empty conditional states
	err = m.AddConditionalTransition("state1", "event5", "state2", ConditionNeverVisitedAll)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'conditionStates' cannot be empty")
}

func TestTriggerWithMultipleConditionalTransitions(t *testing.T) {
	m := NewMachine(context.Background(), "test-machine", &MockStorage{t: t})
	m.AddState("start")
	m.AddState("path_a")
	m.AddState("path_b")
	m.AddState("path_c")
	m.AddState("finish")

	err := m.SetInitialState("start")
	assert.NoError(t, err)

	// Add a regular transition as fallback
	err = m.AddTransition("start", "choose_path", "path_c")
	assert.NoError(t, err)

	// Add multiple conditional transitions for the same event
	// First condition - if user has never visited path_a, go to path_a
	err = m.AddConditionalTransition("start", "choose_path", "path_a", ConditionNeverVisitedOneOf, "path_a")
	assert.NoError(t, err)

	// Second condition - if user has already visited path_a but not path_b, go to path_b
	err = m.AddConditionalTransition("start", "choose_path", "path_b", ConditionIfVisitedOneOf, "path_a")
	assert.NoError(t, err)
	err = m.AddConditionalTransition("start", "choose_path", "path_b", ConditionNeverVisitedOneOf, "path_b")
	assert.NoError(t, err)
	ctx := context.TODO()
	// Test first condition (should go to path_a since we've never visited it)
	newState, err := m.Trigger(ctx, "choose_path")
	assert.NoError(t, err)
	assert.Equal(t, State("path_a"), newState)
	assert.Equal(t, State("path_a"), m.CurrentState())

	// Go back to start for next test
	err = m.AddTransition("path_a", "back", "start")
	assert.NoError(t, err)
	_, err = m.Trigger(ctx, "back")
	assert.NoError(t, err)

	// Test second condition (should go to path_b since we've visited path_a but not path_b)
	newState, err = m.Trigger(ctx, "choose_path")
	assert.NoError(t, err)
	assert.Equal(t, State("path_b"), newState)
	assert.Equal(t, State("path_b"), m.CurrentState())

	// Go back to start for next test
	err = m.AddTransition("path_b", "back", "start")
	assert.NoError(t, err)
	_, err = m.Trigger(ctx, "back")
	assert.NoError(t, err)

	// Test fallback to regular transition (we've visited both path_a and path_b, so no condition matches)
	newState, err = m.Trigger(ctx, "choose_path")
	assert.NoError(t, err)
	assert.Equal(t, State("path_c"), newState)
	assert.Equal(t, State("path_c"), m.CurrentState())
}

func TestConditionalTransitionEvaluationOrder(t *testing.T) {
	m := NewMachine(context.Background(), "test-machine", &MockStorage{t: t})
	m.AddState("start")
	m.AddState("route1")
	m.AddState("route2")
	m.AddState("route3")

	err := m.SetInitialState("start")
	assert.NoError(t, err)

	// Regular transition as last resort
	err = m.AddTransition("start", "go", "route3")
	assert.NoError(t, err)

	// Add conditional transitions in reverse priority order to test that order of addition doesn't matter
	err = m.AddConditionalTransition("start", "go", "route1", ConditionNeverVisitedOneOf, "route1", "route2")
	assert.NoError(t, err)

	err = m.AddConditionalTransition("start", "go", "route2", ConditionNeverVisitedOneOf, "route2")
	assert.NoError(t, err)

	ctx := context.TODO()
	// The most specific condition (route2) should be chosen first even though it was added second
	newState, err := m.Trigger(ctx, "go")
	assert.NoError(t, err)
	assert.Equal(t, State("route2"), newState, "The most specific condition should be evaluated first")
}

func TestConditionWithIfVisitedAll(t *testing.T) {
	m := NewMachine(context.Background(), "test-machine", &MockStorage{t: t})
	m.AddState("step1")
	m.AddState("step2")
	m.AddState("step3")
	m.AddState("completion")

	err := m.SetInitialState("step1")
	assert.NoError(t, err)

	// Add transitions to navigate through steps
	err = m.AddTransition("step1", "next", "step2")
	assert.NoError(t, err)

	err = m.AddTransition("step2", "next", "step3")
	assert.NoError(t, err)

	// From step3, you can go to completion only if you've visited all previous steps
	err = m.AddConditionalTransition("step3", "complete", "completion", ConditionIfVisitedAll, "step1", "step2")
	assert.NoError(t, err)

	ctx := context.TODO()
	// Navigate through the steps
	_, err = m.Trigger(ctx, "next") // step1 -> step2
	assert.NoError(t, err)

	_, err = m.Trigger(ctx, "next") // step2 -> step3
	assert.NoError(t, err)

	// Now try to complete - should work because we've visited all required steps
	newState, err := m.Trigger(ctx, "complete")
	assert.NoError(t, err)
	assert.Equal(t, State("completion"), newState)

	// Test the negative case - create a new machine and try to skip steps
	m2 := NewMachine(context.Background(), "test-machine2", &MockStorage{t: t})
	m2.AddState("step1")
	m2.AddState("step2")
	m2.AddState("step3")
	m2.AddState("completion")

	err = m2.SetInitialState("step3") // Start directly at step3, skipping step1 and step2
	assert.NoError(t, err)

	err = m2.AddConditionalTransition("step3", "complete", "completion", ConditionIfVisitedAll, "step1", "step2")
	assert.NoError(t, err)

	// Fallback transition for when condition isn't met
	err = m2.AddTransition("step3", "complete", "step1")
	assert.NoError(t, err)

	// Should go to fallback since we haven't visited all required states
	newState, err = m2.Trigger(ctx, "complete")
	assert.NoError(t, err)
	assert.Equal(t, State("step1"), newState, "Should fall back to regular transition when condition isn't met")
}

func TestConditionWithNeverVisitedAll(t *testing.T) {
	m := NewMachine(context.Background(), "test-machine", &MockStorage{t: t})
	m.AddState("lobby")
	m.AddState("tutorial")
	m.AddState("advanced")
	m.AddState("expert")

	err := m.SetInitialState("lobby")
	assert.NoError(t, err)

	// Can go to tutorial only if never visited any advanced or expert content
	err = m.AddConditionalTransition("lobby", "start", "tutorial", ConditionNeverVisitedAll, "advanced", "expert")
	assert.NoError(t, err)

	// Fallback to advanced if condition not met
	err = m.AddTransition("lobby", "start", "advanced")
	assert.NoError(t, err)

	ctx := context.TODO()
	// First time should go to tutorial
	newState, err := m.Trigger(ctx, "start")
	assert.NoError(t, err)
	assert.Equal(t, State("tutorial"), newState)

	// Add a way back to lobby
	err = m.AddTransition("tutorial", "back", "lobby")
	assert.NoError(t, err)

	_, err = m.Trigger(ctx, "back")
	assert.NoError(t, err)

	// Mark as visited advanced without actually going there
	now := time.Now()
	m.visitedStates = append(m.visitedStates, VisitedState{State: "advanced", At: now})

	// Now condition should not be met, should fall back to regular transition
	newState, err = m.Trigger(ctx, "start")
	assert.NoError(t, err)
	assert.Equal(t, State("advanced"), newState, "Should use regular transition when NeverVisitedAll condition is not met")
}

func TestMultipleConditionalTransitionsWithSaveRestore(t *testing.T) {
	storage := &MockStorage{t: t}
	m := NewMachine(context.Background(), "test-machine", storage)
	m.AddState("start")
	m.AddState("middle")
	m.AddState("end")

	ctx := context.TODO()
	err := m.SetInitialState("start")
	assert.NoError(t, err)

	// Add conditional transitions
	err = m.AddConditionalTransition("start", "proceed", "middle", ConditionNeverVisitedOneOf, "middle")
	assert.NoError(t, err)

	err = m.AddConditionalTransition("middle", "proceed", "end", ConditionIfVisitedOneOf, "start")
	assert.NoError(t, err)

	// Save the machine
	err = m.Save()
	assert.NoError(t, err)

	// Restore to a new machine instance
	newMachine := NewMachine(context.Background(), "test-machine", storage)
	err = newMachine.Restore()
	assert.NoError(t, err)

	// Verify all conditional transitions were properly saved and restored
	assert.Equal(t, m.conditionalsTransitions, newMachine.conditionalsTransitions)

	// Test that transitions still work after restore
	newState, err := newMachine.Trigger(ctx, "proceed")
	assert.NoError(t, err)
	assert.Equal(t, State("middle"), newState)

	newState, err = newMachine.Trigger(ctx, "proceed")
	assert.NoError(t, err)
	assert.Equal(t, State("end"), newState)
}
