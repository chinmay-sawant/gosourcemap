import React from 'react';
import { X, Code, FileText, Activity } from 'lucide-react';
import './Sidebar.css';

const Sidebar = ({ node, onClose }) => {
  if (!node) return null;

  const data = node.original || node; // Handle File nodes vs CodeNodes

  return (
    <div className="sidebar">
      <div className="sidebar-header">
        <h2>Details</h2>
        <button onClick={onClose} className="close-btn">
          <X size={20} />
        </button>
      </div>

      <div className="sidebar-content">
        <div className="info-group">
          <label>Name</label>
          <div className="value primary">{data.name}</div>
        </div>

        <div className="info-group">
          <label>Type</label>
          <div className="value tag" data-type={node.type}>{node.type}</div>
        </div>

        {data.file_path && (
          <div className="info-group">
            <label>File Path</label>
            <div className="value file-path">{data.file_path}</div>
          </div>
        )}

        {data.line_number > 0 && (
          <div className="info-group">
            <label>Line Number</label>
            <div className="value">{data.line_number}</div>
          </div>
        )}

        {data.comments && data.comments.length > 0 && (
          <div className="info-group">
            <label>Comments</label>
            <pre className="comments-block">
              {data.comments.join('\n')}
            </pre>
          </div>
        )}
      </div>
    </div>
  );
};

export default Sidebar;
