import { useState, useEffect, useCallback, useMemo } from 'react';
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
  const initialQuery = useMemo(() => searchParams.get('q') || '', [searchParams]);
  const [query, setQuery] = useState(initialQuery);
  const [debouncedQuery] = useDebounce(query, 200);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  const updateSearchParams = useCallback((newQuery: string) => {
    const params = new URLSearchParams();
    if (newQuery) {
      params.set('q', newQuery);
    }
    setSearchParams(params);
  }, [setSearchParams]);
  
  const [debouncedSetSearchParams] = useDebounce(updateSearchParams, 300);

  const performSearch = useCallback(async (query: string) => {
    if (!query.trim() || query.trim().length < 3) {
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const data = await search(query);
      onSearchResultsChange(data);
    } catch (err) {
      setError('Failed to fetch search results');
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  }, [onSearchResultsChange]);

  useEffect(() => {
    performSearch(debouncedQuery);
  }, [debouncedQuery, performSearch]);

  const handleSearch = useCallback((newQuery: string) => {
    setQuery(newQuery);
    debouncedSetSearchParams(newQuery);
  }, [debouncedSetSearchParams]);

  return <Search 
    results={searchResults}
    query={query}
    onQueryChange={handleSearch}
    isLoading={isLoading}
    error={error}
    autoFocus={true}
  />;
} 