import React, { useState, useEffect } from 'react';
import './Notification.css';

const NotificationContext = React.createContext();

export const useNotification = () => {
  const context = React.useContext(NotificationContext);
  if (!context) {
    throw new Error('useNotification must be used within NotificationProvider');
  }
  return context;
};

export const NotificationProvider = ({ children }) => {
  const [notifications, setNotifications] = useState([]);

  const showNotification = (message, type = 'info', duration = 3000) => {
    const id = Date.now() + Math.random();
    const notification = { id, message, type };
    
    setNotifications(prev => [...prev, notification]);
    
    if (duration > 0) {
      setTimeout(() => {
        removeNotification(id);
      }, duration);
    }
    
    return id;
  };

  const removeNotification = (id) => {
    setNotifications(prev => prev.filter(n => n.id !== id));
  };

  const showSuccess = (message, duration) => showNotification(message, 'success', duration);
  const showError = (message, duration) => showNotification(message, 'error', duration);
  const showInfo = (message, duration) => showNotification(message, 'info', duration);
  const showWarning = (message, duration) => showNotification(message, 'warning', duration);

  const confirm = (message, onConfirm, onCancel) => {
    const id = Date.now() + Math.random();
    const notification = {
      id,
      message,
      type: 'confirm',
      onConfirm: () => {
        if (onConfirm) onConfirm();
        removeNotification(id);
      },
      onCancel: () => {
        if (onCancel) onCancel();
        removeNotification(id);
      }
    };
    
    setNotifications(prev => [...prev, notification]);
    return id;
  };

  return (
    <NotificationContext.Provider value={{ showSuccess, showError, showInfo, showWarning, confirm }}>
      {children}
      <div className="notifications-container">
        {notifications.map(notification => (
          <NotificationItem
            key={notification.id}
            notification={notification}
            onClose={() => removeNotification(notification.id)}
          />
        ))}
      </div>
    </NotificationContext.Provider>
  );
};

const NotificationItem = ({ notification, onClose }) => {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    setTimeout(() => setIsVisible(true), 10);
  }, []);

  if (notification.type === 'confirm') {
    return (
      <div className={`notification notification-confirm ${isVisible ? 'visible' : ''}`}>
        <div className="notification-content">
          <p>{notification.message}</p>
          <div className="notification-actions">
            <button className="btn-confirm" onClick={notification.onConfirm}>
              OK
            </button>
            <button className="btn-cancel" onClick={notification.onCancel}>
              Отмена
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className={`notification notification-${notification.type} ${isVisible ? 'visible' : ''}`}>
      <div className="notification-content">
        <p>{notification.message}</p>
        <button className="notification-close" onClick={onClose}>×</button>
      </div>
    </div>
  );
};

export default NotificationProvider;





