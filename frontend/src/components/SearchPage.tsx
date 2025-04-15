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

interface SearchPageProps {
  searchResults: SearchResultData[];
  onSearchResultsChange: (results: SearchResultData[]) => void;
}

export function SearchPage({ searchResults, onSearchResultsChange }: SearchPageProps) {
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
        onSearchResultsChange(data);
      } catch (err) {
        setError('Failed to fetch search results');
        console.error(err);
      } finally {
        setIsLoading(false);
      }
    };

    performSearch();
  }, [debouncedQuery, onSearchResultsChange]);

  const handleSearch = (newQuery: string) => {
    setQuery(newQuery);
    debouncedSetSearchParams(newQuery);
  };

  return (
    <Search 
      results={searchResults}
      onResultsChange={onSearchResultsChange}
      query={query}
      onQueryChange={handleSearch}
      isLoading={isLoading}
      error={error}
      autoFocus
    />
  );
} 