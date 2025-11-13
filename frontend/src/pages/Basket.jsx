import React, { useState, useEffect, useCallback } from 'react';
import { useAuth } from '../AuthContext';
import { useNotification } from '../components/Notification';
import { basketAPI, ordersAPI } from '../api';
import './Basket.css';

const Basket = () => {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [ordering, setOrdering] = useState(false);
  const { user } = useAuth();
  const { showSuccess, showError, confirm } = useNotification();

  const loadBasket = useCallback(async () => {
    if (!user?.id) return;
    
    try {
      const response = await basketAPI.getByUserId(user.id);
      setItems(response.data || []);
    } catch (error) {
      console.error('Ошибка загрузки корзины:', error);
    } finally {
      setLoading(false);
    }
  }, [user]);

  useEffect(() => {
    loadBasket();
  }, [loadBasket]);

  const updateQuantity = async (id, newQuantity) => {
    if (newQuantity <= 0) {
      return deleteItem(id);
    }

    try {
      await basketAPI.update(id, { quantity: newQuantity });
      loadBasket();
    } catch (error) {
      console.error('Ошибка обновления:', error);
      showError('Не удалось обновить количество');
    }
  };

  const deleteItem = async (id) => {
    try {
      await basketAPI.delete(id);
      loadBasket();
    } catch (error) {
      console.error('Ошибка удаления:', error);
      showError('Не удалось удалить товар');
    }
  };

  const calculateTotal = () => {
    return items.reduce((total, item) => {
      const price = item.price || 0;
      return total + (price * item.quantity);
    }, 0).toFixed(2);
  };

  const handleOrder = async () => {
    if (items.length === 0) {
      showError('Корзина пуста');
      return;
    }

    confirm(
      'Подтвердите оформление заказа',
      async () => {
        await processOrder();
      }
    );
  };

  const processOrder = async () => {

    setOrdering(true);
    try {
      // Создаем заказ
      const orderRes = await ordersAPI.create({ user_id: user.id });
      const orderId = orderRes.data.id;

      // Добавляем товары в заказ
      for (const item of items) {
        await ordersAPI.addProduct({
          order_id: orderId,
          product_size_id: item.product_size_id,
          quantity: item.quantity,
        });
      }

      // Очищаем корзину
      for (const item of items) {
        await basketAPI.delete(item.id);
      }

      showSuccess('Заказ успешно оформлен! Проверьте почту.');
      loadBasket();
    } catch (error) {
      console.error('Ошибка оформления заказа:', error);
      showError('Не удалось оформить заказ');
    } finally {
      setOrdering(false);
    }
  };

  if (loading) {
    return (
      <div className="page-loading">
        <div className="spinner"></div>
      </div>
    );
  }

  return (
    <div className="basket-container">
      <h1 className="page-title">Корзина</h1>

      {items.length === 0 ? (
        <div className="empty-state">
          <p>Ваша корзина пуста</p>
        </div>
      ) : (
        <>
          <div className="basket-items">
            {items.map(item => (
              <div key={item.id} className="basket-item">
                {item.image_url ? (
                  <img
                    src={item.image_url}
                    alt={item.product_name}
                    className="basket-item-image"
                  />
                ) : (
                  <div className="basket-item-placeholder">
                    <span>Нет фото</span>
                  </div>
                )}
                
                <div className="basket-item-info">
                  <h3>{item.product_name || 'Товар'}</h3>
                  <p>Размер: {item.size}</p>
                  <p className="item-price">
                    {item.price ? `${item.price} ₽` : '0 ₽'} × {item.quantity}
                  </p>
                </div>

                <div className="basket-item-actions">
                  <div className="quantity-control">
                    <button
                      onClick={() => updateQuantity(item.id, item.quantity - 1)}
                      className="qty-btn"
                    >
                      −
                    </button>
                    <span className="quantity">{item.quantity}</span>
                    <button
                      onClick={() => updateQuantity(item.id, item.quantity + 1)}
                      className="qty-btn"
                    >
                      +
                    </button>
                  </div>
                  <button
                    onClick={() => deleteItem(item.id)}
                    className="delete-btn"
                  >
                    Удалить
                  </button>
                </div>
              </div>
            ))}
          </div>

          <div className="basket-total">
            <h2>Итого: {calculateTotal()} ₽</h2>
            <button 
              className="order-btn" 
              onClick={handleOrder}
              disabled={ordering}
            >
              {ordering ? 'Оформление...' : 'Оформить заказ'}
            </button>
          </div>
        </>
      )}
    </div>
  );
};

export default Basket;

