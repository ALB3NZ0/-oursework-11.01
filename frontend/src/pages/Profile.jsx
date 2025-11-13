import React, { useState, useEffect } from 'react';
import { useAuth } from '../AuthContext';
import { useNotification } from '../components/Notification';
import { userAPI, passwordAPI, ordersAPI } from '../api';
import './Profile.css';

const Profile = () => {
  const { user, updateUser } = useAuth();
  const { showSuccess, showError } = useNotification();
  const [activeTab, setActiveTab] = useState('profile');
  const [loading, setLoading] = useState(false);
  
  // Profile data - use correct field names from backend
  const [profileData, setProfileData] = useState({
    fullname: '',
    email: '',
  });

  // Update profile data when user changes
  useEffect(() => {
    if (user) {
      setProfileData({
        fullname: (user.fullname || '').trim(),
        email: (user.email || '').trim(),
      });
    }
  }, [user]);
  
  // Password data
  const [passwordData, setPasswordData] = useState({
    oldPassword: '',
    newPassword: '',
    confirmPassword: '',
  });
  
  // Confirmation code
  const [showConfirmModal, setShowConfirmModal] = useState(false);
  const [confirmationCode, setConfirmationCode] = useState('');
  const [confirmLoading, setConfirmLoading] = useState(false);

  // Orders history
  const [orders, setOrders] = useState([]);
  const [ordersLoading, setOrdersLoading] = useState(false);
  const [orderDetails, setOrderDetails] = useState({});

  // Load orders when orders tab is active
  useEffect(() => {
    if (activeTab === 'orders' && user?.id) {
      loadOrders().catch(err => console.error('Ошибка загрузки заказов:', err));
    }
  }, [activeTab, user?.id]);

  const loadOrders = async () => {
    if (!user?.id) {
      console.error('User ID is not available');
      setOrders([]);
      setOrderDetails({});
      return;
    }

    setOrdersLoading(true);
    try {
      console.log('🔄 Loading orders for user ID:', user.id);
      const response = await ordersAPI.getByUserId(user.id);
      console.log('📦 Orders response:', response);
      console.log('📦 Orders response.data:', response.data);
      
      // Обработка нового формата ответа с пагинацией
      let loadedOrders = [];
      if (response.data) {
        if (response.data.data && Array.isArray(response.data.data)) {
          // Новый формат с пагинацией
          loadedOrders = response.data.data;
          console.log('✅ Используется формат с пагинацией. Заказов:', loadedOrders.length);
        } else if (Array.isArray(response.data)) {
          // Старый формат (массив)
          loadedOrders = response.data;
          console.log('✅ Используется старый формат (массив). Заказов:', loadedOrders.length);
        } else {
          // Неожиданный формат
          console.warn('⚠️ Неожиданный формат ответа:', response.data);
          loadedOrders = [];
        }
      } else {
        console.warn('⚠️ response.data отсутствует');
        loadedOrders = [];
      }
      
      console.log('📦 Orders loaded:', loadedOrders);
      setOrders(loadedOrders);
      
      // Load order details for each order
      const details = {};
      if (loadedOrders.length > 0) {
        for (const order of loadedOrders) {
          try {
            console.log('🔍 Loading details for order:', order.id);
            const detailsResponse = await ordersAPI.getProductsByOrderId(order.id);
            console.log('📋 Order details response:', detailsResponse);
            details[order.id] = Array.isArray(detailsResponse.data) ? detailsResponse.data : [];
            console.log('✅ Details loaded for order', order.id, ':', details[order.id].length, 'items');
          } catch (error) {
            console.error('❌ Ошибка загрузки деталей заказа', order.id, ':', error);
            console.error('Error response:', error.response);
            details[order.id] = [];
          }
        }
      }
      console.log('📊 All order details:', details);
      setOrderDetails(details);
    } catch (error) {
      console.error('❌ Ошибка загрузки заказов:', error);
      console.error('Error response:', error.response);
      console.error('Error message:', error.message);
      
      // Показываем пользователю ошибку только если это не 404 (нет заказов)
      if (error.response?.status !== 404) {
        showError('Не удалось загрузить заказы. Попробуйте обновить страницу.');
      }
      
      setOrders([]);
      setOrderDetails({});
    } finally {
      setOrdersLoading(false);
    }
  };

  const handleUpdateProfile = async (e) => {
    e.preventDefault();
    setLoading(true);

    try {
      await userAPI.update(user.id, {
        fullname: profileData.fullname,
        email: profileData.email,
      });
      
      // Update user in context and localStorage
      const updatedUser = { ...user, ...profileData };
      updateUser(updatedUser);
      
      showSuccess('Профиль успешно обновлен');
    } catch (error) {
      console.error('Ошибка обновления профиля:', error);
      showError('Не удалось обновить профиль');
    } finally {
      setLoading(false);
    }
  };

  const handleChangePassword = async (e) => {
    e.preventDefault();
    
    if (passwordData.newPassword !== passwordData.confirmPassword) {
      showError('Пароли не совпадают');
      return;
    }

    if (passwordData.newPassword.length < 8) {
      showError('Пароль должен содержать минимум 8 символов');
      return;
    }

    setLoading(true);

    try {
      await passwordAPI.change({
        old_password: passwordData.oldPassword,
        new_password: passwordData.newPassword,
      });
      
      // Open confirmation modal
      setShowConfirmModal(true);
      showSuccess('Код подтверждения отправлен на ваш email');
    } catch (error) {
      console.error('Ошибка смены пароля:', error);
      showError(error.response?.data?.error || 'Не удалось отправить запрос на смену пароля');
    } finally {
      setLoading(false);
    }
  };

  const handleConfirmPasswordChange = async () => {
    setConfirmLoading(true);

    try {
      await passwordAPI.confirmChange({
        code: confirmationCode,
      });
      
      showSuccess('Пароль успешно изменен');
      setShowConfirmModal(false);
      setConfirmationCode('');
      setPasswordData({
        oldPassword: '',
        newPassword: '',
        confirmPassword: '',
      });
    } catch (error) {
      console.error('Ошибка подтверждения смены пароля:', error);
      showError(error.response?.data?.error || 'Неверный код подтверждения');
    } finally {
      setConfirmLoading(false);
    }
  };

  return (
    <div className="profile-container">
      <h1 className="page-title">Мой профиль</h1>

      <div className="profile-tabs">
        <button
          className={activeTab === 'profile' ? 'tab-active' : 'tab'}
          onClick={() => setActiveTab('profile')}
        >
          Данные профиля
        </button>
        <button
          className={activeTab === 'password' ? 'tab-active' : 'tab'}
          onClick={() => setActiveTab('password')}
        >
          Смена пароля
        </button>
        <button
          className={activeTab === 'orders' ? 'tab-active' : 'tab'}
          onClick={() => setActiveTab('orders')}
        >
          История заказов
        </button>
      </div>

      <div className="profile-content">
        {activeTab === 'profile' && (
          <div className="profile-form-card">
            <h2>Редактировать профиль</h2>
            <form onSubmit={handleUpdateProfile}>
              <div className="form-group">
                <label>Полное имя</label>
                <input
                  type="text"
                  value={profileData.fullname}
                  onChange={(e) => setProfileData({ ...profileData, fullname: e.target.value })}
                  required
                  placeholder="Иван Иванов"
                />
              </div>

              <div className="form-group">
                <label>Email</label>
                <input
                  type="email"
                  value={profileData.email}
                  onChange={(e) => setProfileData({ ...profileData, email: e.target.value })}
                  required
                  placeholder="user@example.com"
                />
              </div>

              <button type="submit" className="btn-primary" disabled={loading}>
                {loading ? 'Сохранение...' : 'Сохранить изменения'}
              </button>
            </form>
          </div>
        )}
        
        {activeTab === 'password' && (
          <div className="profile-form-card">
            <h2>Сменить пароль</h2>
            <p className="password-info">
              При смене пароля на ваш email будет отправлен код подтверждения.
            </p>
            <form onSubmit={handleChangePassword}>
              <div className="form-group">
                <label>Текущий пароль</label>
                <input
                  type="password"
                  value={passwordData.oldPassword}
                  onChange={(e) => setPasswordData({ ...passwordData, oldPassword: e.target.value })}
                  required
                  placeholder="••••••••"
                />
              </div>

              <div className="form-group">
                <label>Новый пароль</label>
                <input
                  type="password"
                  value={passwordData.newPassword}
                  onChange={(e) => setPasswordData({ ...passwordData, newPassword: e.target.value })}
                  required
                  placeholder="Минимум 8 символов"
                  minLength={8}
                />
              </div>

              <div className="form-group">
                <label>Подтвердите новый пароль</label>
                <input
                  type="password"
                  value={passwordData.confirmPassword}
                  onChange={(e) => setPasswordData({ ...passwordData, confirmPassword: e.target.value })}
                  required
                  placeholder="••••••••"
                />
              </div>

              <button type="submit" className="btn-primary" disabled={loading}>
                {loading ? 'Отправка...' : 'Сменить пароль'}
              </button>
            </form>
          </div>
        )}

        {activeTab === 'orders' && (
          <div className="profile-form-card">
            <h2>История заказов</h2>
            
            {ordersLoading ? (
              <div className="loading-text">Загрузка заказов...</div>
            ) : !orders || orders.length === 0 ? (
              <div className="empty-state">
                <p>У вас пока нет заказов</p>
              </div>
            ) : (
              <div className="orders-list">
                {orders.map((order) => {
                  const details = orderDetails[order.id] || [];
                  const totalAmount = details.reduce((sum, item) => sum + (item.price * item.quantity), 0);
                  const totalItems = details.reduce((sum, item) => sum + item.quantity, 0);
                  
                  // Форматирование даты с проверкой
                  let orderDateStr = 'Дата не указана';
                  if (order.order_date) {
                    try {
                      const orderDate = new Date(order.order_date);
                      if (!isNaN(orderDate.getTime())) {
                        orderDateStr = orderDate.toLocaleDateString('ru-RU', {
                          year: 'numeric',
                          month: 'long',
                          day: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit'
                        });
                      }
                    } catch (e) {
                      console.error('Ошибка форматирования даты:', e);
                    }
                  }
                  
                  return (
                    <div key={order.id} className="order-card">
                      <div className="order-header">
                        <div className="order-info">
                          <h3>Заказ #{order.id}</h3>
                          <p className="order-date">
                            {orderDateStr}
                          </p>
                        </div>
                        <div className="order-status">
                          <span className="status-badge">Обработан</span>
                        </div>
                      </div>
                      
                      <div className="order-details">
                        {details.length > 0 ? (
                          <>
                            <div className="order-items">
                              {details.map((item) => (
                                <div key={item.id} className="order-item">
                                  {item.image_url && (
                                    <img src={item.image_url.split(',')[0].trim()} alt={item.product_name || 'Товар'} className="order-item-image" />
                                  )}
                                  <div className="order-item-info">
                                    <h4>{item.product_name || 'Товар'}</h4>
                                    <p>Размер: {item.size || 'N/A'} • Кол-во: {item.quantity || 0} шт</p>
                                  </div>
                                  <div className="order-item-price">
                                    {item.price && item.quantity ? (item.price * item.quantity).toFixed(2) : '0.00'} ₽
                                  </div>
                                </div>
                              ))}
                            </div>
                            <div className="order-summary">
                              <div className="order-summary-item">
                                <span>Товаров:</span>
                                <span>{totalItems} шт</span>
                              </div>
                              <div className="order-summary-item order-total">
                                <span>Итого:</span>
                                <span>{totalAmount.toFixed(2)} ₽</span>
                              </div>
                            </div>
                          </>
                        ) : (
                          <div className="order-empty">
                            <p>Детали заказа загружаются или заказ пуст</p>
                          </div>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Confirmation Modal */}
      {showConfirmModal && (
        <div className="modal-overlay" onClick={() => setShowConfirmModal(false)}>
          <div className="modal-content confirm-modal" onClick={(e) => e.stopPropagation()}>
            <h3>Подтвердите смену пароля</h3>
            <p>Введите код подтверждения, отправленный на ваш email</p>
            
            <div className="form-group">
              <input
                type="text"
                value={confirmationCode}
                onChange={(e) => setConfirmationCode(e.target.value)}
                placeholder="Введите 6-значный код"
                maxLength={6}
                className="code-input"
              />
            </div>

            <div className="modal-actions">
              <button
                onClick={() => setShowConfirmModal(false)}
                className="btn-secondary"
                disabled={confirmLoading}
              >
                Отмена
              </button>
              <button
                onClick={handleConfirmPasswordChange}
                className="btn-primary"
                disabled={confirmationCode.length !== 6 || confirmLoading}
              >
                {confirmLoading ? 'Подтверждение...' : 'Подтвердить'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Profile;

