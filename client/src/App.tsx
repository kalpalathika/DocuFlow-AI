import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import UploadPage from './pages/UploadPage';
import ChatPage from './pages/ChatPage';
import CompletePage from './pages/CompletePage';

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<UploadPage />} />
        <Route path="/chat/:sessionId" element={<ChatPage />} />
        <Route path="/complete/:sessionId" element={<CompletePage />} />
      </Routes>
    </Router>
  );
}

export default App;
