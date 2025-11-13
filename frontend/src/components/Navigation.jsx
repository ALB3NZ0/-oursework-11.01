import React, { useState, useEffect, useRef } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../AuthContext';
import './Navigation.css';

const Navigation = () => {
  const { user, logout, isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const menuRef = useRef(null);

  const handleLogout = () => {
    logout();
    navigate('/login');
    setMobileMenuOpen(false);
  };

  const toggleMobileMenu = () => {
    setMobileMenuOpen(!mobileMenuOpen);
  };

  // Закрытие меню при клике вне его или на затемнение
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (mobileMenuOpen) {
        // Закрываем при клике вне меню или на затемнение
        if (!menuRef.current?.contains(event.target) && 
            !event.target.closest('.mobile-menu-toggle')) {
          setMobileMenuOpen(false);
        }
      }
    };

    // Обработчик для затемнения (клик вне меню)
    const handleBackdropClick = (event) => {
      if (mobileMenuOpen && !menuRef.current?.contains(event.target) && 
          !event.target.closest('.mobile-menu-toggle')) {
        setMobileMenuOpen(false);
      }
    };

    if (mobileMenuOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      document.addEventListener('click', handleBackdropClick);
      // Предотвращаем прокрутку страницы при открытом меню
      document.body.style.overflow = 'hidden';
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('click', handleBackdropClick);
      document.body.style.overflow = '';
    };
  }, [mobileMenuOpen]);

  return (
    <nav className="navbar">
      <div className="nav-container">
        <Link to="/about" className="nav-logo" onClick={() => setMobileMenuOpen(false)}>ShoesStore</Link>

        <button className="mobile-menu-toggle" onClick={toggleMobileMenu}>
          <span></span>
          <span></span>
          <span></span>
        </button>

        <div ref={menuRef} className={`nav-menu ${mobileMenuOpen ? 'nav-menu-open' : ''}`}>
          {isAuthenticated() ? (
            <>
              {/* Для админа показываем только админ-панель */}
              {user && user.role_id === 1 ? (
                <>
                  <Link to="/admin" className="nav-link nav-link-admin" onClick={() => setMobileMenuOpen(false)}>
                    Админ-панель
                  </Link>
                  <span className="nav-user">
                    {user?.fullname || user?.email}
                  </span>
                  <button onClick={handleLogout} className="nav-logout">
                    Выйти
                  </button>
                </>
              ) : user && user.role_id === 2 ? (
                <>
                  {/* Для менеджера показываем только панель менеджера */}
                  <Link to="/manager" className="nav-link nav-link-manager" onClick={() => setMobileMenuOpen(false)}>
                    Панель менеджера
                  </Link>
                  <span className="nav-user">
                    {user?.fullname || user?.email}
                  </span>
                  <button onClick={handleLogout} className="nav-logout">
                    Выйти
                  </button>
                </>
              ) : (
                <>
                  {/* Для обычных пользователей */}
              <Link to="/" className="nav-link" onClick={() => setMobileMenuOpen(false)}>Каталог</Link>
              <Link to="/favorites" className="nav-link" onClick={() => setMobileMenuOpen(false)}>Избранное</Link>
              <Link to="/basket" className="nav-link" onClick={() => setMobileMenuOpen(false)}>Корзина</Link>
              <Link to="/profile" className="nav-link" onClick={() => setMobileMenuOpen(false)}>Профиль</Link>
              <Link to="/support" className="nav-link" onClick={() => setMobileMenuOpen(false)}>Поддержка</Link>

                  <span className="nav-user">
                    {user?.fullname || user?.email}
                  </span>

                  <button onClick={handleLogout} className="nav-logout">
                    Выйти
                  </button>
                </>
              )}
            </>
          ) : (
            <>
              <Link to="/" className="nav-link" onClick={() => setMobileMenuOpen(false)}>Каталог</Link>
              <Link to="/support" className="nav-link" onClick={() => setMobileMenuOpen(false)}>Поддержка</Link>
              <Link to="/login" className="nav-link" onClick={() => setMobileMenuOpen(false)}>Войти</Link>
              <Link to="/register" className="nav-button" onClick={() => setMobileMenuOpen(false)}>Регистрация</Link>
            </>
          )}
        </div>
      </div>
    </nav>
  );
};

export default Navigation;
