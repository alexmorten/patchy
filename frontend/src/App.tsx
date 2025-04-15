import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { SearchPage } from './components/SearchPage';
import { SearchResultDetail } from './components/SearchResultDetail';
import './styles/App.css';

function App() {
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
              element={<SearchPage />}
            />
            <Route 
              path="/result/:id" 
              element={<SearchResultDetail results={[]} />} 
            />
          </Routes>
        </main>
      </div>
    </Router>
  );
}

export default App;
