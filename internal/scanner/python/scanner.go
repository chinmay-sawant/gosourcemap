package python

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"

	"github.com/chinmay-sawant/gosourcemapper/internal/models"
	"github.com/google/uuid"
)

type PythonScanner struct{}

func NewPythonScanner() *PythonScanner {
	return &PythonScanner{}
}

func (s *PythonScanner) Scan(filePath string, content []byte) ([]*models.CodeNode, error) {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	var nodes []*models.CodeNode

	reDef := regexp.MustCompile(`^def\s+(\w+)`)
	reClass := regexp.MustCompile(`^class\s+(\w+)`)
	reHTTP := regexp.MustCompile(`(requests\.(get|post|put|delete|patch)|urllib|httpx)`)
	reCMD := regexp.MustCompile(`(subprocess\.run|os\.system|exec)`)
	reMethodCall := regexp.MustCompile(`(\w+)\.(\w+)\s*\(`)
	reFuncCall := regexp.MustCompile(`([a-z_]\w*)\s*\(`)

	lineNumber := 0
	var comments []string
	var currentIndent int
	var currentFunc *models.CodeNode
	var funcBodyLines []string

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Count leading spaces for indentation
		indent := len(line) - len(strings.TrimLeft(line, " \t"))

		if strings.HasPrefix(trimmedLine, "#") {
			comments = append(comments, trimmedLine)
			continue
		}

		// Check if we've exited the current function
		if currentFunc != nil && trimmedLine != "" && indent <= currentIndent {
			// End of function - extract call refs
			refs := extractPythonCallRefs(funcBodyLines, reMethodCall, reFuncCall)
			if len(refs) > 0 {
				currentFunc.UnresolvedRefs = refs
			}
			currentFunc = nil
			funcBodyLines = nil
		}

		// Collect function body
		if currentFunc != nil && trimmedLine != "" {
			funcBodyLines = append(funcBodyLines, line)
		}

		if match := reClass.FindStringSubmatch(trimmedLine); len(match) > 1 {
			nodes = append(nodes, &models.CodeNode{
				ID:         uuid.New().String(),
				Type:       models.NodeClass,
				Name:       match[1],
				Language:   "python",
				FilePath:   filePath,
				LineNumber: lineNumber,
				Comments:   cloneAndReverse(comments),
			})
			comments = nil
			continue
		}

		if match := reDef.FindStringSubmatch(trimmedLine); len(match) > 1 {
			node := &models.CodeNode{
				ID:         uuid.New().String(),
				Type:       models.NodeFunction,
				Name:       match[1],
				Language:   "python",
				FilePath:   filePath,
				LineNumber: lineNumber,
				Comments:   cloneAndReverse(comments),
			}
			nodes = append(nodes, node)
			comments = nil

			// Start tracking function body
			currentFunc = node
			currentIndent = indent
			funcBodyLines = nil
			continue
		}

		if match := reHTTP.FindStringSubmatch(trimmedLine); len(match) > 1 {
			nodes = append(nodes, &models.CodeNode{
				ID:         uuid.New().String(),
				Type:       models.NodeHTTPCall,
				Name:       match[1], // e.g. requests.get
				Language:   "python",
				FilePath:   filePath,
				LineNumber: lineNumber,
			})
		}

		if match := reCMD.FindStringSubmatch(trimmedLine); len(match) > 1 {
			nodes = append(nodes, &models.CodeNode{
				ID:         uuid.New().String(),
				Type:       "CMD_EXEC", // Custom type for Python CMD
				Name:       match[1],
				Language:   "python",
				FilePath:   filePath,
				LineNumber: lineNumber,
			})
		}

		if trimmedLine != "" {
			comments = nil
		}
	}

	// Handle last function
	if currentFunc != nil && len(funcBodyLines) > 0 {
		refs := extractPythonCallRefs(funcBodyLines, reMethodCall, reFuncCall)
		if len(refs) > 0 {
			currentFunc.UnresolvedRefs = refs
		}
	}

	return nodes, nil
}

// extractPythonCallRefs extracts function/method call refs from function body
func extractPythonCallRefs(lines []string, reMethodCall, reFuncCall *regexp.Regexp) []string {
	refSet := make(map[string]bool)
	for _, line := range lines {
		// Method calls: obj.method()
		matches := reMethodCall.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 2 {
				obj := match[1]
				method := match[2]
				// Skip self and common builtins
				if obj != "self" && !isPythonBuiltin(obj) {
					refSet[obj+"."+method] = true
				}
				// Also add just the method name
				refSet[method] = true
			}
		}
		// Direct function calls: func()
		funcMatches := reFuncCall.FindAllStringSubmatch(line, -1)
		for _, match := range funcMatches {
			if len(match) > 1 {
				funcName := match[1]
				if !isPythonBuiltin(funcName) && !isPythonKeyword(funcName) {
					refSet[funcName] = true
				}
			}
		}
	}
	refs := make([]string, 0, len(refSet))
	for ref := range refSet {
		refs = append(refs, ref)
	}
	return refs
}

func isPythonBuiltin(name string) bool {
	builtins := map[string]bool{
		"print": true, "len": true, "range": true, "str": true, "int": true,
		"float": true, "list": true, "dict": true, "set": true, "tuple": true,
		"bool": true, "type": true, "isinstance": true, "hasattr": true,
		"getattr": true, "setattr": true, "open": true, "super": true,
		"enumerate": true, "zip": true, "map": true, "filter": true,
		"sorted": true, "reversed": true, "any": true, "all": true,
		"min": true, "max": true, "sum": true, "abs": true, "round": true,
	}
	return builtins[name]
}

func isPythonKeyword(name string) bool {
	keywords := map[string]bool{
		"if": true, "else": true, "elif": true, "for": true, "while": true,
		"return": true, "def": true, "class": true, "import": true, "from": true,
		"try": true, "except": true, "finally": true, "with": true, "as": true,
		"raise": true, "pass": true, "break": true, "continue": true,
	}
	return keywords[name]
}

func cloneAndReverse(input []string) []string {
	if len(input) == 0 {
		return nil
	}
	output := make([]string, len(input))
	for i, v := range input {
		output[len(input)-1-i] = v
	}
	return output
}
