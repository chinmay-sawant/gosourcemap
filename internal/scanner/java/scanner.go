package java

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/chinmay-sawant/gosourcemapper/internal/models"
	"github.com/google/uuid"
)

type JavaScanner struct{}

func NewJavaScanner() *JavaScanner {
	return &JavaScanner{}
}

func (s *JavaScanner) Scan(filePath string, content []byte) ([]*models.CodeNode, error) {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	var nodes []*models.CodeNode

	// Heuristics (Regex)
	// Spring Components
	reClass := regexp.MustCompile(`(public|protected|private)?\s*(class|interface)\s+(\w+)`)
	reMethod := regexp.MustCompile(`(public|protected|private)\s+[\w<>]+\s+(\w+)\s*\(.*\)`)
	reHTTP := regexp.MustCompile(`(RestTemplate|WebClient|HttpClient|MockMvc)`) // Usage
	reMethodCall := regexp.MustCompile(`(\w+)\.(\w+)\s*\(`)                     // Method calls

	var currentClassName string
	var comments []string
	lineNumber := 0

	// Track methods and their call refs
	type methodInfo struct {
		node    *models.CodeNode
		content []string // Lines of the method body
	}
	var currentMethod *methodInfo
	var braceCount int

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "//") || strings.HasPrefix(trimmedLine, "*") || strings.HasPrefix(trimmedLine, "/*") {
			comments = append(comments, trimmedLine)
			continue
		}

		if match := reClass.FindStringSubmatch(line); len(match) > 3 {
			nodeType := models.NodeClass
			if match[2] == "interface" {
				nodeType = models.NodeInterface
			}
			currentClassName = match[3]
			nodes = append(nodes, &models.CodeNode{
				ID:         uuid.New().String(),
				Type:       nodeType,
				Name:       currentClassName,
				Language:   "java",
				FilePath:   filePath,
				LineNumber: lineNumber,
				Comments:   cloneAndReverse(comments),
			})
			comments = nil // Reset
			continue
		}

		if match := reMethod.FindStringSubmatch(line); len(match) > 2 {
			methodName := match[2]
			fullName := methodName
			if currentClassName != "" {
				fullName = fmt.Sprintf("%s.%s", currentClassName, methodName)
			}
			node := &models.CodeNode{
				ID:         uuid.New().String(),
				Type:       models.NodeFunction,
				Name:       fullName,
				Language:   "java",
				FilePath:   filePath,
				LineNumber: lineNumber,
				Comments:   cloneAndReverse(comments),
			}
			nodes = append(nodes, node)
			comments = nil

			// Start tracking method body
			currentMethod = &methodInfo{node: node}
			braceCount = strings.Count(line, "{") - strings.Count(line, "}")
			if braceCount > 0 {
				currentMethod.content = append(currentMethod.content, line)
			}
			continue
		}

		// Track method body
		if currentMethod != nil {
			currentMethod.content = append(currentMethod.content, line)
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")
			if braceCount <= 0 {
				// End of method - extract call refs
				refs := extractJavaCallRefs(currentMethod.content, reMethodCall)
				if len(refs) > 0 {
					currentMethod.node.UnresolvedRefs = refs
				}
				currentMethod = nil
			}
			continue
		}

		if match := reHTTP.FindStringSubmatch(line); len(match) > 1 {
			nodes = append(nodes, &models.CodeNode{
				ID:         uuid.New().String(),
				Type:       models.NodeHTTPCall,
				Name:       match[1],
				Language:   "java",
				FilePath:   filePath,
				LineNumber: lineNumber,
			})
		}

		if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "@") {
			// Reset comments if we hit code that isn't a class/method
			comments = nil
		}
	}

	return nodes, nil
}

// extractJavaCallRefs extracts method call references from method body lines
func extractJavaCallRefs(lines []string, reMethodCall *regexp.Regexp) []string {
	refSet := make(map[string]bool)
	for _, line := range lines {
		matches := reMethodCall.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 2 {
				ref := match[1] + "." + match[2]
				// Skip common patterns that aren't method calls
				if !isJavaBuiltin(match[1]) {
					refSet[ref] = true
					// Also add just the method name
					refSet[match[2]] = true
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

// isJavaBuiltin checks if the identifier is a common Java builtin
func isJavaBuiltin(name string) bool {
	builtins := map[string]bool{
		"System": true, "String": true, "Integer": true, "Long": true,
		"Double": true, "Float": true, "Boolean": true, "List": true,
		"Map": true, "Set": true, "Arrays": true, "Collections": true,
		"Optional": true, "Stream": true, "Objects": true, "Math": true,
	}
	return builtins[name]
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
