import { useState, useRef, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

interface SearchResultProps {
  id: string;
  text: string;
}

export const SearchResult = ({ id, text }: SearchResultProps) => {
  const [isExpanded, setIsExpanded] = useState(false);
  const [contentHeight, setContentHeight] = useState(0);
  const contentRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();

  useEffect(() => {
    if (contentRef.current) {
      setContentHeight(contentRef.current.scrollHeight);
    }
  }, [text]);

  const toggleExpand = () => {
    setIsExpanded(!isExpanded);
  };

  return (
    <div className="result-item">
      <h3 className="result-title">
        <button 
          className="result-link"
          onClick={() => navigate(`/result/${id}`)}
        >
          {id}
        </button>
      </h3>
      <div 
        ref={contentRef}
        className={`result-content ${isExpanded ? 'expanded' : ''}`}
        style={isExpanded ? { maxHeight: `${contentHeight}px` } : {}}
        dangerouslySetInnerHTML={{ __html: text }}
      />
      <button 
        className="expand-button"
        onClick={toggleExpand}
      >
        {isExpanded ? 'Show Less' : 'Show More'}
      </button>
    </div>
  );
}; 