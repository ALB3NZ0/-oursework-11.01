import React from 'react';
import { Link } from 'react-router-dom';
import './Footer.css';

const Footer = () => {
  return (
    <footer className="footer">
      <div className="footer-container">
        <div className="footer-content">
          <div className="footer-section">
            <h3 className="footer-title">ShoesStore</h3>
            <p className="footer-text">Лучший выбор обуви для всей семьи</p>
          </div>
          
          <div className="footer-section">
            <h4 className="footer-heading">Навигация</h4>
            <ul className="footer-links">
              <li><Link to="/">Каталог</Link></li>
              <li><Link to="/about">О нас</Link></li>
              <li><Link to="/support">Поддержка</Link></li>
            </ul>
          </div>
          
          <div className="footer-section">
            <h4 className="footer-heading">Контакты</h4>
            <ul className="footer-links">
              <li>Email: info@shoesstore.ru</li>
              <li>Телефон: +7 (800) 123-45-67</li>
              <li>Адрес: г. Москва, ул. Примерная, д. 1</li>
            </ul>
          </div>
          
          <div className="footer-section">
            <h4 className="footer-heading">Режим работы</h4>
            <ul className="footer-links">
              <li>Пн-Пт: 9:00 - 21:00</li>
              <li>Сб-Вс: 10:00 - 20:00</li>
            </ul>
          </div>
        </div>
        
        <div className="footer-bottom">
          <p className="footer-copyright">
            © {new Date().getFullYear()} ShoesStore. Все права защищены.
          </p>
        </div>
      </div>
    </footer>
  );
};

export default Footer;

