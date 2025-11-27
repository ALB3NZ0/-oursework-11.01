import React from 'react';
import { useTheme } from '../ThemeContext';
import './ThemeToggle.css';

const ThemeToggle = () => {
  const { theme, toggleTheme } = useTheme();

  return (
    <button 
      onClick={toggleTheme} 
      className="theme-toggle-fixed"
      aria-label={theme === 'light' ? 'Переключить на темную тему' : 'Переключить на светлую тему'}
      title={theme === 'light' ? 'Темная тема' : 'Светлая тема'}
    >
      {theme === 'light' ? '🌙' : '☀️'}
    </button>
  );
};

export default ThemeToggle;















