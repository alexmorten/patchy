import { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useDebounce } from 'use-debounce';
import { Search } from './Search';
import { search } from '../services/api';
import '../styles/Search.css';

interface SearchResultData {
  id: string;
  text: string;
  url: string;
}

export function SearchPage() {
  const [searchResults, setSearchResults] = useState<SearchResultData[]>([]);
  const [searchParams, setSearchParams] = useSearchParams();
  const initialQuery = searchParams.get('q') || '';
  const [query, setQuery] = useState(initialQuery);
  const [debouncedQuery] = useDebounce(query, 200);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [debouncedSetSearchParams] = useDebounce((newQuery: string) => {
    const params = new URLSearchParams();
    if (newQuery) {
      params.set('q', newQuery);
    }
    setSearchParams(params);
  }, 300);

  useEffect(() => {
    const performSearch = async () => {
      if (!debouncedQuery.trim() || debouncedQuery.trim().length < 3) {
        return;
      }

      setIsLoading(true);
      setError(null);

      try {
        const data = await search(debouncedQuery);
        setSearchResults(data);
      } catch (err) {
        setError('Failed to fetch search results');
        console.error(err);
      } finally {
        setIsLoading(false);
      }
    };

    performSearch();
  }, [debouncedQuery]);

  const handleSearch = (newQuery: string) => {
    setQuery(newQuery);
    debouncedSetSearchParams(newQuery);
  };

  return (
    <Search 
      results={searchResults}
      onResultsChange={setSearchResults}
      query={query}
      onQueryChange={handleSearch}
      isLoading={isLoading}
      error={error}
    />
  );
} 