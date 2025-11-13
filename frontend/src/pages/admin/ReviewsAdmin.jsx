import React, { useState, useEffect } from 'react';
import { adminAPI } from '../../api';
import Pagination from '../../components/Pagination';
import './AdminComponents.css';

const ReviewsAdmin = () => {
  const [reviews, setReviews] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editingReview, setEditingReview] = useState(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(1);
  const [formData, setFormData] = useState({
    rating: 5,
    comment: '',
  });

  useEffect(() => {
    loadData();
  }, [currentPage, limit]);

  const loadData = async () => {
    setLoading(true);
    try {
      const response = await adminAPI.reviews.getAll(currentPage, limit);
      
      // Обработка нового формата ответа с пагинацией
      if (response.data && response.data.data) {
        setReviews(response.data.data || []);
        setTotal(response.data.total || 0);
        setTotalPages(response.data.total_pages || 1);
      } else {
        // Fallback для старого формата (массив)
        setReviews(response.data || []);
        setTotal(response.data?.length || 0);
        setTotalPages(1);
      }
    } catch (error) {
      console.error('Ошибка загрузки:', error);
      alert('Не удалось загрузить отзывы');
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = (review) => {
    setEditingReview(review);
    setFormData({
      rating: review.rating || 5,
      comment: review.comment || '',
    });
    setShowModal(true);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      await adminAPI.reviews.update(editingReview.id, formData);
      alert('Отзыв успешно обновлен!');
      setShowModal(false);
      loadData();
    } catch (error) {
      console.error('Ошибка сохранения:', error);
      alert(error.response?.data?.error || 'Не удалось обновить отзыв');
    }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Вы уверены, что хотите удалить этот отзыв?')) return;
    try {
      await adminAPI.reviews.delete(id);
      alert('Отзыв успешно удален!');
      loadData();
    } catch (error) {
      console.error('Ошибка удаления:', error);
      alert('Не удалось удалить отзыв');
    }
  };

  if (loading) return <div className="loading-text">Загрузка...</div>;

  return (
    <div className="admin-section">
      <div className="section-header">
        <h2>Управление отзывами</h2>
        <button className="btn-secondary" onClick={loadData}>
          Обновить
        </button>
      </div>

      <div className="data-table-container">
        <table className="data-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Товар ID</th>
              <th>Пользователь ID</th>
              <th>Рейтинг</th>
              <th>Комментарий</th>
              <th>Дата</th>
              <th>Действия</th>
            </tr>
          </thead>
          <tbody>
            {reviews.length === 0 ? (
              <tr>
                <td colSpan="7" className="empty-state">Нет отзывов</td>
              </tr>
            ) : (
              reviews.map(review => (
                <tr key={review.id}>
                  <td>{review.id}</td>
                  <td>{review.product_id}</td>
                  <td>{review.user_id}</td>
                  <td>
                    <span className="rating">
                      {'⭐'.repeat(review.rating)}
                    </span>
                  </td>
                  <td className="comment-cell">
                    {review.comment || 'Нет комментария'}
                  </td>
                  <td>
                    {new Date(review.date).toLocaleDateString('ru-RU')}
                  </td>
                  <td>
                    <button className="btn-edit" onClick={() => handleEdit(review)}>
                      ✏️
                    </button>
                    <button className="btn-delete" onClick={() => handleDelete(review.id)}>
                      🗑️
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      <Pagination
        currentPage={currentPage}
        totalPages={totalPages}
        onPageChange={(page) => setCurrentPage(page)}
        limit={limit}
        total={total}
      />

      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal-content admin-modal" onClick={(e) => e.stopPropagation()}>
            <h3>Редактировать отзыв</h3>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label>Рейтинг *</label>
                <select
                  value={formData.rating}
                  onChange={(e) => setFormData({ ...formData, rating: parseInt(e.target.value) })}
                  required
                >
                  <option value={1}>⭐ 1</option>
                  <option value={2}>⭐⭐ 2</option>
                  <option value={3}>⭐⭐⭐ 3</option>
                  <option value={4}>⭐⭐⭐⭐ 4</option>
                  <option value={5}>⭐⭐⭐⭐⭐ 5</option>
                </select>
              </div>

              <div className="form-group">
                <label>Комментарий</label>
                <textarea
                  value={formData.comment}
                  onChange={(e) => setFormData({ ...formData, comment: e.target.value })}
                  rows={5}
                  placeholder="Текст отзыва..."
                />
              </div>

              <div className="modal-actions">
                <button type="button" className="btn-secondary" onClick={() => setShowModal(false)}>
                  Отмена
                </button>
                <button type="submit" className="btn-primary">
                  Сохранить
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default ReviewsAdmin;

