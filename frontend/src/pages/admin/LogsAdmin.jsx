import React, { useState, useEffect } from 'react';
import { adminAPI } from '../../api';
import Pagination from '../../components/Pagination';
import './AdminComponents.css';

const LogsAdmin = () => {
  const [logs, setLogs] = useState([]);
  const [allLogs, setAllLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('all');
  const [currentPage, setCurrentPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(1);

  useEffect(() => {
    loadData();
  }, [currentPage, limit]);

  const loadData = async () => {
    setLoading(true);
    try {
      const response = await adminAPI.logs.getAll(currentPage, limit);
      
      // Обработка нового формата ответа с пагинацией
      let logsData = [];
      if (response.data && response.data.data) {
        logsData = response.data.data || [];
        setTotal(response.data.total || 0);
        setTotalPages(response.data.total_pages || 1);
      } else {
        // Fallback для старого формата (массив)
        logsData = response.data || [];
        setTotal(logsData.length || 0);
        setTotalPages(1);
      }
      
      setAllLogs(logsData);
    } catch (error) {
      console.error('Ошибка загрузки:', error);
      alert('Не удалось загрузить логи');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Вы уверены, что хотите удалить этот лог?')) return;
    try {
      await adminAPI.logs.delete(id);
      alert('Лог успешно удален!');
      loadData();
    } catch (error) {
      console.error('Ошибка удаления:', error);
      alert('Не удалось удалить лог');
    }
  };

  const getActionIcon = (action) => {
    const icons = {
      CREATE: '➕',
      UPDATE: '✏️',
      DELETE: '🗑️',
      LOGIN: '🔐',
      LOGOUT: '🚪',
    };
    return icons[action] || '📋';
  };

  const filteredLogs = filter === 'all' 
    ? allLogs 
    : allLogs.filter(log => log.action === filter);

  if (loading) return <div className="loading-text">Загрузка...</div>;

  return (
    <div className="admin-section">
      <div className="section-header">
        <h2>Журнал действий</h2>
        <div className="header-actions">
          <select
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className="filter-select"
          >
            <option value="all">Все действия</option>
            <option value="CREATE">Создание</option>
            <option value="UPDATE">Обновление</option>
            <option value="DELETE">Удаление</option>
            <option value="LOGIN">Вход</option>
            <option value="LOGOUT">Выход</option>
          </select>
          <button className="btn-secondary" onClick={loadData}>
            Обновить
          </button>
        </div>
      </div>

      <div className="data-table-container">
        <table className="data-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Действие</th>
              <th>Пользователь ID</th>
              <th>Сущность</th>
              <th>ID сущности</th>
              <th>Детали</th>
              <th>Дата</th>
              <th>Действия</th>
            </tr>
          </thead>
          <tbody>
            {filteredLogs.length === 0 ? (
              <tr>
                <td colSpan="8" className="empty-state">Нет логов</td>
              </tr>
            ) : (
              filteredLogs.map(log => (
                <tr key={log.id}>
                  <td>{log.id}</td>
                  <td>
                    <span className={`action-badge action-${log.action}`}>
                      {getActionIcon(log.action)} {log.action}
                    </span>
                  </td>
                  <td>{log.user_id}</td>
                  <td>{log.entity || 'N/A'}</td>
                  <td>{log.entity_id || 'N/A'}</td>
                  <td className="details-cell">{log.details || 'Нет деталей'}</td>
                  <td>
                    {new Date(log.created_at).toLocaleString('ru-RU')}
                  </td>
                  <td>
                    <button className="btn-delete" onClick={() => handleDelete(log.id)}>
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
        onPageChange={(page) => {
          setCurrentPage(page);
        }}
        limit={limit}
        total={total}
      />
    </div>
  );
};

export default LogsAdmin;

