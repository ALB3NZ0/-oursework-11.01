import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { authAPI, passwordAPI } from '../api';
import { useAuth } from '../AuthContext';
import { useNotification } from '../components/Notification';
import './Auth.css';

const Login = () => {
  const { showSuccess, showError } = useNotification();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();

  // Password reset states
  const [showResetModal, setShowResetModal] = useState(false);
  const [resetEmail, setResetEmail] = useState('');
  const [resetCode, setResetCode] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [resetStep, setResetStep] = useState('email'); // 'email' or 'confirm'
  const [resetLoading, setResetLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    console.log('🚀 Attempting login for:', email);

    try {
      console.log('📤 Sending login request to backend...');
      const response = await authAPI.login({ email, password });
      console.log('✅ Login response received:', response);
      
      const { token, user } = response.data;
      console.log('🔑 Token received, length:', token.length);
      console.log('👤 User data received:', user);

      // Normalize user data to lowercase field names
      if (!user) {
        throw new Error('User data not received from server');
      }
      
      const normalizedUser = {
        id: user.id || user.ID || 0,
        fullname: (user.fullname || user.FullName || user.fullName || '').trim(),
        email: (user.email || user.Email || '').trim(),
        role_id: user.role_id || user.roleId || user.RoleID || 3
      };
      
      console.log('👤 Normalized user data:', normalizedUser);

      // Use full user data from backend response
      login(normalizedUser, token);
      console.log('✅ Login successful, navigating based on role');
      
      // Перенаправляем в зависимости от роли
      if (normalizedUser.role_id === 1) {
        navigate('/admin');
      } else if (normalizedUser.role_id === 2) {
        navigate('/manager');
      } else {
        navigate('/');
      }
    } catch (err) {
      console.error('❌ Login error:', err);
      console.error('Error response:', err.response);
      console.error('Error data:', err.response?.data);
      setError(err.response?.data || err.message || 'Ошибка входа. Проверьте данные.');
    } finally {
      setLoading(false);
    }
  };

  const handlePasswordReset = async (e) => {
    e.preventDefault();
    setError('');
    setResetLoading(true);

    try {
      await passwordAPI.reset({ email: resetEmail });
      showSuccess('Код восстановления отправлен на ваш email');
      setResetStep('confirm');
      setError('');
    } catch (err) {
      console.error('Password reset error:', err);
      const errorMsg = err.response?.data || err.message || 'Не удалось отправить запрос';
      setError(errorMsg);
      showError(errorMsg);
    } finally {
      setResetLoading(false);
    }
  };

  const handleConfirmPasswordReset = async (e) => {
    e.preventDefault();
    
    if (newPassword !== confirmPassword) {
      setError('Пароли не совпадают');
      return;
    }

    if (newPassword.length < 8) {
      setError('Пароль должен содержать минимум 8 символов');
      return;
    }

    setResetLoading(true);

    try {
      await passwordAPI.confirmReset({
        email: resetEmail,
        code: resetCode,
        password: newPassword,
      });
      
      showSuccess('Пароль успешно восстановлен! Теперь вы можете войти.');
      setShowResetModal(false);
      setResetEmail('');
      setResetCode('');
      setNewPassword('');
      setConfirmPassword('');
      setResetStep('email');
      setError('');
    } catch (err) {
      console.error('Confirm password reset error:', err);
      const errorMsg = err.response?.data || err.message || 'Неверный код или не удалось восстановить пароль';
      setError(errorMsg);
      showError(errorMsg);
    } finally {
      setResetLoading(false);
    }
  };

  return (
    <div className="auth-container">
      <div className="auth-card">
        <h1 className="auth-title">Shoes Store</h1>
        <h2 className="auth-subtitle">Вход в систему</h2>

        {error && <div className="error-message">{error}</div>}

        <form onSubmit={handleSubmit} className="auth-form">
          <div className="form-group">
            <label htmlFor="email">Email</label>
            <input
              type="email"
              id="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              placeholder="your@email.com"
            />
          </div>

          <div className="form-group">
            <label htmlFor="password">Пароль</label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              placeholder="••••••••"
            />
          </div>

          <button type="submit" className="auth-button" disabled={loading}>
            {loading ? 'Вход...' : 'Войти'}
          </button>
        </form>

        <p className="auth-link">
          Нет аккаунта? <Link to="/register">Зарегистрироваться</Link>
        </p>

        <p className="auth-link">
          <a 
            href="#" 
            onClick={(e) => {
              e.preventDefault();
              setShowResetModal(true);
              setResetStep('email');
            }}
            style={{ color: '#666', textDecoration: 'underline', cursor: 'pointer' }}
          >
            Забыли пароль?
          </a>
        </p>
      </div>

      {/* Password Reset Modal */}
      {showResetModal && (
        <div className="modal-overlay" onClick={() => {
          setShowResetModal(false);
          setResetEmail('');
          setResetCode('');
          setNewPassword('');
          setConfirmPassword('');
          setResetStep('email');
        }}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h3>{resetStep === 'email' ? 'Восстановление пароля' : 'Подтверждение смены пароля'}</h3>
            
            {error && <div className="error-message" style={{ marginBottom: '1rem' }}>{error}</div>}

            {resetStep === 'email' ? (
              <form onSubmit={handlePasswordReset}>
                <div className="form-group">
                  <label>Email</label>
                  <input
                    type="email"
                    value={resetEmail}
                    onChange={(e) => setResetEmail(e.target.value)}
                    required
                    placeholder="your@email.com"
                  />
                </div>

                <button type="submit" className="btn-primary" disabled={resetLoading}>
                  {resetLoading ? 'Отправка...' : 'Отправить код'}
                </button>
              </form>
            ) : (
              <form onSubmit={handleConfirmPasswordReset}>
                <div className="form-group">
                  <label>Код подтверждения</label>
                  <input
                    type="text"
                    value={resetCode}
                    onChange={(e) => setResetCode(e.target.value)}
                    required
                    placeholder="Введите 6-значный код"
                    maxLength={6}
                  />
                </div>

                <div className="form-group">
                  <label>Новый пароль</label>
                  <input
                    type="password"
                    value={newPassword}
                    onChange={(e) => setNewPassword(e.target.value)}
                    required
                    placeholder="Минимум 8 символов"
                    minLength={8}
                  />
                </div>

                <div className="form-group">
                  <label>Подтвердите пароль</label>
                  <input
                    type="password"
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    required
                    placeholder="••••••••"
                  />
                </div>

                <button type="submit" className="btn-primary" disabled={resetLoading}>
                  {resetLoading ? 'Восстановление...' : 'Восстановить пароль'}
                </button>
              </form>
            )}

            <button
              className="btn-secondary"
              onClick={() => {
                setShowResetModal(false);
                setResetEmail('');
                setResetCode('');
                setNewPassword('');
                setConfirmPassword('');
                setResetStep('email');
              }}
              disabled={resetLoading}
              style={{ marginTop: '1rem', width: '100%' }}
            >
              Отмена
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default Login;

