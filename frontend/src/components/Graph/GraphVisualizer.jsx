import React, { useRef, useCallback, useState, useMemo } from 'react';
import ForceGraph2D from 'react-force-graph-2d';
import * as d3 from 'd3';
import './GraphVisualizer.css';

const GraphVisualizer = ({ data, onNodeClick }) => {
  const fgRef = useRef();
  const [selectedNode, setSelectedNode] = useState(null);
  const [connectedNodes, setConnectedNodes] = useState(new Set());

  // Build REVERSE adjacency list from links (memoized for performance)
  // Only track predecessors/callers, NOT successors/callees
  // If link is source -> target, we store: target -> source (reverse direction)
  // This allows us to traverse BACKWARDS from a node to find its callers
  const reverseAdjacencyList = useMemo(() => {
    const adj = new Map();
    data.links.forEach(link => {
      const sourceId = link.source.id || link.source;
      const targetId = link.target.id || link.target;
      
      // Only store reverse direction: target points to source (caller)
      if (!adj.has(targetId)) adj.set(targetId, new Set());
      adj.get(targetId).add(sourceId);
    });
    return adj;
  }, [data.links]);

  // Configure forces after component mounts
  const handleEngineInit = useCallback(() => {
    const fg = fgRef.current;
    if (fg) {
      fg.d3Force('charge').strength(-2500);
      fg.d3Force('link').distance(150);
      fg.d3Force('center').strength(0.005);
      fg.d3Force('collide', d3.forceCollide(15));
    }
  }, []);

  // BFS to find all PREDECESSOR nodes (callers/sources) - goes BACKWARDS only
  // If chain is: 1 -> 2 -> 3 -> 4, clicking 3 will find: 2, 1
  const findAllPredecessorNodes = useCallback((startNodeId) => {
    const visited = new Set();
    const queue = [startNodeId];
    
    while (queue.length > 0) {
      const currentId = queue.shift();
      if (visited.has(currentId)) continue;
      visited.add(currentId);
      
      // Get predecessors (nodes that call/link to current node)
      const predecessors = reverseAdjacencyList.get(currentId);
      if (predecessors) {
        predecessors.forEach(predecessorId => {
          if (!visited.has(predecessorId)) {
            queue.push(predecessorId);
          }
        });
      }
    }
    
    // Remove the start node from connected set (it's the selected node)
    visited.delete(startNodeId);
    return visited;
  }, [reverseAdjacencyList]);

  // Build adjacency when a node is clicked - traces BACKWARDS to callers
  const handleNodeClick = useCallback((node) => {
    // Find ALL predecessor nodes (callers) using BFS - goes backwards only
    const connected = findAllPredecessorNodes(node.id);

    setSelectedNode(node.id);
    setConnectedNodes(connected);

    // Center view on node
    fgRef.current.centerAt(node.x, node.y, 1000);
    fgRef.current.zoom(3, 1500);
    onNodeClick(node);
  }, [findAllPredecessorNodes, onNodeClick]);

  // Determine link color based on selection state
  const getLinkColor = useCallback((link) => {
    if (!selectedNode) {
      return link.color || '#e67e22'; // Default dark orange
    }
    const sourceId = link.source.id || link.source;
    const targetId = link.target.id || link.target;
    
    // Check if BOTH ends of the link are in the connected set (including selected node)
    const sourceInSet = sourceId === selectedNode || connectedNodes.has(sourceId);
    const targetInSet = targetId === selectedNode || connectedNodes.has(targetId);
    
    if (sourceInSet && targetInSet) {
      return '#90ee90'; // Light green for connected links
    }
    return '#333'; // Dim for non-connected
  }, [selectedNode, connectedNodes]);

  // Determine link width based on selection state
  const getLinkWidth = useCallback((link) => {
    if (!selectedNode) {
      return link.type === 'call' ? 2 : 1;
    }
    const sourceId = link.source.id || link.source;
    const targetId = link.target.id || link.target;
    
    // Check if BOTH ends of the link are in the connected set (including selected node)
    const sourceInSet = sourceId === selectedNode || connectedNodes.has(sourceId);
    const targetInSet = targetId === selectedNode || connectedNodes.has(targetId);
    
    if (sourceInSet && targetInSet) {
      return 3; // Thicker for highlighted
    }
    return 0.5; // Thinner for dimmed
  }, [selectedNode, connectedNodes]);

  // Custom node canvas rendering - now shows labels for connected nodes
  const nodeCanvasObject = useCallback((node, ctx, globalScale) => {
    const isFile = node.type === 'FILE';
    const isSelected = node.id === selectedNode;
    const isConnected = connectedNodes.has(node.id);
    
    // Determine radius
    const baseRadius = isFile ? 12 : 6;
    const radius = baseRadius * (node.val ? Math.sqrt(node.val) / 2 : 1);
    
    // Determine color
    let color = node.color;
    if (isSelected) {
      color = '#00ff88';
    } else if (selectedNode && isConnected) {
      color = '#90ee90';
    }
    
    // Draw circle
    ctx.beginPath();
    ctx.arc(node.x, node.y, radius, 0, 2 * Math.PI);
    ctx.fillStyle = color;
    ctx.fill();
    
    // Draw label for:
    // 1. FILE nodes (always)
    // 2. Selected or connected nodes (when there's a selection)
    // 3. Any node when zoomed in enough
    const shouldShowLabel = isFile || (selectedNode && (isSelected || isConnected)) || globalScale > 1.5;
    
    if (shouldShowLabel) {
      const label = node.name;
      const fontSize = isFile ? 12 / globalScale : 10 / globalScale;
      ctx.font = `${fontSize}px Sans-Serif`;
      ctx.textAlign = 'center';
      ctx.textBaseline = 'top';
      ctx.fillStyle = isFile ? '#ffffff' : (isSelected || isConnected) ? '#ffffff' : '#aaaaaa';
      ctx.fillText(label, node.x, node.y + radius + 2);
    }
  }, [selectedNode, connectedNodes]);

  // Clear selection when clicking background
  const handleBackgroundClick = useCallback(() => {
    setSelectedNode(null);
    setConnectedNodes(new Set());
  }, []);

  return (
    <div className="graph-container">
      <ForceGraph2D
        ref={fgRef}
        graphData={data}
        nodeLabel={node => node.type !== 'FILE' ? node.name : null}
        nodeCanvasObject={nodeCanvasObject}
        nodePointerAreaPaint={(node, color, ctx) => {
          const isFile = node.type === 'FILE';
          const radius = isFile ? 12 : 6;
          ctx.fillStyle = color;
          ctx.beginPath();
          ctx.arc(node.x, node.y, radius * 1.5, 0, 2 * Math.PI);
          ctx.fill();
        }}
        linkColor={getLinkColor}
        linkWidth={getLinkWidth}
        linkDirectionalArrowLength={link => link.type === 'call' ? 6 : 0}
        linkDirectionalArrowRelPos={1}
        backgroundColor="#0f0f0f"
        onNodeClick={handleNodeClick}
        onBackgroundClick={handleBackgroundClick}
        d3VelocityDecay={0.4}
        d3AlphaDecay={0.01}
        cooldownTicks={300}
        warmupTicks={200}
        onEngineStop={() => fgRef.current.zoomToFit(400, 50)}
        onEngineTick={() => {
          if (fgRef.current && !fgRef.current._forcesConfigured) {
            handleEngineInit();
            fgRef.current._forcesConfigured = true;
          }
        }}
      />
    </div>
  );
};

export default GraphVisualizer;
