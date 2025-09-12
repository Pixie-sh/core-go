package state_machine

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pixie-sh/errors-go"

	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
	"github.com/pixie-sh/core-go/pkg/types"
	"github.com/pixie-sh/core-go/pkg/types/maps"
	"github.com/pixie-sh/core-go/pkg/types/slices"
	"github.com/pixie-sh/core-go/pkg/uid"
)

var NilState State = ""

type State string

func (s State) String() string {
	return string(s)
}

type Event string

func (e Event) String() string {
	return string(e)
}

type Transitions = map[State]map[Event]State

type Transition struct {
	From  State
	Event Event
	To    State
}

type ConditionEnum = string

const (
	ConditionNeverVisitedOneOf ConditionEnum = "never_visited_one_of"
	ConditionIfVisitedOneOf    ConditionEnum = "if_visited_one_of"
	ConditionNeverVisitedAll   ConditionEnum = "never_visited_all"
	ConditionIfVisitedAll      ConditionEnum = "if_visited_all"
)

type ConditionalTransitions = map[State]map[Event][]Condition
type Condition struct {
	To       State         `json:"to"`
	If       ConditionEnum `json:"if"`
	IfStates []State       `json:"if_states"`
}

type VisitedState struct {
	State State     `json:"state"`
	At    time.Time `json:"when"`
}

type MachineModel struct {
	Guid                   string                 `json:"guid"`
	ID                     string                 `json:"id"`
	States                 []State                `json:"states"`
	VisitedStates          []VisitedState         `json:"visited_states"`
	Transitions            Transitions            `json:"transitions"`
	ConditionalTransitions ConditionalTransitions `json:"conditional_transitions"`
	CurrentState           State                  `json:"current_state"`
	CurrentStateAt         *time.Time             `json:"current_state_at"`
}

type CheckerMachine interface {
	ID() string
	CurrentOrVisited(state State) error
	VisitedOneOf(state ...State) error
	VisitedAll(state ...State) error
	Visited() []VisitedState
	CurrentState() State
}

type TriggerMachine interface {
	CheckerMachine

	Trigger(ctx context.Context, event Event) (State, error)
	Save() error
}

type Machine struct {
	guid    string
	id      string
	storage StateMachineStorage

	mu                      sync.Mutex
	states                  map[State]struct{}
	visitedStates           []VisitedState
	transitions             Transitions
	conditionalsTransitions ConditionalTransitions
	currentState            State
	currentStateAt          time.Time
	entityStorage           StateMachineEntityStorage
}

type StateMachineStorage interface {
	Store(model MachineModel) error
	Get(machineID string) (MachineModel, error)
}

type StateMachineEntityStorage interface {
	StoreCurrentState(state State) error
}

type nullEntityStorage struct{}

func (n nullEntityStorage) StoreCurrentState(State) error { return nil }

func NewMachine(_ context.Context, machineID string, storage StateMachineStorage, withStateMachineEntityStorage ...StateMachineEntityStorage) *Machine {
	var entityStorage StateMachineEntityStorage = nullEntityStorage{}
	if len(withStateMachineEntityStorage) > 0 {
		entityStorage = withStateMachineEntityStorage[0]
	}

	m := &Machine{
		guid:                    uid.NewUUID(),
		id:                      machineID,
		storage:                 storage,
		states:                  make(map[State]struct{}),
		visitedStates:           []VisitedState{},
		transitions:             make(Transitions),
		conditionalsTransitions: make(ConditionalTransitions),
		entityStorage:           entityStorage,
	}

	return m
}

func (m *Machine) AddState(state State) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states[state] = struct{}{}
}

func (m *Machine) SetInitialState(state State) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.states[state]; !exists {
		return errors.New("state does not exist")
	}
	m.currentState = state
	m.currentStateAt = time.Now().UTC()
	return nil
}

