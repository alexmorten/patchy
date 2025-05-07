import { SearchResult } from './SearchResult';
import '../styles/Search.css';

interface SearchResultData {
  id: string;
  text: string;
  url: string;
}

interface SearchProps {
  results: SearchResultData[];
  query: string;
  onQueryChange: (query: string) => void;
  isLoading: boolean;
  error: string | null;
  autoFocus?: boolean;
}

export const Search = ({ 
  results, 
  query, 
  onQueryChange,
  isLoading,
  error,
  autoFocus = false
}: SearchProps) => {
  const onSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    // Don't make a separate API call from the form submit
    // The parent SearchPage component is already handling API calls via useEffect
  };

  return (
    <div className="search-container">
      <form onSubmit={onSubmit} className="search-form">
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
              text={result.text}
            />
        ))}
      </div>
    </div>
  );
}; 