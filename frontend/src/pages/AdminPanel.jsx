import React, { useState, useEffect } from 'react';
import { useAuth } from '../AuthContext';
import ProductsAdmin from './admin/ProductsAdmin';
import BrandsAdmin from './admin/BrandsAdmin';
import CategoriesAdmin from './admin/CategoriesAdmin';
import UsersAdmin from './admin/UsersAdmin';
import OrdersAdmin from './admin/OrdersAdmin';
import ReviewsAdmin from './admin/ReviewsAdmin';
import LogsAdmin from './admin/LogsAdmin';
import BackupAdmin from './admin/BackupAdmin';
import './AdminPanel.css';

const AdminPanel = () => {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState('products');

  const tabs = [
    { id: 'products', label: 'Товары', key: '1' },
    { id: 'brands', label: 'Бренды', key: '2' },
    { id: 'categories', label: 'Категории', key: '3' },
    { id: 'users', label: 'Пользователи', key: '4' },
    { id: 'orders', label: 'Заказы', key: '5' },
    { id: 'reviews', label: 'Отзывы', key: '6' },
    { id: 'logs', label: 'Логи', key: '7' },
    { id: 'backup', label: 'Бэкапы', key: '8' },
  ];

  // Горячие клавиши для переключения вкладок
  useEffect(() => {
    const handleKeyPress = (e) => {
      // Проверяем, что пользователь не вводит текст в поле ввода
      if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA' || e.target.isContentEditable) {
        return;
      }

      // Переключение по цифрам 1-8
      const digit = e.key;
      if (digit >= '1' && digit <= '8') {
        const tabIndex = parseInt(digit) - 1;
        if (tabs[tabIndex]) {
          setActiveTab(tabs[tabIndex].id);
        }
      }
    };

    window.addEventListener('keydown', handleKeyPress);
    return () => {
      window.removeEventListener('keydown', handleKeyPress);
    };
  }, [tabs]);

  // Проверка что пользователь админ
  const isAdmin = user && user.role_id === 1;

  if (!isAdmin) {
    return (
      <div className="admin-panel-container">
        <div className="access-denied">
          <h2>Доступ запрещен</h2>
          <p>У вас нет прав для доступа к админ-панели.</p>
          <p>Требуется роль администратора.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="admin-panel-container">
      <div className="admin-header">
        <h1 className="page-title">Админ-панель</h1>
        <div className="admin-info">
          <span className="admin-badge">Администратор</span>
          <span className="admin-user">{user?.fullname || user?.email}</span>
        </div>
      </div>

      <div className="admin-tabs">
        {tabs.map(tab => (
          <button
            key={tab.id}
            className={`admin-tab ${activeTab === tab.id ? 'active' : ''}`}
            onClick={() => setActiveTab(tab.id)}
            title={`${tab.label} (${tab.key})`}
          >
            <span className="tab-label">{tab.label}</span>
          </button>
        ))}
      </div>

      <div className="admin-content">
        {activeTab === 'products' && <ProductsAdmin />}
        {activeTab === 'brands' && <BrandsAdmin />}
        {activeTab === 'categories' && <CategoriesAdmin />}
        {activeTab === 'users' && <UsersAdmin />}
        {activeTab === 'orders' && <OrdersAdmin />}
        {activeTab === 'reviews' && <ReviewsAdmin />}
        {activeTab === 'logs' && <LogsAdmin />}
        {activeTab === 'backup' && <BackupAdmin />}
      </div>
    </div>
  );
};

export default AdminPanel;









