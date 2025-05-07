import { useParams, useNavigate } from 'react-router-dom';
import { useEffect, useState, useCallback, useRef } from 'react';
import { getResult } from '../services/api';
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
  const [result, setResult] = useState<SearchResultData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Use refs to track previous values and avoid unnecessary fetches
  const prevIdRef = useRef(id);
  const prevResultsRef = useRef(results);
  
  const fetchResult = useCallback(async () => {
    if (!id) return;
    
    // First try to find the result in the search results
    const searchResult = results.find(r => r.id === id);
    if (searchResult) {
      setResult(searchResult);
      setIsLoading(false);
      return;
    }
    
    // If not found in search results, fetch from the API
    try {
      const data = await getResult(id);
      setResult(data);
    } catch (err) {
      setError('Failed to fetch result');
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  }, [id, results]);

  useEffect(() => {
    // Only fetch if id or results changed meaningfully
    const idChanged = id !== prevIdRef.current;
    const resultsChanged = results !== prevResultsRef.current;
    
    if (idChanged || resultsChanged) {
      setIsLoading(true);
      fetchResult();
      
      // Update refs with current values
      prevIdRef.current = id;
      prevResultsRef.current = results;
    }
  }, [id, results, fetchResult]);

  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        navigate(-1);
      }
    };

    window.addEventListener('keydown', handleEscape);
    return () => window.removeEventListener('keydown', handleEscape);
  }, [navigate]);

  if (isLoading) {
    return (
      <div className="result-detail-overlay">
        <div className="result-detail-container">
          <div className="result-detail-loading">Loading...</div>
        </div>
      </div>
    );
  }

  if (error || !result) {
    return (
      <div className="result-detail-overlay">
        <div className="result-detail-container">
          <div className="result-detail-error">
            {error || 'Result not found'}
            <button 
              className="result-detail-close"
              onClick={() => navigate(-1)}
            >
              Close
            </button>
          </div>
        </div>
      </div>
    );
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
          /* Prevent reflow by using a key */
          key={result.id}
        />
      </div>
    </div>
  );
}; 