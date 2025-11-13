import React, { useState, useEffect } from 'react';
import { adminAPI, brandsAPI, categoriesAPI } from '../../api';
import { useNotification } from '../../components/Notification';
import Pagination from '../../components/Pagination';
import './AdminComponents.css';

const ProductsAdmin = () => {
  const { showSuccess, showError, showInfo, confirm } = useNotification();
  const [products, setProducts] = useState([]);
  const [brands, setBrands] = useState([]);
  const [categories, setCategories] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editingProduct, setEditingProduct] = useState(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(1);
  const [formData, setFormData] = useState({
    name: '',
    image_url: '',
    price: '',
    brand_id: '',
    category_id: '',
  });
  const [formErrors, setFormErrors] = useState({});

  useEffect(() => {
    loadData();
  }, [currentPage, limit]);

  const loadData = async () => {
    setLoading(true);
    try {
      const [productsRes, brandsRes, categoriesRes] = await Promise.all([
        adminAPI.products.getAll(currentPage, limit),
        brandsAPI.getAll(),
        categoriesAPI.getAll(),
      ]);
      
      // Обработка нового формата ответа с пагинацией
      if (productsRes.data && productsRes.data.data) {
        setProducts(productsRes.data.data || []);
        setTotal(productsRes.data.total || 0);
        setTotalPages(productsRes.data.total_pages || 1);
      } else {
        // Fallback для старого формата (массив)
        setProducts(productsRes.data || []);
        setTotal(productsRes.data?.length || 0);
        setTotalPages(1);
      }
      
      setBrands(brandsRes.data || []);
      setCategories(categoriesRes.data || []);
    } catch (error) {
      console.error('Ошибка загрузки данных:', error);
      showError('Не удалось загрузить данные');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingProduct(null);
    setFormData({
      name: '',
      image_url: '',
      price: '',
      brand_id: '',
      category_id: '',
    });
    setFormErrors({});
    setShowModal(true);
  };

  const handleEdit = (product) => {
    setEditingProduct(product);
    setFormData({
      name: product.name || '',
      image_url: product.image_url || '',
      price: product.price || '',
      brand_id: product.brand_id || '',
      category_id: product.category_id || '',
    });
    setFormErrors({});
    setShowModal(true);
  };

  const validateForm = () => {
    const errors = {};
    
    if (!formData.name.trim()) {
      errors.name = 'Название обязательно';
    }
    
    if (!formData.image_url.trim()) {
      errors.image_url = 'Фото обязательно';
    } else {
      // Проверка на валидный URL
      try {
        const urls = formData.image_url.split(',').map(url => url.trim());
        urls.forEach(url => {
          new URL(url);
        });
      } catch (e) {
        errors.image_url = 'Неверный формат URL';
      }
    }
    
    if (!formData.price || parseFloat(formData.price) <= 0) {
      errors.price = 'Цена должна быть больше 0';
    }
    
    if (!formData.brand_id) {
      errors.brand_id = 'Бренд обязателен';
    }
    
    if (!formData.category_id) {
      errors.category_id = 'Категория обязательна';
    }
    
    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!validateForm()) {
      showError('Пожалуйста, исправьте ошибки в форме');
      return;
    }
    
    try {
      const submitData = {
        ...formData,
        price: parseFloat(formData.price),
        brand_id: parseInt(formData.brand_id),
        category_id: parseInt(formData.category_id),
      };

      if (editingProduct) {
        await adminAPI.products.update(editingProduct.id, submitData);
        showSuccess('Товар успешно обновлен');
      } else {
        await adminAPI.products.create(submitData);
        showSuccess('Товар успешно создан');
      }
      
      setShowModal(false);
      setFormErrors({});
      loadData();
    } catch (error) {
      console.error('Ошибка сохранения:', error);
      const errorMessage = error.response?.data || error.message || 'Не удалось сохранить товар';
      showError(errorMessage);
    }
  };

  const handleDelete = async (id) => {
    confirm(
      'Вы уверены, что хотите удалить этот товар? Это действие нельзя отменить.',
      async () => {
        try {
          await adminAPI.products.delete(id);
          showSuccess('Товар успешно удален');
          loadData();
        } catch (error) {
          console.error('Ошибка удаления:', error);
          const errorMessage = error.response?.data?.message || error.response?.data || error.message || 'Не удалось удалить товар';
          showError(`Ошибка удаления товара: ${errorMessage}`);
        }
      }
    );
  };


  if (loading) {
    return <div className="loading-text">Загрузка...</div>;
  }

  return (
    <div className="admin-section">
      <div className="section-header">
        <h2>Управление товарами</h2>
        <div className="header-actions">
          <button className="btn-primary" onClick={handleCreate}>
            Добавить товар
          </button>
        </div>
      </div>

      <div className="data-table-container">
        <table className="data-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Изображение</th>
              <th>Название</th>
              <th>Цена</th>
              <th>Бренд</th>
              <th>Категория</th>
              <th>Действия</th>
            </tr>
          </thead>
          <tbody>
            {products.length === 0 ? (
              <tr>
                <td colSpan="7" className="empty-state">Нет товаров</td>
              </tr>
            ) : (
              products.map(product => {
                const brand = brands.find(b => b.id === product.brand_id);
                const category = categories.find(c => c.id === product.category_id);
                
                return (
                  <tr key={product.id}>
                    <td>{product.id}</td>
                    <td>
                      {product.image_url ? (
                        <img src={product.image_url.split(',')[0].trim()} alt={product.name} className="table-image" />
                      ) : (
                        <span>Нет фото</span>
                      )}
                    </td>
                    <td>{product.name}</td>
                    <td>{product.price} ₽</td>
                    <td>{brand?.brand_name || 'N/A'}</td>
                    <td>{category?.category_name || 'N/A'}</td>
                    <td>
                      <button className="btn-edit" onClick={() => handleEdit(product)} title="Редактировать">
                        Редактировать
                      </button>
                      <button className="btn-delete" onClick={() => handleDelete(product.id)} title="Удалить">
                        Удалить
                      </button>
                    </td>
                  </tr>
                );
              })
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
            <h3>{editingProduct ? 'Редактировать товар' : 'Создать товар'}</h3>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label>Название *</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => {
                    setFormData({ ...formData, name: e.target.value });
                    if (formErrors.name) {
                      setFormErrors({ ...formErrors, name: '' });
                    }
                  }}
                  required
                  className={formErrors.name ? 'error' : ''}
                />
                {formErrors.name && <span className="error-message">{formErrors.name}</span>}
              </div>

              <div className="form-group">
                <label>Фото (через запятую) *</label>
                <input
                  type="url"
                  value={formData.image_url}
                  onChange={(e) => {
                    setFormData({ ...formData, image_url: e.target.value });
                    if (formErrors.image_url) {
                      setFormErrors({ ...formErrors, image_url: '' });
                    }
                  }}
                  placeholder="https://site.com/1.jpg, https://site.com/2.jpg"
                  required
                  className={formErrors.image_url ? 'error' : ''}
                />
                {formErrors.image_url && <span className="error-message">{formErrors.image_url}</span>}
              </div>

              <div className="form-group">
                <label>Цена *</label>
                <input
                  type="number"
                  step="0.01"
                  min="0"
                  value={formData.price}
                  onChange={(e) => {
                    setFormData({ ...formData, price: e.target.value });
                    if (formErrors.price) {
                      setFormErrors({ ...formErrors, price: '' });
                    }
                  }}
                  required
                  className={formErrors.price ? 'error' : ''}
                />
                {formErrors.price && <span className="error-message">{formErrors.price}</span>}
              </div>

              <div className="form-group">
                <label>Бренд *</label>
                <select
                  value={formData.brand_id}
                  onChange={(e) => {
                    setFormData({ ...formData, brand_id: e.target.value });
                    if (formErrors.brand_id) {
                      setFormErrors({ ...formErrors, brand_id: '' });
                    }
                  }}
                  required
                  className={formErrors.brand_id ? 'error' : ''}
                >
                  <option value="">Выберите бренд</option>
                  {brands.map(brand => (
                    <option key={brand.id} value={brand.id}>
                      {brand.brand_name}
                    </option>
                  ))}
                </select>
                {formErrors.brand_id && <span className="error-message">{formErrors.brand_id}</span>}
              </div>

              <div className="form-group">
                <label>Категория *</label>
                <select
                  value={formData.category_id}
                  onChange={(e) => {
                    setFormData({ ...formData, category_id: e.target.value });
                    if (formErrors.category_id) {
                      setFormErrors({ ...formErrors, category_id: '' });
                    }
                  }}
                  required
                  className={formErrors.category_id ? 'error' : ''}
                >
                  <option value="">Выберите категорию</option>
                  {categories.map(category => (
                    <option key={category.id} value={category.id}>
                      {category.category_name}
                    </option>
                  ))}
                </select>
                {formErrors.category_id && <span className="error-message">{formErrors.category_id}</span>}
              </div>

              <div className="modal-actions">
                <button type="button" className="btn-secondary" onClick={() => setShowModal(false)}>
                  Отмена
                </button>
                <button type="submit" className="btn-primary">
                  {editingProduct ? 'Сохранить' : 'Создать'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default ProductsAdmin;



