import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './AuthContext';
import { ThemeProvider } from './ThemeContext';
import { NotificationProvider } from './components/Notification';
import Navigation from './components/Navigation';
import Footer from './components/Footer';
import ThemeToggle from './components/ThemeToggle';
import Login from './pages/Login';
import Register from './pages/Register';
import Catalog from './pages/Catalog';
import Basket from './pages/Basket';
import Favorites from './pages/Favorites';
import Profile from './pages/Profile';
import Support from './pages/Support';
import ManagerPanel from './pages/ManagerPanel';
import AdminPanel from './pages/AdminPanel';
import About from './pages/About';
import './App.css';

const PrivateRoute = ({ children }) => {
  const { isAuthenticated, loading } = useAuth();
  
  if (loading) {
    return (
      <div className="loading-screen">
        <div className="spinner"></div>
      </div>
    );
  }
  
  return isAuthenticated() ? children : <Navigate to="/login" />;
};

// Компонент для главной страницы - перенаправляет в зависимости от роли
const HomeRoute = () => {
  const { isAuthenticated, loading, user } = useAuth();
  
  if (loading) {
    return (
      <div className="loading-screen">
        <div className="spinner"></div>
      </div>
    );
  }
  
  if (isAuthenticated()) {
    // Админы идут на админ-панель
    if (user?.role_id === 1) {
      return <Navigate to="/admin" replace />;
    }
    // Менеджеры идут на панель менеджера
    if (user?.role_id === 2) {
      return <Navigate to="/manager" replace />;
    }
  }
  
  // Обычные пользователи и неавторизованные видят каталог
  return <Catalog />;
};

const PublicRoute = ({ children }) => {
  const { isAuthenticated, loading, user } = useAuth();
  
  if (loading) {
    return (
      <div className="loading-screen">
        <div className="spinner"></div>
      </div>
    );
  }
  
  if (isAuthenticated()) {
    // Перенаправляем авторизованных пользователей в зависимости от роли
    if (user?.role_id === 1) {
      return <Navigate to="/admin" />;
    }
    if (user?.role_id === 2) {
      return <Navigate to="/manager" />;
    }
    // Обычные пользователи идут на каталог
    return <Navigate to="/" />;
  }
  
  return children;
};

function App() {
  return (
    <ThemeProvider>
      <NotificationProvider>
        <AuthProvider>
          <Router>
            <div className="App">
              <Navigation />
              <main className="main-content">
                <Routes>
                  <Route path="/login" element={<PublicRoute><Login /></PublicRoute>} />
                  <Route path="/register" element={<PublicRoute><Register /></PublicRoute>} />
                  <Route path="/about" element={<About />} />
                  <Route path="/" element={<HomeRoute />} />
                  <Route 
                    path="/basket" 
                    element={<PrivateRoute><Basket /></PrivateRoute>} 
                  />
                  <Route 
                    path="/favorites" 
                    element={<PrivateRoute><Favorites /></PrivateRoute>} 
                  />
                  <Route 
                    path="/profile" 
                    element={<PrivateRoute><Profile /></PrivateRoute>} 
                  />
                  <Route 
                    path="/support" 
                    element={<Support />} 
                  />
                  <Route 
                    path="/manager" 
                    element={<PrivateRoute><ManagerPanel /></PrivateRoute>} 
                  />
                  <Route 
                    path="/admin" 
                    element={<PrivateRoute><AdminPanel /></PrivateRoute>} 
                  />
                </Routes>
              </main>
              <Footer />
              <ThemeToggle />
            </div>
          </Router>
        </AuthProvider>
      </NotificationProvider>
    </ThemeProvider>
  );
}

export default App;


