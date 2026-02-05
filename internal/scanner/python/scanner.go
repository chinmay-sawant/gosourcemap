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

	lineNumber := 0
	var comments []string

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "#") {
			comments = append(comments, line)
			continue
		}

		if match := reClass.FindStringSubmatch(line); len(match) > 1 {
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

		if match := reDef.FindStringSubmatch(line); len(match) > 1 {
			nodes = append(nodes, &models.CodeNode{
				ID:         uuid.New().String(),
				Type:       models.NodeFunction,
				Name:       match[1],
				Language:   "python",
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
				Name:       match[1], // e.g. requests.get
				Language:   "python",
				FilePath:   filePath,
				LineNumber: lineNumber,
			})
		}

		if match := reCMD.FindStringSubmatch(line); len(match) > 1 {
			nodes = append(nodes, &models.CodeNode{
				ID:         uuid.New().String(),
				Type:       "CMD_EXEC", // Custom type for Python CMD
				Name:       match[1],
				Language:   "python",
				FilePath:   filePath,
				LineNumber: lineNumber,
			})
		}

		if line != "" {
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
	for i, v := range input {
		output[len(input)-1-i] = v
	}
	return output
}
