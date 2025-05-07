import { useState, useRef, useEffect, memo } from 'react';
import { useNavigate } from 'react-router-dom';

interface SearchResultProps {
  id: string;
  text: string;
}

export const SearchResult = memo(({ id, text }: SearchResultProps) => {
  const [isExpanded, setIsExpanded] = useState(false);
  const [contentHeight, setContentHeight] = useState(0);
  const contentRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();
  
  useEffect(() => {
    // Only update content height when the component mounts
    // or when isExpanded changes
    if (contentRef.current) {
      setContentHeight(contentRef.current.scrollHeight);
    }
  }, [isExpanded]);

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
});