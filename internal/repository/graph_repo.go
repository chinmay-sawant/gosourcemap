package repository

import (
	"path/filepath"
	"strings"
	"sync"

	"github.com/chinmay-sawant/gosourcemapper/internal/models"
)

type GraphRepository interface {
	SaveNode(node *models.CodeNode)
	GetNode(id string) (*models.CodeNode, bool)
	GetAllNodes() []*models.CodeNode
	GetNodesPaginated(offset, limit int, skipExts, skipDirs []string) ([]*models.CodeNode, int, error) // returns nodes, total, error
	Clear()
}

type InMemoryGraphRepository struct {
	mu         sync.RWMutex
	nodes      map[string]*models.CodeNode
	orderedIDs []string
}

func NewInMemoryGraphRepository() *InMemoryGraphRepository {
	return &InMemoryGraphRepository{
		nodes:      make(map[string]*models.CodeNode),
		orderedIDs: make([]string, 0),
	}
}

func (r *InMemoryGraphRepository) SaveNode(node *models.CodeNode) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Check update vs insert
	if _, exists := r.nodes[node.ID]; !exists {
		r.orderedIDs = append(r.orderedIDs, node.ID)
	}
	r.nodes[node.ID] = node
}

func (r *InMemoryGraphRepository) GetNode(id string) (*models.CodeNode, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	node, exists := r.nodes[id]
	return node, exists
}

func (r *InMemoryGraphRepository) GetAllNodes() []*models.CodeNode {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Return in order
	nodes := make([]*models.CodeNode, 0, len(r.orderedIDs))
	for _, id := range r.orderedIDs {
		if node, ok := r.nodes[id]; ok {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (r *InMemoryGraphRepository) GetNodesPaginated(offset, limit int, skipExts, skipDirs []string) ([]*models.CodeNode, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Pre-process filters
	skipExtMap := make(map[string]bool)
	for _, ext := range skipExts {
		skipExtMap[strings.ToLower(ext)] = true
	}

	nodes := make([]*models.CodeNode, 0, limit)
	// Standard cursor pagination usually returns items after cursor.
	// But here "total" usually means total available items to help frontend know "Estimate".
	// However, calculating total filtered items is O(N).
	// Let's stick to traversing.

	// Strategy:
	// We need to Skip 'offset' number of MATCHING items.
	// Then collect 'limit' number of MATCHING items.
	// This makes random access slow (O(N)), but 'nextToken' is just an index in 'orderedIDs'.
	// Actually, if nextToken is index in orderedIDs, we just start there.
	// But if we filter, the "offset" index in orderedIDs might jump.
	// IF nextToken = "Index in orderedIDs", then:
	// 1. Start at `orderedIDs[token]`
	// 2. Iterate forward.
	// 3. Check filter.
	// 4. If Match -> Add to result.
	// 5. If Result.len == limit -> break.
	// 6. NextToken = Current Index + 1.

	// We are changing "offset" semantics from "Number of items to skip" to "Index to start at".
	// My previous implementation used `ids := r.orderedIDs[offset:end]`. That assumes offset IS the index.
	// That works fine.

	startIndex := offset
	if startIndex >= len(r.orderedIDs) {
		return []*models.CodeNode{}, len(r.orderedIDs), nil
	}

	currentIdx := startIndex
	for currentIdx < len(r.orderedIDs) && len(nodes) < limit {
		id := r.orderedIDs[currentIdx]
		node, ok := r.nodes[id]
		if !ok {
			currentIdx++
			continue
		}

		// Filter Logic
		if shouldSkip(node, skipExtMap, skipDirs) {
			currentIdx++
			continue
		}

		nodes = append(nodes, node)
		currentIdx++
	}

	// Returned length is how far we went in the original array
	// nextToken calculation needs the *index* we stopped at.
	// The `total` could be just len(orderedIDs) or total matching.
	// Let's return len(orderedIDs) as "Total Scanned DB Size".
	return nodes, currentIdx, nil
}

func shouldSkip(node *models.CodeNode, skipExts map[string]bool, skipDirs []string) bool {
	// Check extension
	ext := strings.ToLower(filepath.Ext(node.FilePath))
	if skipExts[ext] {
		return true
	}

	// Check directories
	for _, dir := range skipDirs {
		// "skip_dir" e.g., "venv". If path contains "/venv/" or starts with "venv/"
		// Simple approach: strings.Contains for now, or check path segments.
		// "it should ignore the python specific directories venv"
		// Better: split path and check.
		// Or simpler: `strings.Contains(path, "/"+dir+"/")` ?
		// Use filepath.Split logic ideally, but text search is faster for prototype.
		// Let's check segments to be safe? No, contains is fine if "dir" is distinct.
		// Edge case: "my_venv" contains "venv". User probably wants exact dir name match.
		// Let's use simple containment for now but guard with slashes.
		// Normalized path
		cleaned := filepath.ToSlash(node.FilePath)
		if strings.Contains(cleaned, "/"+dir+"/") || strings.HasPrefix(cleaned, dir+"/") {
			return true
		}
	}

	return false
}

func (r *InMemoryGraphRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodes = make(map[string]*models.CodeNode)
	r.orderedIDs = make([]string, 0)
}
