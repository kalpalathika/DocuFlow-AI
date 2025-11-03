import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import UploadPage from './pages/UploadPage';
import ChatPage from './pages/ChatPage';

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<UploadPage />} />
        <Route path="/chat/:sessionId" element={<ChatPage />} />
        {/* More routes will be added here */}
      </Routes>
    </Router>
  );
}

export default App;