// CurrentOrVisited has the user been in this step or is he currently?
func (m Machine) CurrentOrVisited(state State) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentState == state {
		return nil
	}

	err := m.visitedOneOfUnlocked(state)
	if err != nil {
		return errors.NewValidationError("Missing steps in flow", &errors.FieldError{
			Field:   "state",
			Rule:    "stateNotVisited",
			Param:   state.String(),
			Message: fmt.Sprintf("Please complete first all steps before %s", state),
		}).WithErrorCode(errors.StateMachineStateNotVisitedErrorCode)
	}

	return nil
}

func (m Machine) visitedOneOfUnlocked(state ...State) error {
	for _, visited := range state {
		contains := slices.Find(m.visitedStates, func(item VisitedState) bool {
			return item.State == visited
		})

		if !types.IsEmpty(contains) {
			return nil
		}
	}

	return errors.NewValidationError("no requested states were visited", &errors.FieldError{
		Field: "state",
		Rule:  "stateNotVisited",
		Param: strings.Join(slices.Map(state, func(item State) string {
			return item.String()
		}), ";"),
		Message: fmt.Sprintf("no state were visited"),
	})
}

// VisitedOneOf return error if none of provided states were visited;
// return nil if at least one state of requested one was visited
func (m Machine) VisitedOneOf(state ...State) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.visitedOneOfUnlocked(state...)
}

// VisitedAll return error if at least is not visited
// return nil if all were visited
func (m Machine) VisitedAll(state ...State) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, err := m.visitedAllUnlocked(state...)
	return err
}

func (m Machine) visitedAllUnlocked(state ...State) ([]State, error) {
	var notVisited []State
	for _, stateToLookup := range state {
		contains := slices.Find(m.visitedStates, func(item VisitedState) bool {
			return item.State == stateToLookup
		})

		if types.IsEmpty(contains) {
			notVisited = append(notVisited, stateToLookup)
		}
	}

	if len(notVisited) > 0 {
		var fields []*errors.FieldError
		for _, not := range notVisited {
			fields = append(fields, &errors.FieldError{
				Field:   "state",
				Param:   not.String(),
				Rule:    "stateNotVisited",
				Message: fmt.Sprintf("state '%s' is not visited", not),
			})
		}

		return notVisited, errors.NewValidationError("some states are not visited", fields...).WithErrorCode(errors.StateMachineStateNotVisitedErrorCode)
	}

	return nil, nil
}

// AddTransition adds a transition between two states in the state machine.
// It associates a specific event with a transition from a 'from' state to a 'to' state.
//
// Parameters:
// - from: The state from which the transition originates. This must already exist in the machine.
// - event: The event that triggers the transition.
// - to: The state to transition to when the event occurs. This must also exist in the machine.
//
// Returns:
// - An error if either the 'from' or 'to' state does not exist in the state machine.
//
// If the 'from' state has no transitions defined yet, this function initializes the map
// for that state. Then, it creates or updates the mapping of the given event to the 'to' state.
func (m *Machine) AddTransition(from State, event Event, to State) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.states[from]; !exists {
		return errors.New("'from' state does not exist").WithErrorCode(errors.StateMachineInvalidStateErrorCode)
	}
	if _, exists := m.states[to]; !exists {
		return errors.New("'to' state does not exist").WithErrorCode(errors.StateMachineInvalidStateErrorCode)
	}

	if m.transitions[from] == nil {
		m.transitions[from] = make(map[Event]State)
	}
	m.transitions[from][event] = to
	return nil
}

