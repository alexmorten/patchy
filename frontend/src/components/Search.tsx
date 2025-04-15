import { search } from '../services/api';
import { SearchResult } from './SearchResult';
import '../styles/Search.css';

interface SearchResultData {
  id: string;
  text: string;
  url: string;
}

interface SearchProps {
  results: SearchResultData[];
  onResultsChange: (results: SearchResultData[]) => void;
  query: string;
  onQueryChange: (query: string) => void;
  isLoading: boolean;
  error: string | null;
  autoFocus?: boolean;
}

const highlightText = (text: string, query: string) => {
  if (!query) return text;
  const regex = new RegExp(`(${query})`, 'gi');
  return text.replace(regex, '<mark>$1</mark>');
};

export const Search = ({ 
  results, 
  onResultsChange, 
  query, 
  onQueryChange,
  isLoading,
  error,
  autoFocus = false
}: SearchProps) => {
  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim()) return;

    try {
      const data = await search(query);
      onResultsChange(data);
    } catch (err) {
      console.error(err);
    }
  };

  return (
    <div className="search-container">
      <form onSubmit={handleSearch} className="search-form">
        <input
          type="text"
          value={query}
          onChange={(e) => onQueryChange(e.target.value)}
          placeholder="Search..."
          className="search-input"
          autoFocus={autoFocus}
        />
        <button type="submit" className="search-button" disabled={isLoading}>
          {isLoading ? 'Searching...' : 'Search'}
        </button>
      </form>

      {error && <div className="error-message">{error}</div>}

      <div className="results-container">
        {results.map((result) => (
          <SearchResult
            key={result.id}
            id={result.id}
            text={highlightText(result.text, query)}
          />
        ))}
      </div>
    </div>
  );
}; 