package golang

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/chinmay-sawant/gosourcemapper/internal/models"
	"github.com/google/uuid"
)

type GoScanner struct{}

func NewGoScanner() *GoScanner {
	return &GoScanner{}
}

// funcInfo stores function info for dependency resolution
type funcInfo struct {
	node     *models.CodeNode
	funcDecl *ast.FuncDecl
}

func (s *GoScanner) Scan(filePath string, content []byte) ([]*models.CodeNode, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var nodes []*models.CodeNode
	var funcInfos []funcInfo

	// Phase 1: Collect all functions, interfaces, and HTTP calls
	ast.Inspect(file, func(n ast.Node) bool {
		switch t := n.(type) {
		case *ast.FuncDecl:
			codeNode := s.parseFunction(fset, file, t, filePath)
			nodes = append(nodes, codeNode)
			funcInfos = append(funcInfos, funcInfo{node: codeNode, funcDecl: t})
		case *ast.TypeSpec:
			if _, ok := t.Type.(*ast.InterfaceType); ok {
				nodes = append(nodes, s.parseInterface(fset, file, t, filePath))
			}
		case *ast.CallExpr:
			// Detect HTTP Clients (Basic Heuristic)
			if httpNode := s.parseHTTPCall(fset, t, filePath); httpNode != nil {
				nodes = append(nodes, httpNode)
			}
		}
		return true
	})

	// Phase 2: Extract unresolved call references for cross-file resolution
	for _, fi := range funcInfos {
		if fi.funcDecl.Body == nil {
			continue
		}
		refs := s.extractCallRefs(fi.funcDecl.Body)
		if len(refs) > 0 {
			fi.node.UnresolvedRefs = refs
		}
	}

	return nodes, nil
}

// extractCallRefs extracts function call references from a function body
func (s *GoScanner) extractCallRefs(body *ast.BlockStmt) []string {
	refSet := make(map[string]bool)

	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		ref := s.extractCallRef(call)
		if ref != "" {
			refSet[ref] = true
		}

		return true
	})

	refs := make([]string, 0, len(refSet))
	for ref := range refSet {
		refs = append(refs, ref)
	}
	return refs
}

// extractCallRef extracts a call reference string from a CallExpr
func (s *GoScanner) extractCallRef(call *ast.CallExpr) string {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		// Direct function call: foo()
		return fn.Name
	case *ast.SelectorExpr:
		// Method or package call: obj.Method() or pkg.Func()
		return s.selectorToString(fn)
	}
	return ""
}

// selectorToString converts a SelectorExpr to a dotted string
// e.g., handlers.RegisterRoutes or h.service.GetNodes
func (s *GoScanner) selectorToString(sel *ast.SelectorExpr) string {
	var parts []string
	current := ast.Expr(sel)

	for {
		switch t := current.(type) {
		case *ast.SelectorExpr:
			parts = append([]string{t.Sel.Name}, parts...)
			current = t.X
		case *ast.Ident:
			parts = append([]string{t.Name}, parts...)
			return joinParts(parts)
		default:
			return joinParts(parts)
		}
	}
}

func joinParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += "." + parts[i]
	}
	return result
}

func (s *GoScanner) parseFunction(fset *token.FileSet, file *ast.File, fn *ast.FuncDecl, filePath string) *models.CodeNode {
	comments := s.extractComments(fset, file, fn.Pos())
	name := fn.Name.Name
	if fn.Recv != nil {
		// Method
		for _, field := range fn.Recv.List {
			typeExpr := field.Type
			if star, ok := typeExpr.(*ast.StarExpr); ok {
				if ident, ok := star.X.(*ast.Ident); ok {
					name = fmt.Sprintf("(%s).%s", ident.Name, name)
				}
			} else if ident, ok := typeExpr.(*ast.Ident); ok {
				name = fmt.Sprintf("(%s).%s", ident.Name, name)
			}
		}
	}

	return &models.CodeNode{
		ID:         uuid.New().String(),
		Type:       models.NodeFunction,
		Name:       name,
		Language:   "go",
		FilePath:   filePath,
		LineNumber: fset.Position(fn.Pos()).Line,
		Comments:   comments,
	}
}

func (s *GoScanner) parseInterface(fset *token.FileSet, file *ast.File, typeSpec *ast.TypeSpec, filePath string) *models.CodeNode {
	comments := s.extractComments(fset, file, typeSpec.Pos())
	return &models.CodeNode{
		ID:         uuid.New().String(),
		Type:       models.NodeInterface,
		Name:       typeSpec.Name.Name,
		Language:   "go",
		FilePath:   filePath,
		LineNumber: fset.Position(typeSpec.Pos()).Line,
		Comments:   comments,
	}
}

func (s *GoScanner) parseHTTPCall(fset *token.FileSet, call *ast.CallExpr, filePath string) *models.CodeNode {
	// Heuristic: Check for http.Get, http.Post, http.NewRequest
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok {
			if ident.Name == "http" && (sel.Sel.Name == "Get" || sel.Sel.Name == "Post" || sel.Sel.Name == "NewRequest") {
				return &models.CodeNode{
					ID:         uuid.New().String(),
					Type:       models.NodeHTTPCall,
					Name:       fmt.Sprintf("http.%s", sel.Sel.Name),
					Language:   "go",
					FilePath:   filePath,
					LineNumber: fset.Position(call.Pos()).Line,
				}
			}
		}
	}
	return nil
}

// extractComments recursively finds comments appearing immediately before the position.
func (s *GoScanner) extractComments(fset *token.FileSet, file *ast.File, pos token.Pos) []string {
	var relevantGroups []*ast.CommentGroup
	targetLine := fset.Position(pos).Line - 1

	for {
		found := false
		for _, cg := range file.Comments {
			endLine := fset.Position(cg.End()).Line
			if endLine == targetLine {
				relevantGroups = append(relevantGroups, cg)
				startLine := fset.Position(cg.Pos()).Line
				targetLine = startLine - 1
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	var comments []string
	for _, cg := range relevantGroups {
		comments = append(comments, cg.Text())
	}

	return comments
}
