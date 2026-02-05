package repository

import (
	"sync"

	"github.com/chinmay-sawant/gosourcemapper/internal/models"
)

type GraphRepository interface {
	SaveNode(node *models.CodeNode)
	GetNode(id string) (*models.CodeNode, bool)
	GetAllNodes() []*models.CodeNode
	Clear()
}

type InMemoryGraphRepository struct {
	mu    sync.RWMutex
	nodes map[string]*models.CodeNode
}

func NewInMemoryGraphRepository() *InMemoryGraphRepository {
	return &InMemoryGraphRepository{
		nodes: make(map[string]*models.CodeNode),
	}
}

func (r *InMemoryGraphRepository) SaveNode(node *models.CodeNode) {
	r.mu.Lock()
	defer r.mu.Unlock()
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
	nodes := make([]*models.CodeNode, 0, len(r.nodes))
	for _, node := range r.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

func (r *InMemoryGraphRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodes = make(map[string]*models.CodeNode)
}