// AddConditionalTransition adds a conditional transition to the state machine.
// This allows moving from one state to another based on specific conditions.
// The condition defines whether one or more states must have been visited or not
// for the transition to be valid. If the condition is met, the state machine
// transitions to the specified target state.
//
// Parameters:
// - from: The state from which the transition originates.
// - event: The event that triggers the transition.
// - to: The state to transition to if the condition is met.
// - condition: The type of condition to check (e.g., ConditionNeverVisitedOneOf, ConditionIfVisitedAll).
// - conditionStates: The list of states involved in the condition.
//
// Returns:
// - An error if there is an issue with the configuration of the transition, such as:
//   - The 'from' or 'to' state does not exist.
//   - The condition is invalid.
//   - The 'conditionStates' contains a state that does not exist.
//   - No 'conditionStates' are provided.
func (m *Machine) AddConditionalTransition(
	from State,
	event Event,
	to State,
	condition ConditionEnum,
	conditionStates ...State,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.states[from]; !exists {
		return errors.New("'from' state does not exist").WithErrorCode(errors.StateMachineInvalidStateErrorCode)
	}
	if _, exists := m.states[to]; !exists {
		return errors.New("'to' state does not exist").WithErrorCode(errors.StateMachineInvalidStateErrorCode)
	}

	switch condition {
	case ConditionNeverVisitedOneOf:
	case ConditionIfVisitedOneOf:
	case ConditionNeverVisitedAll:
	case ConditionIfVisitedAll:
		break
	default:
		return errors.New("invalid condition").WithErrorCode(errors.StateMachineInvalidStateErrorCode)
	}

	if len(conditionStates) == 0 {
		return errors.New("'conditionStates' cannot be empty").WithErrorCode(errors.StateMachineInvalidStateErrorCode)
	}

	var fieldErrors []*errors.FieldError
	for ci, cstate := range conditionStates {
		if _, exists := m.states[cstate]; !exists {
			fieldErrors = append(fieldErrors, &errors.FieldError{
				Field:   fmt.Sprintf("conditionStates[%d]", ci),
				Rule:    "StateMustExist",
				Param:   cstate.String(),
				Message: "state must exist",
			})
		}
	}
	if len(fieldErrors) > 0 {
		return errors.NewValidationError("invalid condition states", fieldErrors...).WithErrorCode(errors.StateMachineInvalidStateErrorCode)
	}

	if m.conditionalsTransitions[from] == nil {
		m.conditionalsTransitions[from] = make(map[Event][]Condition)
	}

	toSort := append(m.conditionalsTransitions[from][event], Condition{
		To:       to,
		If:       condition,
		IfStates: conditionStates,
	})
	sort.SliceStable(toSort, func(i, j int) bool {
		return len(toSort[i].IfStates) < len(toSort[j].IfStates)
	})

	m.conditionalsTransitions[from][event] = toSort
	return nil
}

// Trigger handles the transition of the state machine by processing the given event.
// If there are conditional transitions defined for the current state and event,
// it evaluates the conditions to determine the validity of the transition.
// If no valid conditional transition is found, it checks for a direct transition.
// It returns the next state if the transition is successful or an error if the
// event is invalid or the transition conditions are not met.
func (m *Machine) Trigger(ctx context.Context, event Event) (State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errFields = make(map[State][]*errors.FieldError)
	nextStateConditions, existsConditional := m.conditionalsTransitions[m.currentState][event]
	if existsConditional {
		for _, nextStateCondition := range nextStateConditions {
			fieldErr, err := m.evaluateCondition(event, nextStateCondition.If, nextStateCondition.IfStates)
			if err != nil {
				return NilState, err
			}

			if fieldErr != nil {
				_, ok := errFields[nextStateCondition.To]
				if !ok {
					errFields[nextStateCondition.To] = []*errors.FieldError{}
				}

				errFields[nextStateCondition.To] = append(errFields[nextStateCondition.To], fieldErr)
				continue
			}
		}

		for _, nextStateCondition := range nextStateConditions {
			_, withErr := errFields[nextStateCondition.To]
			if !withErr {
				return m.affectStateWithNextStateUnlocked(nextStateCondition.To, time.Now().UTC())
			}
		}
	}

	if len(errFields) > 0 {
		pixiecontext.GetCtxLogger(ctx).With("event", event).
			With("state", m.currentState).
			With("errFields", errFields).
			Debug("invalid contional transition")
	}

	nextState, exists := m.transitions[m.currentState][event]
	if !exists {
		errList := maps.MapSliceValues(errFields)
		return NilState, errors.NewValidationError("Invalid transition", append(errList, &errors.FieldError{
			Field:   "event",
			Rule:    "InvalidTransition",
			Param:   event.String(),
			Message: fmt.Sprintf("invalid transition from '%s' with '%s'", m.currentState, event),
		})...).WithErrorCode(errors.StateMachineInvalidTransitionErrorCode)
	}

	return m.affectStateWithNextStateUnlocked(nextState, time.Now().UTC())
}

