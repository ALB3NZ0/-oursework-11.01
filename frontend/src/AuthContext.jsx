import React, { createContext, useState, useContext, useEffect } from 'react';

const AuthContext = createContext();

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Восстанавливаем пользователя из localStorage
    const storedUser = localStorage.getItem('user');
    const token = localStorage.getItem('token');
    
    if (storedUser && token) {
      try {
        const userData = JSON.parse(storedUser);
        // Нормализуем данные на случай старых форматов
        const normalizedUser = {
          id: userData.id || userData.ID || 0,
          fullname: (userData.fullname || userData.FullName || '').trim(),
          email: (userData.email || userData.Email || '').trim(),
          role_id: userData.role_id || userData.roleId || userData.RoleID || 3
        };
        setUser(normalizedUser);
      } catch (error) {
        console.error('Error parsing user data from localStorage:', error);
        // Очищаем поврежденные данные
        localStorage.removeItem('user');
        localStorage.removeItem('token');
      }
    }
    setLoading(false);
  }, []);

  const login = (userData, token) => {
    localStorage.setItem('token', token);
    localStorage.setItem('user', JSON.stringify(userData));
    setUser(userData);
  };

  const logout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    setUser(null);
  };

  const updateUser = (userData) => {
    localStorage.setItem('user', JSON.stringify(userData));
    setUser(userData);
  };

  const isAuthenticated = () => !!user;

  return (
    <AuthContext.Provider value={{ user, login, logout, updateUser, isAuthenticated, loading }}>
      {children}
    </AuthContext.Provider>
  );
};


