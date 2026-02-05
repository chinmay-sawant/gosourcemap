package scanner

import "github.com/chinmay-sawant/gosourcemapper/internal/models"

type Scanner interface {
	// Scan parses the given file content and returns a list of CodeNodes
	Scan(filePath string, content []byte) ([]*models.CodeNode, error)
}
