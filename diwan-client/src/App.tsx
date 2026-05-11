import { Routes, Route, Navigate } from 'react-router-dom';
import { useStore } from './store';
import Login from './pages/Login';
import Register from './pages/Register';
import Workspace from './pages/Workspace';

function App() {
  const token = useStore((s) => s.token);

  return (
    <Routes>
      <Route path="/login" element={!token ? <Login /> : <Navigate to="/" />} />
      <Route path="/register" element={!token ? <Register /> : <Navigate to="/" />} />
      <Route path="/*" element={token ? <Workspace /> : <Navigate to="/login" />} />
    </Routes>
  );
}

export default App;
