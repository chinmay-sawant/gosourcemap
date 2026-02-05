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

func (s *GoScanner) Scan(filePath string, content []byte) ([]*models.CodeNode, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var nodes []*models.CodeNode

	ast.Inspect(node, func(n ast.Node) bool {
		switch t := n.(type) {
		case *ast.FuncDecl:
			nodes = append(nodes, s.parseFunction(fset, node, t, filePath))
		case *ast.TypeSpec:
			if _, ok := t.Type.(*ast.InterfaceType); ok {
				nodes = append(nodes, s.parseInterface(fset, node, t, filePath))
			}
		case *ast.CallExpr:
			// Detect HTTP Clients (Basic Heuristic)
			if httpNode := s.parseHTTPCall(fset, t, filePath); httpNode != nil {
				nodes = append(nodes, httpNode)
			}
		}
		return true
	})

	return nodes, nil
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

	// Iterate all comment groups in the file
	// Since file.Comments is sorted by position, we can iterate backwards or standard.
	// We want the group that ends exactly at targetLine.

	// 1. Find the comment group that is immediately above the function
	// 2. If found, new targetLine becomes the line above that comment group
	// 3. Repeat

	// Build a map of endLine -> CommentGroup for O(1) lookup would be ideal,
	// but iterating is fine for file-level operations.

	for {
		found := false
		for _, cg := range file.Comments {
			endLine := fset.Position(cg.End()).Line
			if endLine == targetLine {
				relevantGroups = append(relevantGroups, cg)
				startLine := fset.Position(cg.Pos()).Line
				targetLine = startLine - 1
				found = true
				break // Start search again with new targetLine
			}
		}
		if !found {
			// Check for empty lines? The prompt says "if there are more comments above... recursively add"
			// Implicitly assuming contiguity or standard go comments.
			// If we didn't find a comment ending at targetLine, we stop.
			// NOTE: This simple logic handles "contiguous" blocks.
			// If there's a blank line, targetLine won't match.
			break
		}
	}

	// Process gathered groups
	var comments []string
	for _, cg := range relevantGroups {
		comments = append(comments, cg.Text())
	}

	// Sort: user requested "sort it revert i guess cause will be going up"
	// The loop finds closest first, so it is already closest-to-farthest (bottom-up).
	// So 'comments' index 0 is the one right above the function.

	return comments
}
