import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { useState } from 'react';
import { Search } from './components/Search';
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
          <h1>Patchy</h1>
        </header>
        <main className="app-main">
          <Routes>
            <Route 
              path="/" 
              element={<Search onResultsChange={setSearchResults} />} 
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
