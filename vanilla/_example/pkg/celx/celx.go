// Package celx is a minimal CEL evaluation helper standing in for the equivalent core-more
// package (github.com/theopenlane/core/pkg/celx), providing just the surface entityops-generated
// code calls: a boolean CEL evaluator bound to a native Go struct type via json tags
package celx

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
)

const (
	entityVarTarget = "target"
	entityVarSource = "source"
)

// EnvConfig configures the CEL environment
type EnvConfig struct {
	// CrossTypeNumericComparisons allows comparisons between different numeric CEL types
	CrossTypeNumericComparisons bool
}

// EvalConfig configures CEL program evaluation; empty today, reserved for parity with the
// core-more package's richer configuration surface
type EvalConfig struct{}

// StrictEnvConfig returns a conservative default EnvConfig
func StrictEnvConfig() EnvConfig {
	return EnvConfig{CrossTypeNumericComparisons: true}
}

// FastEvalConfig returns a default EvalConfig
func FastEvalConfig() EvalConfig {
	return EvalConfig{}
}

// NativeEntityEvaluator evaluates boolean CEL expressions against typed entity values, binding
// the candidate entity to "target" and an optional source entity to "source" as native CEL
// struct types, with fields accessed by their json tag
type NativeEntityEvaluator struct {
	env        *cel.Env
	targetType reflect.Type
	sourceType reflect.Type
}

// NewNativeEntityEvaluator builds a typed entity evaluator whose "target" (and optional
// "source") variables are the native CEL types of targetType and sourceType
func NewNativeEntityEvaluator(envCfg EnvConfig, _ EvalConfig, targetType reflect.Type, sourceType reflect.Type) (*NativeEntityEvaluator, error) {
	targetType = derefType(targetType)

	nativeArgs := []any{targetType, ext.ParseStructTag("json")}

	vars := []cel.EnvOption{
		cel.Variable(entityVarTarget, cel.ObjectType(objectTypeName(targetType))),
	}

	if sourceType != nil {
		sourceType = derefType(sourceType)

		if sourceType != targetType {
			nativeArgs = append(nativeArgs, sourceType)
		}

		vars = append(vars, cel.Variable(entityVarSource, cel.ObjectType(objectTypeName(sourceType))))
	}

	env, err := cel.NewEnv(append(vars,
		cel.CrossTypeNumericComparisons(envCfg.CrossTypeNumericComparisons),
		ext.NativeTypes(nativeArgs...),
	)...)
	if err != nil {
		return nil, fmt.Errorf("build cel env: %w", err)
	}

	return &NativeEntityEvaluator{env: env, targetType: targetType, sourceType: sourceType}, nil
}

// EvaluateBool evaluates the expression against the candidate entity JSON, exposed as "target"
func (n *NativeEntityEvaluator) EvaluateBool(ctx context.Context, expression string, data json.RawMessage) (bool, error) {
	target, err := decodeNative(data, n.targetType)
	if err != nil {
		return false, err
	}

	return n.evaluateBool(ctx, expression, map[string]any{entityVarTarget: target})
}

// EvaluateBoolWithSource evaluates the expression with both the target and source entities bound
func (n *NativeEntityEvaluator) EvaluateBoolWithSource(ctx context.Context, expression string, targetData, sourceData json.RawMessage) (bool, error) {
	target, err := decodeNative(targetData, n.targetType)
	if err != nil {
		return false, err
	}

	source, err := decodeNative(sourceData, n.sourceType)
	if err != nil {
		return false, err
	}

	return n.evaluateBool(ctx, expression, map[string]any{entityVarTarget: target, entityVarSource: source})
}

func (n *NativeEntityEvaluator) evaluateBool(ctx context.Context, expression string, vars map[string]any) (bool, error) {
	ast, iss := n.env.Compile(expression)
	if iss != nil && iss.Err() != nil {
		return false, iss.Err()
	}

	prg, err := n.env.Program(ast)
	if err != nil {
		return false, err
	}

	out, _, err := prg.ContextEval(ctx, vars)
	if err != nil {
		return false, err
	}

	result, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("expression %q did not evaluate to a boolean", expression)
	}

	return result, nil
}

func decodeNative(data json.RawMessage, t reflect.Type) (any, error) {
	value := reflect.New(t).Interface()
	if err := json.Unmarshal(data, value); err != nil {
		return nil, err
	}

	return value, nil
}

func derefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return t
}

func objectTypeName(t reflect.Type) string {
	return derefType(t).String()
}
