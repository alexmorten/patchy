import { useParams, useNavigate } from 'react-router-dom';
import { useEffect } from 'react';
import '../styles/SearchResultDetail.css';

interface SearchResultData {
  id: string;
  text: string;
  url: string;
}

interface SearchResultDetailProps {
  results: SearchResultData[];
}

export const SearchResultDetail = ({ results }: SearchResultDetailProps) => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const result = results.find(r => r.id === id);

  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        navigate(-1);
      }
    };

    window.addEventListener('keydown', handleEscape);
    return () => window.removeEventListener('keydown', handleEscape);
  }, [navigate]);

  if (!result) {
    return null;
  }

  return (
    <div className="result-detail-overlay" onClick={() => navigate(-1)}>
      <div className="result-detail-container" onClick={e => e.stopPropagation()}>
        <div className="result-detail-header">
          <h1 className="result-detail-title">{result.id}</h1>
          <button 
            className="result-detail-close"
            onClick={() => navigate(-1)}
          >
            Close
          </button>
        </div>
        <div 
          className="result-detail-content"
          dangerouslySetInnerHTML={{ __html: result.text }}
        />
      </div>
    </div>
  );
}; 