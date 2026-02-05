import { useState, useEffect, useCallback } from 'react';
import axios from 'axios';

const NODE_COLORS = {
  FILE: '#ffffff',
  FUNCTION: '#61dafb', // React Blue
  CLASS: '#f1c40f',    // Yellow
  INTERFACE: '#2ecc71',// Green
  HTTP_CALL: '#ff6b6b',// Red
  CMD_EXEC: '#e67e22', // Orange
  DEFAULT: '#95a5a6'
};

const NODE_VALS = {
  FILE: 10,
  FUNCTION: 5,
  CLASS: 7,
  INTERFACE: 6,
  HTTP_CALL: 3,
  CMD_EXEC: 3,
  DEFAULT: 3
};

export const useGraphData = () => {
  const [data, setData] = useState({ nodes: [], links: [] });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [nextToken, setNextToken] = useState(null);
  const [limit, setLimit] = useState(100); // Default limit

  const [filters, setFilters] = useState({ 
    skipExt: ['.java', '.py', '.js', '.css', '.html', '.json', '.md', '.txt', '.xml', '.yml', '.yaml'], 
    skipDir: ['node_modules', 'dist', 'build', 'venv', '.git']
  });

  const fetchNodes = useCallback(async (token = null, reset = false, currentFilters = filters) => {
    setLoading(true);
    try {
      const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
      const params = { 
        limit,
        skip_ext: currentFilters.skipExt.join(','),
        skip_dir: currentFilters.skipDir.join(',')
      };
      if (token) params.nextToken = token;

      const response = await axios.get(`${baseURL}/v1/nodes`, { params });
      const { nodes: codeNodes, nextToken: newToken } = response.data;

      const newNodes = [];
      const newLinks = [];
      
      codeNodes.forEach(node => {
        // Code Node
        newNodes.push({
          id: node.id,
          name: node.name,
          type: node.type, 
          val: NODE_VALS[node.type] || NODE_VALS.DEFAULT,
          color: NODE_COLORS[node.type] || NODE_COLORS.DEFAULT,
          original: node
        });

        // File Node (Synthetic)
        if (node.file_path) {
             const fileId = `file:${node.file_path}`;
             newNodes.push({
              id: fileId,
              name: node.file_path.split('/').pop(),
              type: 'FILE',
              val: NODE_VALS.FILE,
              color: NODE_COLORS.FILE,
              fullPath: node.file_path
            });

            newLinks.push({
              source: fileId,
              target: node.id,
              color: '#444',
              type: 'file'
            });
        }

        // Dependency Links (function calls)
        if (node.dependencies && node.dependencies.length > 0) {
          node.dependencies.forEach(depId => {
            newLinks.push({
              source: node.id,
              target: depId,
              color: '#e67e22', // Orange for call relationships
              type: 'call'
            });
          });
        }
      });

      setData(prev => {
        if (reset) {
            // Dedup within the new batch only
            const uniqueNodes = new Map();
            newNodes.forEach(n => uniqueNodes.set(n.id, n));
            const validNodes = Array.from(uniqueNodes.values());
            const validNodeKeys = new Set(validNodes.map(n => n.id));

            // Filter links to ensures both endpoints exist
            const validLinks = newLinks.filter(l => 
                validNodeKeys.has(l.source) && validNodeKeys.has(l.target)
            );
            
            return { nodes: validNodes, links: validLinks };
        }

        // Merge logic
        const combinedNodes = [...prev.nodes];
        const existingIds = new Set(prev.nodes.map(n => n.id));
        
        newNodes.forEach(n => {
            if (!existingIds.has(n.id)) {
                combinedNodes.push(n);
                existingIds.add(n.id);
            }
        });

        const allNodeIds = new Set(combinedNodes.map(n => n.id));
        const allLinks = [...prev.links, ...newLinks].filter(l => {
            const sourceId = (l.source && l.source.id) || l.source;
            const targetId = (l.target && l.target.id) || l.target;
            return allNodeIds.has(sourceId) && allNodeIds.has(targetId);
        });

        return {
            nodes: combinedNodes,
            links: allLinks
        };
      });

      setNextToken(newToken);
      setLoading(false);
    } catch (err) {
      console.error("Failed to fetch graph data", err);
      setError(err);
      setLoading(false);
    }
  }, [limit, filters]); // Re-fetch if filters change? Or requires manual trigger?
  // Ideally, if filters change, we should reset.

  // Initial load
  useEffect(() => {
    // Use a microtask to avoid synchronous setState inside useEffect which triggers lint error
    Promise.resolve().then(() => {
      fetchNodes(null, true);
    });
  }, [fetchNodes]);

  const loadMore = () => {
    if (nextToken) {
        fetchNodes(nextToken, false);
    }
  };

  const applyFilters = (newFilters) => {
    setFilters(newFilters);
    // Trigger reset fetch
    fetchNodes(null, true, newFilters);
  };

  return { 
    data, 
    loading, 
    error, 
    loadMore, 
    hasMore: !!nextToken, 
    setLimit, 
    limit, 
    refresh: () => fetchNodes(null, true),
    filters,
    applyFilters
  };
};
