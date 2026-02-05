import React, { useRef, useCallback } from 'react';
import ForceGraph2D from 'react-force-graph-2d';
import * as d3 from 'd3';
import './GraphVisualizer.css';

const GraphVisualizer = ({ data, onNodeClick }) => {
  const fgRef = useRef();

  // Configure forces after component mounts
  const handleEngineInit = useCallback(() => {
    const fg = fgRef.current;
    if (fg) {
      // Increase repulsion force between nodes for better spacing
      // Much stronger repulsion to prevent "ball" effect
      fg.d3Force('charge').strength(-2500);
      // Set minimum link distance to prevent clustering
      fg.d3Force('link').distance(150);
      // Reduce center force to let graph expand freely
      fg.d3Force('center').strength(0.005);
      // Add collision force with radius larger than node size
      fg.d3Force('collide', d3.forceCollide(15));
    }
  }, []);

  return (
    <div className="graph-container">
      <ForceGraph2D
        ref={fgRef}
        graphData={data}
        nodeLabel="name"
        nodeColor={node => node.color}
        nodeVal={node => node.val}
        nodeRelSize={6}
        linkColor={link => link.color || '#555'}
        linkWidth={link => link.type === 'call' ? 2 : 1}
        linkDirectionalArrowLength={link => link.type === 'call' ? 6 : 0}
        linkDirectionalArrowRelPos={1}
        backgroundColor="#0f0f0f"
        onNodeClick={node => {
          // Center view on node
          fgRef.current.centerAt(node.x, node.y, 1000);
          fgRef.current.zoom(3, 1500);
          onNodeClick(node);
        }}
        d3VelocityDecay={0.4}
        d3AlphaDecay={0.01}
        cooldownTicks={300}
        warmupTicks={200}
        onEngineStop={() => fgRef.current.zoomToFit(400, 50)}
        onEngineTick={() => {
          // Initialize forces on first tick
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
