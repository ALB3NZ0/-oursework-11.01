import React, { useState, useEffect } from 'react';
import { adminAPI } from '../../api';
import Pagination from '../../components/Pagination';
import './AdminComponents.css';

const UsersAdmin = () => {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editingUser, setEditingUser] = useState(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(1);
  const [formData, setFormData] = useState({
    fullname: '',
    email: '',
    password: '',
    role_id: '3',
  });

  useEffect(() => {
    loadData();
  }, [currentPage, limit]);

  const loadData = async () => {
    setLoading(true);
    try {
      const response = await adminAPI.users.getAll(currentPage, limit);
      
      // Обработка нового формата ответа с пагинацией
      if (response.data && response.data.data) {
        setUsers(response.data.data || []);
        setTotal(response.data.total || 0);
        setTotalPages(response.data.total_pages || 1);
      } else {
        // Fallback для старого формата (массив)
        setUsers(response.data || []);
        setTotal(response.data?.length || 0);
        setTotalPages(1);
      }
    } catch (error) {
      console.error('Ошибка загрузки:', error);
      alert('Не удалось загрузить пользователей');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingUser(null);
    setFormData({
      fullname: '',
      email: '',
      password: '',
      role_id: '3',
    });
    setShowModal(true);
  };

  const handleEdit = (user) => {
    setEditingUser(user);
    setFormData({
      fullname: user.fullname || '',
      email: user.email || '',
      password: '',
      role_id: user.role_id?.toString() || '3',
    });
    setShowModal(true);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      const submitData = {
        ...formData,
        role_id: parseInt(formData.role_id),
      };
      
      if (!editingUser) {
        if (!formData.password) {
          alert('Пароль обязателен для нового пользователя');
          return;
        }
      } else {
        // При редактировании удаляем password если он пустой
        if (!formData.password) {
          delete submitData.password;
        }
      }

      if (editingUser) {
        await adminAPI.users.update(editingUser.id, submitData);
        alert('Пользователь успешно обновлен!');
      } else {
        await adminAPI.users.create(submitData);
        alert('Пользователь успешно создан!');
      }
      setShowModal(false);
      loadData();
    } catch (error) {
      console.error('Ошибка сохранения:', error);
      alert(error.response?.data?.error || 'Не удалось сохранить пользователя');
    }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Вы уверены, что хотите удалить этого пользователя?')) return;
    try {
      await adminAPI.users.delete(id);
      alert('Пользователь успешно удален!');
      loadData();
    } catch (error) {
      console.error('Ошибка удаления:', error);
      alert('Не удалось удалить пользователя');
    }
  };

  const getRoleName = (roleId) => {
    const roles = { 1: 'Админ', 2: 'Менеджер', 3: 'Пользователь' };
    return roles[roleId] || 'Неизвестно';
  };

  if (loading) return <div className="loading-text">Загрузка...</div>;

  return (
    <div className="admin-section">
      <div className="section-header">
        <h2>Управление пользователями</h2>
        <div className="header-actions">
          <button className="btn-primary" onClick={handleCreate}>
            Добавить пользователя
          </button>
        </div>
      </div>

      <div className="data-table-container">
        <table className="data-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Имя</th>
              <th>Email</th>
              <th>Роль</th>
              <th>Действия</th>
            </tr>
          </thead>
          <tbody>
            {users.length === 0 ? (
              <tr>
                <td colSpan="5" className="empty-state">Нет пользователей</td>
              </tr>
            ) : (
              users.map(user => (
                <tr key={user.id}>
                  <td>{user.id}</td>
                  <td>{user.fullname}</td>
                  <td>{user.email}</td>
                  <td>
                    <span className={`role-badge role-${user.role_id}`}>
                      {getRoleName(user.role_id)}
                    </span>
                  </td>
                  <td>
                    <button className="btn-edit" onClick={() => handleEdit(user)}>
                      ✏️
                    </button>
                    <button className="btn-delete" onClick={() => handleDelete(user.id)}>
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
            <h3>{editingUser ? 'Редактировать пользователя' : 'Создать пользователя'}</h3>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label>Полное имя *</label>
                <input
                  type="text"
                  value={formData.fullname}
                  onChange={(e) => setFormData({ ...formData, fullname: e.target.value })}
                  required
                />
              </div>

              <div className="form-group">
                <label>Email *</label>
                <input
                  type="email"
                  value={formData.email}
                  onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  required
                />
              </div>

              <div className="form-group">
                <label>Пароль {editingUser ? '(оставьте пустым для сохранения текущего)' : '*'}</label>
                <input
                  type="password"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  required={!editingUser}
                  minLength={8}
                />
              </div>

              <div className="form-group">
                <label>Роль *</label>
                <select
                  value={formData.role_id}
                  onChange={(e) => setFormData({ ...formData, role_id: e.target.value })}
                  required
                >
                  <option value="1">Админ</option>
                  <option value="2">Менеджер</option>
                  <option value="3">Пользователь</option>
                </select>
              </div>

              <div className="modal-actions">
                <button type="button" className="btn-secondary" onClick={() => setShowModal(false)}>
                  Отмена
                </button>
                <button type="submit" className="btn-primary">
                  {editingUser ? 'Сохранить' : 'Создать'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default UsersAdmin;



