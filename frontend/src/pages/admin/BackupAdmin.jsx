import React, { useState, useEffect, useRef } from 'react';
import { adminAPI } from '../../api';
import { useNotification } from '../../components/Notification';
import './AdminComponents.css';

const BackupAdmin = () => {
  const { showSuccess, showError, confirm } = useNotification();
  const [backups, setBackups] = useState([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [restoring, setRestoring] = useState(false);
  const fileInputRef = useRef(null);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const response = await adminAPI.backup.getInfo();
      setBackups(response.data?.backup_files || []);
    } catch (error) {
      console.error('Ошибка загрузки:', error);
      showError('Не удалось загрузить информацию о бэкапах');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async () => {
    confirm(
      'Создать новый бэкап базы данных?',
      async () => {
        setCreating(true);
        try {
          await adminAPI.backup.create();
          showSuccess('Бэкап успешно создан!');
          loadData();
        } catch (error) {
          console.error('Ошибка создания:', error);
          showError(error.response?.data?.error || 'Не удалось создать бэкап');
        } finally {
          setCreating(false);
        }
      }
    );
  };

  const handleDelete = async (filename) => {
    confirm(
      `Удалить бэкап ${filename}?`,
      async () => {
        try {
          await adminAPI.backup.delete(filename);
          showSuccess('Бэкап успешно удален!');
          loadData();
        } catch (error) {
          console.error('Ошибка удаления:', error);
          showError('Не удалось удалить бэкап');
        }
      }
    );
  };

  const handleDownload = async (filename) => {
    try {
      const response = await adminAPI.backup.download(filename);
      
      // Создаем blob и скачиваем файл
      const blob = new Blob([response.data], { type: 'application/sql' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
      
      showSuccess('Бэкап успешно скачан!');
    } catch (error) {
      console.error('Ошибка скачивания:', error);
      showError('Не удалось скачать бэкап');
    }
  };

  const handleImport = async (event) => {
    const file = event.target.files[0];
    if (!file) return;

    if (!file.name.endsWith('.sql')) {
      showError('Файл должен иметь расширение .sql');
      return;
    }

    confirm(
      'Восстановить базу данных из этого бэкапа? Все текущие данные будут заменены!',
      async () => {
        setRestoring(true);
        try {
          await adminAPI.backup.restore(file);
          showSuccess('База данных успешно восстановлена из бэкапа!');
          loadData();
        } catch (error) {
          console.error('Ошибка восстановления:', error);
          const errorMessage = error.response?.data?.message || error.response?.data || error.message || 'Не удалось восстановить базу данных';
          showError(`Ошибка восстановления: ${errorMessage}`);
        } finally {
          setRestoring(false);
          // Очищаем input
          if (fileInputRef.current) {
            fileInputRef.current.value = '';
          }
        }
      }
    );
  };

  const formatFileSize = (bytes) => {
    if (!bytes) return 'N/A';
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
  };

  if (loading) return <div className="loading-text">Загрузка...</div>;

  return (
    <div className="admin-section">
      <div className="section-header">
        <h2>Бэкапы</h2>
        <div className="header-actions">
          <label className="btn-secondary" title="Импорт бэкапа" style={{ cursor: 'pointer', margin: 0 }}>
            {restoring ? 'Восстановление...' : '📤 Импорт'}
            <input
              ref={fileInputRef}
              type="file"
              accept=".sql"
              onChange={handleImport}
              disabled={restoring}
              style={{ display: 'none' }}
            />
          </label>
          <button
            className="btn-primary"
            onClick={handleCreate}
            disabled={creating}
          >
            {creating ? 'Создание...' : 'Создать бэкап'}
          </button>
        </div>
      </div>

      <div className="backups-list">
        {backups.length === 0 ? (
          <div className="empty-state">
            <p>Бэкапов пока нет</p>
            <p className="empty-hint">Нажмите "Создать бэкап" для создания первого бэкапа базы данных</p>
          </div>
        ) : (
          backups.map((backup, index) => (
            <div key={index} className="backup-item">
              <div className="backup-info">
                <h3>{backup.filename}</h3>
                <div className="backup-details">
                  <span>📁 {backup.path}</span>
                  <span>📦 {formatFileSize(backup.size_bytes)}</span>
                  <span>📅 {new Date(backup.created).toLocaleString('ru-RU')}</span>
                </div>
              </div>
              <div className="backup-actions">
                <button
                  className="btn-secondary"
                  onClick={() => handleDownload(backup.filename)}
                  title="Скачать бэкап"
                >
                  📥 Экспорт
                </button>
                <button
                  className="btn-delete"
                  onClick={() => handleDelete(backup.filename)}
                  title="Удалить бэкап"
                >
                  🗑️ Удалить
                </button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default BackupAdmin;



