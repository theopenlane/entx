package enthistory

import (
	"context"
	"fmt"

	"entgo.io/ent"
)

// Mutation is an interface that must be implemented by all mutations that are
type Mutation interface {
	Op() ent.Op
	CreateHistoryFromCreate(ctx context.Context) error
	CreateHistoryFromUpdate(ctx context.Context) error
	CreateHistoryFromDelete(ctx context.Context) error
}

// Mutator is an interface that must be implemented by all mutators that are
type Mutator interface {
	Mutate(context.Context, Mutation) (ent.Value, error)
}

// On is a helper function that allows you to create a hook that only runs on a specific operation
func On(hk ent.Hook, op ent.Op) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if m.Op().Is(op) {
				return hk(next).Mutate(ctx, m)
			}

			return next.Mutate(ctx, m)
		})
	}
}

// HistoryHooks returns a list of hooks that can be used to create history entries
func HistoryHooks[T Mutation]() []ent.Hook {
	return []ent.Hook{
		On(historyHookCreate[T](), ent.OpCreate),
		On(historyHookUpdate[T](), ent.OpUpdate|ent.OpUpdateOne),
		On(historyHookDelete[T](), ent.OpDelete|ent.OpDeleteOne),
	}
}

// getTypedMutation is a helper function that allows you to get a typed mutation from an ent.Mutation
func getTypedMutation[T Mutation](m ent.Mutation) (T, error) {
	f, ok := any(m).(T)
	if !ok {
		return f, fmt.Errorf("expected appropriately typed mutation in schema hook, got: %+v", m) //nolint:err113
	}

	return f, nil
}

// historyHookCreate is a hook that creates a history entry when a create operation is performed
func historyHookCreate[T Mutation]() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			mutation, err := getTypedMutation[T](m)
			if err != nil {
				return nil, err
			}

			value, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			err = mutation.CreateHistoryFromCreate(ctx)
			if err != nil {
				return nil, err
			}

			return value, nil
		})
	}
}

// historyHookUpdate is a hook that creates a history entry when an update operation is performed
func historyHookUpdate[T Mutation]() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			mutation, err := getTypedMutation[T](m)
			if err != nil {
				return nil, err
			}

			if err = mutation.CreateHistoryFromUpdate(ctx); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}
}

// historyHookDelete is a hook that creates a history entry when a delete operation is performed
func historyHookDelete[T Mutation]() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			mutation, err := getTypedMutation[T](m)
			if err != nil {
				return nil, err
			}

			if err = mutation.CreateHistoryFromDelete(ctx); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}
}
