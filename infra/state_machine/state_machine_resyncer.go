package state_machine

import (
	"context"

	pixieCtx "github.com/pixie-sh/core-go/pkg/context"
)

// ResyncTransitionsAndStates DO NOT USE unless you know what is going on
// this will screw up any provided stateMachine to the transitions and states you are setting it to
func ResyncTransitionsAndStates(ctx context.Context, stateMachine *Machine, transitions []Transition, states []State) {
	log := pixieCtx.GetCtxLogger(ctx)
	if len(transitions) == 0 || len(states) == 0 {
		log.Warn("cant resync state machine without transitions or states")
		return
	}

	log.Clone().
		With("before_states", stateMachine.states).
		With("before_transitions", stateMachine.transitions).
		Log("before resync states and transitions")

	stateMachine.states = make(map[State]struct{})
	for _, state := range states {
		stateMachine.AddState(state)
	}

	stateMachine.transitions = make(Transitions)
	for _, transition := range transitions {
		_ = stateMachine.AddTransition(transition.From, transition.Event, transition.To)
	}

	log.Clone().
		With("after_states", stateMachine.states).
		With("after_transitions", stateMachine.transitions).
		Log("after resync states and transitions")

	return
}
