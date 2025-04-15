import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { useState } from 'react';
import { SearchPage } from './components/SearchPage';
import { SearchResultDetail } from './components/SearchResultDetail';
import './styles/App.css';

interface SearchResultData {
  id: string;
  text: string;
  url: string;
}

function App() {
  const [searchResults, setSearchResults] = useState<SearchResultData[]>([]);

  return (
    <Router>
      <div className="app">
        <header className="app-header">
          <h1>Patchy Search</h1>
        </header>
        <main className="app-main">
          <Routes>
            <Route 
              path="/" 
              element={
                <SearchPage 
                  searchResults={searchResults}
                  onSearchResultsChange={setSearchResults}
                />
              }
            />
            <Route 
              path="/result/:id" 
              element={<SearchResultDetail results={searchResults} />} 
            />
          </Routes>
        </main>
      </div>
    </Router>
  );
}

export default App;
