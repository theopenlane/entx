// Package gala is a minimal in-memory pub/sub helper standing in for the equivalent core-more
// package (github.com/theopenlane/core/pkg/gala), providing just the surface entityops-generated
// code calls: typed topics, listener registration, and synchronous emit dispatch
package gala

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/samber/do/v2"
)

// TopicName identifies a gala topic
type TopicName string

// Topic is a typed handle to a named topic carrying payloads of type T
type Topic[T any] struct {
	Name TopicName
}

// OperationContext carries the provenance of the mutation that produced an event
type OperationContext struct {
	OwnerID    string `json:"ownerId,omitempty"`
	Operation  string `json:"operation,omitempty"`
	EntityID   string `json:"entityId,omitempty"`
	EntityType string `json:"entityType,omitempty"`
}

// Headers carries event metadata alongside its payload
type Headers struct {
	Metadata json.RawMessage
}

// HandlerContext is passed to a registered listener when its topic receives an event
type HandlerContext struct {
	Context  context.Context
	Injector do.Injector
}

// Definition describes one listener: the topic it subscribes to and its handle function
type Definition[T any] struct {
	Topic      Topic[T]
	Name       string
	Operations []string
	Handle     func(HandlerContext, T) error
}

// ListenerID identifies a registered listener
type ListenerID string

// Receipt reports the outcome of an EmitWithHeaders call
type Receipt struct {
	Err error
}

type registeredHandler func(ctx context.Context, injector do.Injector, event any) error

// Registry holds registered listeners, keyed by topic name
type Registry struct {
	handlers map[TopicName][]registeredHandler
}

// NewRegistry creates an empty listener registry
func NewRegistry() *Registry {
	return &Registry{handlers: map[TopicName][]registeredHandler{}}
}

// Gala is the runtime that dispatches emitted events to registered listeners
type Gala struct {
	registry *Registry
	injector do.Injector
}

// New creates a Gala runtime backed by registry, resolving handler dependencies from injector
func New(registry *Registry, injector do.Injector) *Gala {
	if registry == nil {
		registry = NewRegistry()
	}

	return &Gala{registry: registry, injector: injector}
}

// Registry returns the runtime's listener registry
func (g *Gala) Registry() *Registry {
	return g.registry
}

// RegisterListeners registers def's handler on registry, returning its listener id
func RegisterListeners[T any](registry *Registry, def Definition[T]) ([]ListenerID, error) {
	if registry == nil {
		return nil, fmt.Errorf("gala: nil registry")
	}

	registry.handlers[def.Topic.Name] = append(registry.handlers[def.Topic.Name], func(ctx context.Context, injector do.Injector, event any) error {
		typed, ok := event.(T)
		if !ok {
			return fmt.Errorf("gala: event for topic %q is not of the expected type", def.Topic.Name)
		}

		return def.Handle(HandlerContext{Context: ctx, Injector: injector}, typed)
	})

	return []ListenerID{ListenerID(def.Name)}, nil
}

// EmitWithHeaders synchronously dispatches event to every listener registered for topic
func (g *Gala) EmitWithHeaders(ctx context.Context, topic TopicName, event any, _ Headers) Receipt {
	for _, handle := range g.registry.handlers[topic] {
		if err := handle(ctx, g.injector, event); err != nil {
			return Receipt{Err: err}
		}
	}

	return Receipt{}
}

type operationContextKey struct{}

// WithOperationContext attaches oc to ctx, retrievable later via OperationContextFromContext
func WithOperationContext(ctx context.Context, oc OperationContext) context.Context {
	return context.WithValue(ctx, operationContextKey{}, oc)
}

// OperationContextFromContext returns the OperationContext attached to ctx, if any
func OperationContextFromContext(ctx context.Context) (OperationContext, bool) {
	oc, ok := ctx.Value(operationContextKey{}).(OperationContext)

	return oc, ok
}
