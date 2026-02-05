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

	var currentClassName string
	var comments []string
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "*") || strings.HasPrefix(line, "/*") {
			comments = append(comments, line)
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
			nodes = append(nodes, &models.CodeNode{
				ID:         uuid.New().String(),
				Type:       models.NodeFunction,
				Name:       fullName,
				Language:   "java",
				FilePath:   filePath,
				LineNumber: lineNumber,
				Comments:   cloneAndReverse(comments),
			})
			comments = nil
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

		if line != "" && !strings.HasPrefix(line, "@") {
			// Reset comments if we hit code that isn't a class/method
			comments = nil
		}
	}

	return nodes, nil
}

func cloneAndReverse(input []string) []string {
	if len(input) == 0 {
		return nil
	}
	output := make([]string, len(input))
	// Regular copy. The prompt requirement was "going up from the function", which implies simple preceeding order
	// or reverse order. My Go implementation does closest-first.
	// For this stream implementation, the last comment added is the closest.
	// So reversing it makes it closest-first (bottom-up).
	for i, v := range input {
		output[len(input)-1-i] = v
	}
	return output
}
