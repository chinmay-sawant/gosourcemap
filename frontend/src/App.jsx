import React, { useState } from 'react';
import { useGraphData } from './hooks/useGraphData';
import GraphVisualizer from './components/Graph/GraphVisualizer';
import Sidebar from './components/Sidebar/Sidebar';
import StatsPanel from './components/StatsPanel/StatsPanel';
import { Settings, RefreshCw, DownloadCloud, X } from 'lucide-react';
import './App.css';

function App() {
  const { data, loading, error, loadMore, hasMore, setLimit, limit, refresh, filters, applyFilters } = useGraphData();
  const [selectedNode, setSelectedNode] = useState(null);
  const [showControls, setShowControls] = useState(true);

  // Local state for inputs
  const [extInput, setExtInput] = useState('');
  const [dirInput, setDirInput] = useState('');

  const handleAddExt = (e) => {
    if (e.key === 'Enter' && extInput.trim()) {
        const val = extInput.trim().startsWith('.') ? extInput.trim() : `.${extInput.trim()}`;
        if (!filters.skipExt.includes(val)) {
            applyFilters({ ...filters, skipExt: [...filters.skipExt, val] });
        }
        setExtInput('');
    }
  };

  const removeExt = (ext) => {
      applyFilters({ ...filters, skipExt: filters.skipExt.filter(e => e !== ext) });
  };

  const handleAddDir = (e) => {
    if (e.key === 'Enter' && dirInput.trim()) {
        const val = dirInput.trim();
        if (!filters.skipDir.includes(val)) {
            applyFilters({ ...filters, skipDir: [...filters.skipDir, val] });
        }
        setDirInput('');
    }
  };

  const removeDir = (dir) => {
      applyFilters({ ...filters, skipDir: filters.skipDir.filter(d => d !== dir) });
  };

  if (error) return <div className="error">Error connecting to API: {error.message}</div>;

  return (
    <div className="App">
       {/* Controls */}
       <div className={`controls-panel ${showControls ? 'open' : 'closed'}`}>
          <div className="controls-header">
             <h3>Graph Controls</h3>
             <button onClick={() => setShowControls(!showControls)} className="toggle-btn">
                <Settings size={18} />
             </button>
          </div>
          
          {showControls && (
            <div className="controls-content">
                <div className="control-group">
                    <label>Batch Size: <strong>{limit}</strong></label>
                    <input 
                        type="range" 
                        min="100" 
                        max="2000" 
                        step="100"
                        value={limit} 
                        onChange={(e) => setLimit(Number(e.target.value))} 
                    />
                </div>

                <div className="control-group">
                    <label>Skip Extensions</label>
                    <div className="chip-input">
                        {filters.skipExt.map(ext => (
                            <span key={ext} className="chip">
                                {ext} <X size={12} onClick={() => removeExt(ext)} />
                            </span>
                        ))}
                        <input 
                            type="text" 
                            placeholder="e.g. .py (Enter)" 
                            value={extInput}
                            onChange={(e) => setExtInput(e.target.value)}
                            onKeyDown={handleAddExt}
                        />
                    </div>
                </div>

                <div className="control-group">
                    <label>Skip Directories</label>
                    <div className="chip-input">
                        {filters.skipDir.map(dir => (
                            <span key={dir} className="chip">
                                {dir} <X size={12} onClick={() => removeDir(dir)} />
                            </span>
                        ))}
                        <input 
                            type="text" 
                            placeholder="e.g. venv (Enter)" 
                            value={dirInput}
                            onChange={(e) => setDirInput(e.target.value)}
                            onKeyDown={handleAddDir}
                        />
                    </div>
                </div>

                <div className="actions">
                    <button onClick={() => refresh()} className="btn secondary">
                        <RefreshCw size={16} /> Reset
                    </button>
                    <button 
                        onClick={loadMore} 
                        disabled={!hasMore || loading} 
                        className="btn primary"
                    >
                        {loading ? 'Loading...' : <><DownloadCloud size={16} /> Load More</>}
                    </button>
                </div>
                
                <div className="status-text">
                    {hasMore ? 'Next batch available' : 'All available data loaded'}
                </div>
            </div>
          )}
       </div>

      {loading && data.nodes.length === 0 && <div className="loading-overlay">Loading Graph...</div>}

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
