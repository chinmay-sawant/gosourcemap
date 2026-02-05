package models

type NodeType string

const (
	NodeFunction  NodeType = "FUNCTION"
	NodeInterface NodeType = "INTERFACE"
	NodeHTTPCall  NodeType = "HTTP_CALL"
	NodeClass     NodeType = "CLASS" // For Java/Python
)

// CodeNode represents a semantic unit of code
type CodeNode struct {
	ID           string                 `json:"id"`
	Type         NodeType               `json:"type"`
	Name         string                 `json:"name"`
	Language     string                 `json:"language"`
	FilePath     string                 `json:"file_path"`
	LineNumber   int                    `json:"line_number"`
	Signature    string                 `json:"signature"`    // e.g., "func(a int) error"
	Comments     []string               `json:"comments"`     // Accumulated comments (reverse sorted)
	Metadata     map[string]interface{} `json:"metadata"`     // Language-specific details
	Dependencies []string               `json:"dependencies"` // IDs of other nodes this node calls
}
