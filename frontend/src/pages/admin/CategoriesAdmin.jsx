import React, { useState, useEffect } from 'react';
import { adminAPI } from '../../api';
import { useNotification } from '../../components/Notification';
import './AdminComponents.css';

const CategoriesAdmin = () => {
  const { showSuccess, showError, confirm } = useNotification();
  const [categories, setCategories] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editingCategory, setEditingCategory] = useState(null);
  const [formData, setFormData] = useState({ category_name: '' });

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const response = await adminAPI.categories.getAll();
      setCategories(response.data || []);
    } catch (error) {
      console.error('Ошибка загрузки:', error);
      showError('Не удалось загрузить категории');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingCategory(null);
    setFormData({ category_name: '' });
    setShowModal(true);
  };

  const handleEdit = (category) => {
    setEditingCategory(category);
    setFormData({ category_name: category.category_name || '' });
    setShowModal(true);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      if (editingCategory) {
        await adminAPI.categories.update(editingCategory.id, formData);
        showSuccess('Категория успешно обновлена');
      } else {
        await adminAPI.categories.create(formData);
        showSuccess('Категория успешно создана');
      }
      setShowModal(false);
      loadData();
    } catch (error) {
      console.error('Ошибка сохранения:', error);
      showError(error.response?.data?.error || 'Не удалось сохранить категорию');
    }
  };

  const handleDelete = async (id) => {
    confirm(
      'Вы уверены, что хотите удалить эту категорию?',
      async () => {
        try {
          await adminAPI.categories.delete(id);
          showSuccess('Категория успешно удалена');
          loadData();
        } catch (error) {
          console.error('Ошибка удаления:', error);
          showError('Не удалось удалить категорию');
        }
      }
    );
  };

  if (loading) return <div className="loading-text">Загрузка...</div>;

  return (
    <div className="admin-section">
      <div className="section-header">
        <h2>Управление категориями</h2>
        <button className="btn-primary" onClick={handleCreate}>
          Добавить категорию
        </button>
      </div>

      <div className="data-table-container">
        <table className="data-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Название</th>
              <th>Действия</th>
            </tr>
          </thead>
          <tbody>
            {categories.length === 0 ? (
              <tr>
                <td colSpan="3" className="empty-state">Нет категорий</td>
              </tr>
            ) : (
              categories.map(category => (
                <tr key={category.id}>
                  <td>{category.id}</td>
                  <td>{category.category_name}</td>
                  <td>
                    <button className="btn-edit" onClick={() => handleEdit(category)} title="Редактировать">
                      Редактировать
                    </button>
                    <button className="btn-delete" onClick={() => handleDelete(category.id)} title="Удалить">
                      Удалить
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal-content admin-modal" onClick={(e) => e.stopPropagation()}>
            <h3>{editingCategory ? 'Редактировать категорию' : 'Создать категорию'}</h3>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label>Название категории *</label>
                <input
                  type="text"
                  value={formData.category_name}
                  onChange={(e) => setFormData({ category_name: e.target.value })}
                  required
                />
              </div>
              <div className="modal-actions">
                <button type="button" className="btn-secondary" onClick={() => setShowModal(false)}>
                  Отмена
                </button>
                <button type="submit" className="btn-primary">
                  {editingCategory ? 'Сохранить' : 'Создать'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default CategoriesAdmin;



