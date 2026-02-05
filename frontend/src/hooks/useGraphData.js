import { useState, useEffect } from 'react';
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
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
        const response = await axios.get(`${baseURL}/v1/nodes`);
        const codeNodes = response.data.nodes || [];

        const nodes = [];
        const links = [];
        const fileMap = new Map();

        // 1. Create File Nodes first to act as hubs
        codeNodes.forEach(node => {
          if (!node.file_path) return;
          
          if (!fileMap.has(node.file_path)) {
            const fileName = node.file_path.split('/').pop();
            const fileNode = {
              id: `file:${node.file_path}`,
              name: fileName,
              type: 'FILE',
              val: NODE_VALS.FILE,
              color: NODE_COLORS.FILE,
              fullPath: node.file_path
            };
            fileMap.set(node.file_path, fileNode);
            nodes.push(fileNode);
          }
        });

        // 2. Process Code Nodes and link to File
        codeNodes.forEach(node => {
          const graphNode = {
            id: node.id,
            name: node.name,
            type: node.type, // FUNCTION, CLASS, etc.
            val: NODE_VALS[node.type] || NODE_VALS.DEFAULT,
            color: NODE_COLORS[node.type] || NODE_COLORS.DEFAULT,
            // Store original data for sidebar
            original: node 
          };
          nodes.push(graphNode);

          // Link to File Node
          if (node.file_path && fileMap.has(node.file_path)) {
            links.push({
              source: `file:${node.file_path}`,
              target: node.id,
              color: '#555' // Subtle link color
            });
          }
        });

        setData({ nodes, links });
        setLoading(false);
      } catch (err) {
        console.error("Failed to fetch graph data", err);
        setError(err);
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  return { data, loading, error };
};
