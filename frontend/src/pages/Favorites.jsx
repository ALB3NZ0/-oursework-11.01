import React, { useState, useEffect, useCallback } from 'react';
import { useAuth } from '../AuthContext';
import { useNotification } from '../components/Notification';
import { favoritesAPI, basketAPI } from '../api';
import './Favorites.css';

const Favorites = () => {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const { user } = useAuth();
  const { showSuccess, showError } = useNotification();

  const loadFavorites = useCallback(async () => {
    if (!user?.id) return;
    
    try {
      const response = await favoritesAPI.getByUserId(user.id);
      setItems(response.data || []);
    } catch (error) {
      console.error('Ошибка загрузки избранного:', error);
    } finally {
      setLoading(false);
    }
  }, [user]);

  useEffect(() => {
    loadFavorites();
  }, [loadFavorites]);

  const addToBasket = async (favoriteId, productSizeId) => {
    try {
      await basketAPI.add({
        user_id: user.id,
        product_size_id: productSizeId,
        quantity: 1,
      });
      showSuccess('Товар добавлен в корзину');
    } catch (error) {
      console.error('Ошибка добавления в корзину:', error);
      showError('Не удалось добавить товар в корзину');
    }
  };

  const deleteItem = async (id) => {
    try {
      await favoritesAPI.delete(id);
      loadFavorites();
    } catch (error) {
      console.error('Ошибка удаления:', error);
      showError('Не удалось удалить товар из избранного');
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
    <div className="favorites-container">
      <h1 className="page-title">Избранное</h1>

      {items.length === 0 ? (
        <div className="empty-state">
          <p>У вас пока нет избранных товаров</p>
        </div>
      ) : (
        <div className="favorites-grid">
          {items.map(item => (
            <div key={item.id} className="favorite-item">
              {item.image_url ? (
                <img
                  src={item.image_url}
                  alt={item.product_name}
                  className="favorite-image-img"
                />
              ) : (
                <div className="favorite-image">
                  <span>Нет фото</span>
                </div>
              )}
              
              <h3>{item.product_name}</h3>
              <p>Размер: {item.size}</p>
              <p className="item-price">{item.price || 0} ₽</p>
              
              <div className="favorite-actions">
                <button
                  onClick={() => addToBasket(item.id, item.product_size_id)}
                  className="favorite-add-to-basket-btn"
                >
                  В корзину
                </button>
                <button
                  onClick={() => deleteItem(item.id)}
                  className="remove-btn"
                >
                  Удалить
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default Favorites;

