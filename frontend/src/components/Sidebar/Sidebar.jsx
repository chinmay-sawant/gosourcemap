import React, { useState, useMemo } from 'react';
import { X, ExternalLink, ChevronDown } from 'lucide-react';
import './Sidebar.css';

// Editor configurations with their URI schemes
const EDITORS = [
  { id: 'antigravity', name: 'Antigravity', scheme: 'antigravity://file' },
  { id: 'vscode', name: 'VSCode', scheme: 'vscode://file' },
  { id: 'cursor', name: 'Cursor', scheme: 'cursor://file' },
  { id: 'windsurf', name: 'Windsurf', scheme: 'windsurf://file' },
];

// Detect platform: WSL2, Linux, or Windows
const detectPlatform = () => {
  const userAgent = navigator.userAgent.toLowerCase();
  const platform = navigator.platform.toLowerCase();
  
  // Check if running in WSL2 (Linux but accessed from Windows)
  // WSL2 detection: check if path starts with /home or /mnt and user agent contains Windows
  if (platform.includes('linux') && userAgent.includes('windows')) {
    return 'wsl2';
  }
  if (platform.includes('win')) {
    return 'windows';
  }
  return 'linux';
};

// Convert Linux path to WSL2 Windows-compatible path
// /home/chinmay/... -> \\wsl.localhost\Ubuntu\home\chinmay\...
const convertToWSL2Path = (linuxPath) => {
  if (!linuxPath) return linuxPath;
  // Remove leading slash and replace remaining slashes with backslashes
  const pathWithoutLeadingSlash = linuxPath.startsWith('/') ? linuxPath.substring(1) : linuxPath;
  return `\\\\wsl.localhost\\Ubuntu\\${pathWithoutLeadingSlash.replace(/\//g, '\\\\')}`;
};

const Sidebar = ({ node, onClose }) => {
  const [selectedEditor, setSelectedEditor] = useState(EDITORS[0]); // Default to Antigravity
  const [showDropdown, setShowDropdown] = useState(false);

  // Detect platform once
  const platform = useMemo(() => detectPlatform(), []);

  if (!node) return null;

  const data = node.original || node; // Handle File nodes vs CodeNodes

  // Get the appropriate file path based on platform
  const getFilePath = () => {
    if (!data.file_path) return null;
    
    // On WSL2, convert Linux paths to Windows-compatible WSL paths
    if (platform === 'wsl2') {
      return convertToWSL2Path(data.file_path);
    }
    
    return data.file_path;
  };

  // Generate the URI for opening the file in the selected editor
  const getEditorUri = () => {
    const filePath = getFilePath();
    if (!filePath) return null;
    const lineNum = data.line_number > 0 ? data.line_number : 1;
    // Format: scheme://file/path:line
    return `${selectedEditor.scheme}${filePath}:${lineNum}`;
  };

  // Handle opening the file in the selected editor
  const handleOpenInEditor = () => {
    const uri = getEditorUri();
    if (uri) {
      window.open(uri, '_self');
    }
  };

  // Handle editor selection
  const handleEditorSelect = (editor) => {
    setSelectedEditor(editor);
    setShowDropdown(false);
  };

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
            <label>Open In Editor</label>
            <div className="editor-open-section">
              {/* Editor dropdown */}
              <div className="editor-dropdown">
                <button 
                  className="editor-dropdown-btn"
                  onClick={() => setShowDropdown(!showDropdown)}
                >
                  {selectedEditor.name}
                  <ChevronDown size={14} />
                </button>
                {showDropdown && (
                  <div className="editor-dropdown-menu">
                    {EDITORS.map(editor => (
                      <button
                        key={editor.id}
                        className={`editor-option ${editor.id === selectedEditor.id ? 'active' : ''}`}
                        onClick={() => handleEditorSelect(editor)}
                      >
                        {editor.name}
                      </button>
                    ))}
                  </div>
                )}
              </div>
              {/* Open button */}
              <button 
                className="open-file-btn"
                onClick={handleOpenInEditor}
                title={`Open in ${selectedEditor.name}`}
              >
                <ExternalLink size={14} />
                Open
              </button>
            </div>
            {/* Clickable file path */}
            <div 
              className="value file-path clickable"
              onClick={handleOpenInEditor}
              title={`Click to open in ${selectedEditor.name}`}
            >
              {data.file_path}{data.line_number > 0 ? `:${data.line_number}` : ''}
            </div>
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