func (m Machine) CurrentState() State {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.currentState
}

func (m *Machine) Save() error {
	m.mu.Lock()
	model := MachineModel{
		Guid:                   m.guid,
		ID:                     m.id,
		States:                 maps.MapKeys(m.states),
		VisitedStates:          m.visitedStates,
		Transitions:            m.transitions,
		ConditionalTransitions: m.conditionalsTransitions,
		CurrentState:           m.currentState,
		CurrentStateAt:         &m.currentStateAt,
	}
	m.mu.Unlock()

	err := m.entityStorage.StoreCurrentState(m.currentState)
	if err != nil {
		return err
	}

	return m.storage.Store(model)
}

func (m *Machine) Restore() error {
	restored, err := m.storage.Get(m.id)
	if err != nil {
		return err
	}

	states := make(map[State]struct{}, len(restored.States))
	for _, state := range restored.States {
		states[state] = struct{}{}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.states = states
	m.visitedStates = restored.VisitedStates
	m.transitions = restored.Transitions
	m.conditionalsTransitions = restored.ConditionalTransitions
	m.currentState = restored.CurrentState
	m.currentStateAt = *restored.CurrentStateAt
	m.id = restored.ID
	m.guid = restored.Guid

	return nil
}

func (m Machine) ID() string {
	return m.id
}

func (m Machine) evaluateCondition(event Event, nextStateConditionIf ConditionEnum, ifStates []State) (*errors.FieldError, error) {
	switch nextStateConditionIf {
	case ConditionNeverVisitedOneOf:
		err := m.visitedOneOfUnlocked(ifStates...)
		if err == nil {
			return &errors.FieldError{
				Field: "event",
				Rule:  "InvalidTransition",
				Param: event.String(),
				Message: fmt.Sprintf(
					"invalid transition from '%s' with '%s', conditions '%s' are not met",
					m.currentState,
					event,
					nextStateConditionIf,
				),
			}, nil
		}
	case ConditionIfVisitedOneOf:
		err := m.visitedOneOfUnlocked(ifStates...)
		if err != nil {
			return &errors.FieldError{
				Field: "event",
				Rule:  "InvalidTransition",
				Param: event.String(),
				Message: fmt.Sprintf(
					"invalid transition from '%s' with '%s', conditions '%s' are not met",
					m.currentState,
					event,
					nextStateConditionIf,
				),
			}, nil
		}
	case ConditionNeverVisitedAll:
		notVisited, _ := m.visitedAllUnlocked(ifStates...)
		if len(notVisited) < len(ifStates) || len(notVisited) == 0 {
			return &errors.FieldError{
				Field: "event",
				Rule:  "InvalidTransition",
				Param: event.String(),
				Message: fmt.Sprintf(
					"invalid transition from '%s' with '%s', conditions '%s' are not met",
					m.currentState,
					event,
					nextStateConditionIf,
				),
			}, nil
		}
	case ConditionIfVisitedAll:
		_, err := m.visitedAllUnlocked(ifStates...)
		if err != nil {
			return &errors.FieldError{
				Field: "event",
				Rule:  "InvalidTransition",
				Param: event.String(),
				Message: fmt.Sprintf(
					"invalid transition from '%s' with '%s', conditions '%s' are not met",
					m.currentState,
					event,
					nextStateConditionIf,
				),
			}, nil
		}
	default:
		return nil, errors.
			New("invalid condition '%s'", nextStateConditionIf).
			WithErrorCode(errors.StateMachineInvalidTransitionErrorCode)
	}

	return nil, nil
}

func (m *Machine) affectStateWithNextStateUnlocked(nextState State, now time.Time) (State, error) {
	m.visitedStates = append(m.visitedStates, VisitedState{
		State: m.currentState,
		At:    now,
	})

	m.currentState = nextState
	m.currentStateAt = now
	return m.currentState, nil
}

func (m Machine) Visited() []VisitedState {
	return slices.Copy(m.visitedStates)
}
