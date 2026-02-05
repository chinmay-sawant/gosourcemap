package resolver

import (
	"strings"

	"github.com/chinmay-sawant/gosourcemapper/internal/models"
)

// DependencyResolver resolves cross-file function call dependencies
type DependencyResolver struct {
	// registry maps qualified names to node IDs
	// Keys: "FuncName", "(Type).Method", "pkg.Func", "ClassName.method"
	registry map[string]string
}

// NewDependencyResolver creates a new resolver instance
func NewDependencyResolver() *DependencyResolver {
	return &DependencyResolver{
		registry: make(map[string]string),
	}
}

// BuildRegistry builds a lookup map from all scanned nodes
// This should be called after all files have been scanned
func (r *DependencyResolver) BuildRegistry(nodes []*models.CodeNode) {
	for _, node := range nodes {
		if node.Type != models.NodeFunction && node.Type != models.NodeClass && node.Type != models.NodeInterface {
			continue
		}

		// Register by exact name
		r.registry[node.Name] = node.ID

		// For methods like "(Type).Method" or "ClassName.method", also register just the method name
		if idx := strings.LastIndex(node.Name, "."); idx != -1 {
			methodName := node.Name[idx+1:]
			// Only register if not already taken (avoid collisions)
			if _, exists := r.registry[methodName]; !exists {
				r.registry[methodName] = node.ID
			}
		}

		// For Go methods "(Type).Method", also register without parens
		if strings.HasPrefix(node.Name, "(") {
			if closeIdx := strings.Index(node.Name, ")"); closeIdx != -1 && closeIdx < len(node.Name)-1 {
				typeName := node.Name[1:closeIdx]
				methodPart := node.Name[closeIdx+1:] // ".Method"
				if strings.HasPrefix(methodPart, ".") {
					methodName := methodPart[1:]
					// Register as "Type.Method" (without parens)
					altKey := typeName + "." + methodName
					if _, exists := r.registry[altKey]; !exists {
						r.registry[altKey] = node.ID
					}
				}
			}
		}
	}
}

// ResolveAll resolves UnresolvedRefs to Dependencies for all nodes
func (r *DependencyResolver) ResolveAll(nodes []*models.CodeNode) {
	for _, node := range nodes {
		if len(node.UnresolvedRefs) == 0 {
			continue
		}

		depSet := make(map[string]bool)
		for _, ref := range node.UnresolvedRefs {
			if id := r.resolve(ref); id != "" && id != node.ID {
				depSet[id] = true
			}
		}

		// Convert set to slice
		if len(depSet) > 0 {
			deps := make([]string, 0, len(depSet))
			for id := range depSet {
				deps = append(deps, id)
			}
			node.Dependencies = deps
		}

		// Clear temporary refs
		node.UnresolvedRefs = nil
	}
}

// resolve tries to find a node ID for the given reference
// Matching priority:
// 1. Exact match
// 2. Suffix match (ref matches end of registered name)
// 3. Method name match
func (r *DependencyResolver) resolve(ref string) string {
	// 1. Exact match
	if id, ok := r.registry[ref]; ok {
		return id
	}

	// 2. Try without package prefix: "pkg.Func" -> look for "Func"
	if idx := strings.LastIndex(ref, "."); idx != -1 {
		suffix := ref[idx+1:]
		if id, ok := r.registry[suffix]; ok {
			return id
		}
	}

	// 3. Suffix match: look for any key ending with ".Ref"
	searchSuffix := "." + ref
	for key, id := range r.registry {
		if strings.HasSuffix(key, searchSuffix) {
			return id
		}
	}

	// 4. For Go-style receiver calls like "h.service.GetNodes", try "GetNodes"
	parts := strings.Split(ref, ".")
	if len(parts) > 1 {
		lastPart := parts[len(parts)-1]
		if id, ok := r.registry[lastPart]; ok {
			return id
		}
	}

	return ""
}
