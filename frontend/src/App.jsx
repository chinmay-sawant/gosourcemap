import React, { useState } from 'react';
import { useGraphData } from './hooks/useGraphData';
import GraphVisualizer from './components/Graph/GraphVisualizer';
import Sidebar from './components/Sidebar/Sidebar';
import StatsPanel from './components/StatsPanel/StatsPanel';
import './App.css';

function App() {
  const { data, loading, error } = useGraphData();
  const [selectedNode, setSelectedNode] = useState(null);

  if (loading) return <div className="loading">Loading Graph Data...</div>;
  if (error) return <div className="error">Error connecting to API: {error.message}</div>;

  return (
    <div className="App">
      <GraphVisualizer 
        data={data} 
        onNodeClick={setSelectedNode} 
      />
      <StatsPanel nodes={data.nodes} />
      <Sidebar 
        node={selectedNode} 
        onClose={() => setSelectedNode(null)} 
      />
    </div>
  );
}

export default App;
