import React, { useRef, useCallback, useState } from 'react';
import ForceGraph2D from 'react-force-graph-2d';
import * as d3 from 'd3';
import './GraphVisualizer.css';

const GraphVisualizer = ({ data, onNodeClick }) => {
  const fgRef = useRef();
  const [selectedNode, setSelectedNode] = useState(null);
  const [connectedNodes, setConnectedNodes] = useState(new Set());

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

  // Build adjacency when a node is clicked
  const handleNodeClick = useCallback((node) => {
    // Build set of connected node IDs (direct neighbors only)
    const connected = new Set();
    data.links.forEach(link => {
      const sourceId = link.source.id || link.source;
      const targetId = link.target.id || link.target;
      if (sourceId === node.id) {
        connected.add(targetId);
      }
      if (targetId === node.id) {
        connected.add(sourceId);
      }
    });

    setSelectedNode(node.id);
    setConnectedNodes(connected);

    // Center view on node
    fgRef.current.centerAt(node.x, node.y, 1000);
    fgRef.current.zoom(3, 1500);
    onNodeClick(node);
  }, [data.links, onNodeClick]);

  // Determine link color based on selection state
  const getLinkColor = useCallback((link) => {
    if (!selectedNode) {
      return link.color || '#e67e22'; // Default dark orange
    }
    const sourceId = link.source.id || link.source;
    const targetId = link.target.id || link.target;
    
    // Check if this link connects to the selected node
    if (sourceId === selectedNode || targetId === selectedNode) {
      return '#90ee90'; // Light green for connected links
    }
    return '#333'; // Dim for non-connected
  }, [selectedNode]);

  // Determine link width based on selection state
  const getLinkWidth = useCallback((link) => {
    if (!selectedNode) {
      return link.type === 'call' ? 2 : 1;
    }
    const sourceId = link.source.id || link.source;
    const targetId = link.target.id || link.target;
    
    if (sourceId === selectedNode || targetId === selectedNode) {
      return 3; // Thicker for highlighted
    }
    return 0.5; // Thinner for dimmed
  }, [selectedNode]);

  // Custom node canvas rendering for FILE labels
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
    
    // Draw label for FILE nodes (always) or when zoomed in enough
    if (isFile || globalScale > 1.5) {
      const label = node.name;
      const fontSize = isFile ? 12 / globalScale : 10 / globalScale;
      ctx.font = `${fontSize}px Sans-Serif`;
      ctx.textAlign = 'center';
      ctx.textBaseline = 'top';
      ctx.fillStyle = isFile ? '#ffffff' : '#aaaaaa';
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
