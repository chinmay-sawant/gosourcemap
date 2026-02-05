import React, { useRef } from 'react';
import ForceGraph2D from 'react-force-graph-2d';
import './GraphVisualizer.css';

const GraphVisualizer = ({ data, onNodeClick }) => {
  const fgRef = useRef();

  return (
    <div className="graph-container">
      <ForceGraph2D
        ref={fgRef}
        graphData={data}
        nodeLabel="name"
        nodeColor={node => node.color}
        nodeVal={node => node.val}
        linkColor={() => '#444'}
        backgroundColor="#0f0f0f"
        onNodeClick={node => {
          // Center view on node
          fgRef.current.centerAt(node.x, node.y, 1000);
          fgRef.current.zoom(4, 2000);
          onNodeClick(node);
        }}
        d3VelocityDecay={0.4} // Less jittery
        cooldownTicks={100}
        onEngineStop={() => fgRef.current.zoomToFit(400)}
      />
    </div>
  );
};

export default GraphVisualizer;
