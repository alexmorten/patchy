import { useState } from 'react';
import { search } from '../services/api';
import '../styles/Search.css';

interface SearchResult {
  id: string;
  text: string;
  url: string;
}

const highlightText = (text: string, query: string) => {
  if (!query) return text;
  const regex = new RegExp(`(${query})`, 'gi');
  return text.replace(regex, '<mark>$1</mark>');
};

export const Search = () => {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim()) return;

    setIsLoading(true);
    setError(null);

    try {
      const data = await search(query);
      setResults(data);
    } catch (err) {
      setError('Failed to fetch search results');
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="search-container">
      <form onSubmit={handleSearch} className="search-form">
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search..."
          className="search-input"
        />
        <button type="submit" className="search-button" disabled={isLoading}>
          {isLoading ? 'Searching...' : 'Search'}
        </button>
      </form>

      {error && <div className="error-message">{error}</div>}

      <div className="results-container">
        {results.map((result) => (
          <div key={result.id} className="result-item">
            <h3 className="result-title">
              <a href={result.url} target="_blank" rel="noopener noreferrer">
                {result.id}
              </a>
            </h3>
            <p 
              className="result-content" 
              dangerouslySetInnerHTML={{ __html: highlightText(result.text, query) }} 
            />
          </div>
        ))}
      </div>
    </div>
  );
}; 