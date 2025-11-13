import React, { useState } from 'react';
import { useAuth } from '../AuthContext';
import './Support.css';
import api from '../api';

const Support = () => {
  const { user } = useAuth();
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    message: '',
  });
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState('');

  // Автозаполнение если пользователь залогинен
  React.useEffect(() => {
    if (user) {
      setFormData({
        name: user.fullname || '',
        email: user.email || '',
        message: '',
      });
    }
  }, [user]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    setSuccess(false);

    try {
      const response = await api.post('/support', formData, {
        headers: {
          'Content-Type': 'application/json; charset=utf-8',
        },
      });
      
      if (response.data.status === 'success') {
        setSuccess(true);
        setFormData({
          name: '',
          email: '',
          message: '',
        });
      } else {
        setError(response.data.message || 'Произошла ошибка');
      }
    } catch (err) {
      console.error('Ошибка отправки сообщения:', err);
      const errorMessage = err.response?.data?.message || err.response?.data || err.message || 'Не удалось отправить сообщение';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  return (
    <div className="support-container">
      <div className="support-card">
        <h1 className="page-title">Поддержка</h1>
        <p className="support-description">
          Есть вопросы или нужна помощь? Напишите нам, и мы обязательно вам ответим!
        </p>

        {success && (
          <div className="support-success">
            <h3>✅ Сообщение отправлено!</h3>
            <p>Мы получили ваше сообщение и ответим вам на указанный email в течение 24 часов.</p>
            <button onClick={() => setSuccess(false)} className="btn-secondary">
              Отправить еще одно сообщение
            </button>
          </div>
        )}

        {error && (
          <div className="support-error">
            <h3>❌ Ошибка</h3>
            <p>{error}</p>
            <button onClick={() => setError('')} className="btn-secondary">
              Закрыть
            </button>
          </div>
        )}

        {!success && !error && (
          <form onSubmit={handleSubmit} className="support-form">
            <div className="form-group">
              <label htmlFor="name">Имя *</label>
              <input
                type="text"
                id="name"
                name="name"
                value={formData.name}
                onChange={handleChange}
                required
                minLength={2}
                maxLength={100}
                placeholder="Ваше имя"
                readOnly={!!user}
                disabled={!!user}
                className={user ? 'disabled-input' : ''}
              />
            </div>

            <div className="form-group">
              <label htmlFor="email">Email *</label>
              <input
                type="email"
                id="email"
                name="email"
                value={formData.email}
                onChange={handleChange}
                required
                placeholder="your@email.com"
                readOnly={!!user}
                disabled={!!user}
                className={user ? 'disabled-input' : ''}
              />
            </div>

            <div className="form-group">
              <label htmlFor="message">Сообщение *</label>
              <textarea
                id="message"
                name="message"
                value={formData.message}
                onChange={handleChange}
                required
                minLength={15}
                maxLength={2000}
                rows={6}
                placeholder="Опишите ваш вопрос или проблему (минимум 15 символов)..."
              />
              <span className="char-count">{formData.message.length} / 2000 символов</span>
            </div>

            <div className="form-note">
              <p>Важно: Сообщение должно содержать минимум 15 символов и хотя бы одну букву или цифру.</p>
            </div>

            <button type="submit" className="btn-primary" disabled={loading}>
              {loading ? 'Отправка...' : 'Отправить сообщение'}
            </button>
          </form>
        )}
      </div>
    </div>
  );
};

export default Support;

