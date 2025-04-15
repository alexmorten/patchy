import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { Search } from './components/Search';
import './styles/App.css';

function App() {
  return (
    <Router>
      <div className="app">
        <header className="app-header">
          <h1>Patchy</h1>
        </header>
        <main className="app-main">
          <Routes>
            <Route path="/" element={<Search />} />
            {/* Add more routes here */}
          </Routes>
        </main>
      </div>
    </Router>
  );
}

export default App;
