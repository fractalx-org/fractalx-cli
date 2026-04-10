package validate

import (
	"fmt"
	"strings"

	"github.com/fractalx-org/fractalx-cli/internal/model"
)

// Error holds validation errors and warnings.
type Result struct {
	Errors   []string
	Warnings []string
}

func (r *Result) HasErrors() bool { return len(r.Errors) > 0 }

func (r *Result) Print() {
	for _, w := range r.Warnings {
		fmt.Printf("  \033[33m⚠  %s\033[0m\n", w)
	}
	for _, e := range r.Errors {
		fmt.Printf("  \033[31m✗  %s\033[0m\n", e)
	}
}

// Validate runs all validation rules on the spec and returns a Result.
func Validate(spec *model.ProjectSpec) *Result {
	r := &Result{}
	checkPortUniqueness(spec, r)
	checkCircularDeps(spec, r)
	checkSagaOwners(spec, r)
	return r
}

func checkPortUniqueness(spec *model.ProjectSpec, r *Result) {
	seen := map[int]string{}
	for _, svc := range spec.Services {
		if prev, ok := seen[svc.Port]; ok {
			r.Errors = append(r.Errors, fmt.Sprintf("port %d is used by both '%s' and '%s'", svc.Port, prev, svc.Name))
		} else {
			seen[svc.Port] = svc.Name
		}
	}
}

func checkCircularDeps(spec *model.ProjectSpec, r *Result) {
	// Build adjacency map
	adj := map[string][]string{}
	names := map[string]bool{}
	for _, svc := range spec.Services {
		names[svc.Name] = true
		adj[svc.Name] = svc.Dependencies
	}

	// DFS cycle detection
	visited := map[string]bool{}
	inStack := map[string]bool{}
	var stack []string

	var dfs func(node string) bool
	dfs = func(node string) bool {
		visited[node] = true
		inStack[node] = true
		stack = append(stack, node)

		for _, dep := range adj[node] {
			if !names[dep] {
				r.Warnings = append(r.Warnings, fmt.Sprintf("service '%s' depends on unknown service '%s'", node, dep))
				continue
			}
			if !visited[dep] {
				if dfs(dep) {
					return true
				}
			} else if inStack[dep] {
				// Find cycle start
				idx := 0
				for i, s := range stack {
					if s == dep {
						idx = i
						break
					}
				}
				cycle := append(stack[idx:], dep)
				r.Errors = append(r.Errors, "circular dependency detected: "+strings.Join(cycle, " → "))
				return true
			}
		}

		stack = stack[:len(stack)-1]
		inStack[node] = false
		return false
	}

	for _, svc := range spec.Services {
		if !visited[svc.Name] {
			dfs(svc.Name)
		}
	}
}

func checkSagaOwners(spec *model.ProjectSpec, r *Result) {
	svcNames := map[string]bool{}
	for _, svc := range spec.Services {
		svcNames[svc.Name] = true
	}
	for _, saga := range spec.Sagas {
		if saga.Owner != "" && !svcNames[saga.Owner] {
			r.Errors = append(r.Errors, fmt.Sprintf("saga '%s' references unknown owner service '%s'", saga.SagaID, saga.Owner))
		}
		for _, step := range saga.Steps {
			if step.Service != "" && !svcNames[step.Service] {
				r.Warnings = append(r.Warnings, fmt.Sprintf("saga '%s' step references unknown service '%s'", saga.SagaID, step.Service))
			}
		}
	}
}
