import React, { useState, useEffect } from 'react';
import { adminAPI } from '../../api';
import { useNotification } from '../../components/Notification';
import './AdminComponents.css';

const BrandsAdmin = () => {
  const { showSuccess, showError, confirm } = useNotification();
  const [brands, setBrands] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editingBrand, setEditingBrand] = useState(null);
  const [formData, setFormData] = useState({ brand_name: '' });

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const response = await adminAPI.brands.getAll();
      setBrands(response.data || []);
    } catch (error) {
      console.error('Ошибка загрузки:', error);
      showError('Не удалось загрузить бренды');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingBrand(null);
    setFormData({ brand_name: '' });
    setShowModal(true);
  };

  const handleEdit = (brand) => {
    setEditingBrand(brand);
    setFormData({ brand_name: brand.brand_name || '' });
    setShowModal(true);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      if (editingBrand) {
        await adminAPI.brands.update(editingBrand.id, formData);
        showSuccess('Бренд успешно обновлен');
      } else {
        await adminAPI.brands.create(formData);
        showSuccess('Бренд успешно создан');
      }
      setShowModal(false);
      loadData();
    } catch (error) {
      console.error('Ошибка сохранения:', error);
      showError(error.response?.data?.error || 'Не удалось сохранить бренд');
    }
  };

  const handleDelete = async (id) => {
    confirm(
      'Вы уверены, что хотите удалить этот бренд?',
      async () => {
        try {
          await adminAPI.brands.delete(id);
          showSuccess('Бренд успешно удален');
          loadData();
        } catch (error) {
          console.error('Ошибка удаления:', error);
          showError('Не удалось удалить бренд');
        }
      }
    );
  };

  if (loading) return <div className="loading-text">Загрузка...</div>;

  return (
    <div className="admin-section">
      <div className="section-header">
        <h2>Управление брендами</h2>
        <button className="btn-primary" onClick={handleCreate}>
          Добавить бренд
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
            {brands.length === 0 ? (
              <tr>
                <td colSpan="3" className="empty-state">Нет брендов</td>
              </tr>
            ) : (
              brands.map(brand => (
                <tr key={brand.id}>
                  <td>{brand.id}</td>
                  <td>{brand.brand_name}</td>
                  <td>
                    <button className="btn-edit" onClick={() => handleEdit(brand)} title="Редактировать">
                      Редактировать
                    </button>
                    <button className="btn-delete" onClick={() => handleDelete(brand.id)} title="Удалить">
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
            <h3>{editingBrand ? 'Редактировать бренд' : 'Создать бренд'}</h3>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label>Название бренда *</label>
                <input
                  type="text"
                  value={formData.brand_name}
                  onChange={(e) => setFormData({ brand_name: e.target.value })}
                  required
                />
              </div>
              <div className="modal-actions">
                <button type="button" className="btn-secondary" onClick={() => setShowModal(false)}>
                  Отмена
                </button>
                <button type="submit" className="btn-primary">
                  {editingBrand ? 'Сохранить' : 'Создать'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default BrandsAdmin;



