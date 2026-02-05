import React from 'react';
import './StatsPanel.css';

const StatsPanel = ({ nodes }) => {
  if (!nodes || nodes.length === 0) return null;

  const counts = nodes.reduce((acc, node) => {
    // Skip File nodes for stats? Or include?
    // Let's include everything but maybe separate generic nodes
    if (node.type === 'FILE') return acc;
    acc[node.type] = (acc[node.type] || 0) + 1;
    return acc;
  }, {});

  const total = Object.values(counts).reduce((a, b) => a + b, 0);

  return (
    <div className="stats-panel">
      <h3>Graph Stats</h3>
      <div className="stat-row total">
        <span>Total Nodes</span>
        <span>{total}</span>
      </div>
      <div className="divider"></div>
      {Object.entries(counts).map(([type, count]) => (
        <div key={type} className="stat-row">
          <span className={`dot ${type}`}></span>
          <span className="label">{type}</span>
          <span className="count">{count}</span>
        </div>
      ))}
    </div>
  );
};

export default StatsPanel;
