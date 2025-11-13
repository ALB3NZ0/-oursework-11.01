import React, { useState, useEffect } from 'react';
import { adminAPI } from '../../api';
import Pagination from '../../components/Pagination';
import * as XLSX from 'xlsx';
import './AdminComponents.css';

const OrdersAdmin = () => {
  const [orders, setOrders] = useState([]);
  const [filteredOrders, setFilteredOrders] = useState([]);
  const [orderDetails, setOrderDetails] = useState({});
  const [loading, setLoading] = useState(true);
  const [selectedOrder, setSelectedOrder] = useState(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(1);
  
  // Фильтры
  const [filterOrderId, setFilterOrderId] = useState('');
  const [filterUserId, setFilterUserId] = useState('');
  const [filterDateFrom, setFilterDateFrom] = useState('');
  const [filterDateTo, setFilterDateTo] = useState('');
  const [filterAmountMin, setFilterAmountMin] = useState('');
  const [filterAmountMax, setFilterAmountMax] = useState('');
  
  // Сортировка
  const [sortBy, setSortBy] = useState('date_desc'); // 'date_asc', 'date_desc', 'amount_asc', 'amount_desc', 'id_asc', 'id_desc', 'user_asc', 'user_desc'

  useEffect(() => {
    loadData();
  }, [currentPage, limit]);

  // Применение фильтров и сортировки
  useEffect(() => {
    let filtered = [...orders];

    // Фильтр по ID заказа
    if (filterOrderId) {
      filtered = filtered.filter(order => order.id.toString().includes(filterOrderId));
    }

    // Фильтр по ID пользователя
    if (filterUserId) {
      filtered = filtered.filter(order => order.user_id.toString().includes(filterUserId));
    }

    // Фильтр по дате
    if (filterDateFrom) {
      const fromDate = new Date(filterDateFrom);
      filtered = filtered.filter(order => new Date(order.order_date) >= fromDate);
    }
    if (filterDateTo) {
      const toDate = new Date(filterDateTo);
      toDate.setHours(23, 59, 59, 999); // Конец дня
      filtered = filtered.filter(order => new Date(order.order_date) <= toDate);
    }

    // Вычисляем суммы для каждого заказа и фильтруем по сумме
    const ordersWithAmounts = filtered.map(order => {
      const details = orderDetails[order.id] || [];
      const totalAmount = details.reduce((sum, item) => sum + (item.price * item.quantity), 0);
      return { ...order, totalAmount };
    });

    if (filterAmountMin) {
      const minAmount = parseFloat(filterAmountMin);
      filtered = ordersWithAmounts.filter(order => order.totalAmount >= minAmount);
    }
    if (filterAmountMax) {
      const maxAmount = parseFloat(filterAmountMax);
      filtered = filtered.filter(order => order.totalAmount <= maxAmount);
    }

    // Сортировка
    filtered.sort((a, b) => {
      const detailsA = orderDetails[a.id] || [];
      const detailsB = orderDetails[b.id] || [];
      const totalAmountA = detailsA.reduce((sum, item) => sum + (item.price * item.quantity), 0);
      const totalAmountB = detailsB.reduce((sum, item) => sum + (item.price * item.quantity), 0);

      switch (sortBy) {
        case 'date_asc':
          return new Date(a.order_date) - new Date(b.order_date);
        case 'date_desc':
          return new Date(b.order_date) - new Date(a.order_date);
        case 'amount_asc':
          return totalAmountA - totalAmountB;
        case 'amount_desc':
          return totalAmountB - totalAmountA;
        case 'id_asc':
          return a.id - b.id;
        case 'id_desc':
          return b.id - a.id;
        case 'user_asc':
          return a.user_id - b.user_id;
        case 'user_desc':
          return b.user_id - a.user_id;
        default:
          return 0;
      }
    });

    setFilteredOrders(filtered);
  }, [orders, orderDetails, filterOrderId, filterUserId, filterDateFrom, filterDateTo, filterAmountMin, filterAmountMax, sortBy]);

  const loadData = async () => {
    setLoading(true);
    try {
      const response = await adminAPI.orders.getAll(currentPage, limit);
      
      // Обработка нового формата ответа с пагинацией
      let loadedOrders = [];
      if (response.data && response.data.data) {
        loadedOrders = response.data.data || [];
        setTotal(response.data.total || 0);
        setTotalPages(response.data.total_pages || 1);
      } else {
        // Fallback для старого формата (массив)
        loadedOrders = response.data || [];
        setTotal(loadedOrders.length || 0);
        setTotalPages(1);
      }
      
      setOrders(loadedOrders);
      setFilteredOrders(loadedOrders); // Инициализация отфильтрованных заказов
      
      // Загружаем детали для каждого заказа
      const details = {};
      for (const order of loadedOrders) {
        try {
          const detailsRes = await adminAPI.orderProducts.getByOrderId(order.id);
          details[order.id] = detailsRes.data || [];
        } catch (error) {
          details[order.id] = [];
        }
      }
      setOrderDetails(details);
    } catch (error) {
      console.error('Ошибка загрузки:', error);
      alert('Не удалось загрузить заказы');
    } finally {
      setLoading(false);
    }
  };

  const handleExport = () => {
    try {
      const exportData = [];

      filteredOrders.forEach(order => {
        const details = orderDetails[order.id] || [];
        const totalAmount = details.reduce((sum, item) => sum + (item.price * item.quantity), 0);
        const totalItems = details.reduce((sum, item) => sum + item.quantity, 0);

        if (details.length === 0) {
          // Заказ без товаров
          exportData.push({
            'ID заказа': order.id,
            'ID пользователя': order.user_id,
            'Дата заказа': new Date(order.order_date).toLocaleString('ru-RU'),
            'Товар': '',
            'Размер': '',
            'Количество': 0,
            'Цена за единицу': 0,
            'Сумма по товару': 0,
            'Всего товаров': 0,
            'Общая сумма': 0,
          });
        } else {
          // Каждый товар в заказе - отдельная строка
          details.forEach((item, index) => {
            exportData.push({
              'ID заказа': order.id,
              'ID пользователя': order.user_id,
              'Дата заказа': new Date(order.order_date).toLocaleString('ru-RU'),
              'Товар': item.product_name || '',
              'Размер': item.size || '',
              'Количество': item.quantity,
              'Цена за единицу': item.price,
              'Сумма по товару': (item.price * item.quantity).toFixed(2),
              'Всего товаров': index === 0 ? totalItems : '',
              'Общая сумма': index === 0 ? totalAmount.toFixed(2) : '',
            });
          });
        }
      });

      const worksheet = XLSX.utils.json_to_sheet(exportData);
      const workbook = XLSX.utils.book_new();
      XLSX.utils.book_append_sheet(workbook, worksheet, 'Заказы');

      const filename = `orders_export_${new Date().toISOString().split('T')[0]}.xlsx`;
      XLSX.writeFile(workbook, filename);
      alert('Данные успешно экспортированы!');
    } catch (error) {
      console.error('Ошибка экспорта:', error);
      alert('Не удалось экспортировать данные');
    }
  };

  if (loading) return <div className="loading-text">Загрузка...</div>;

  return (
    <div className="admin-section">
      <div className="section-header">
        <h2>Управление заказами</h2>
        <div className="header-actions">
          <button className="btn-secondary" onClick={handleExport} title="Экспорт в Excel">
            📥 Экспорт
          </button>
          <button className="btn-secondary" onClick={loadData}>
            Обновить
          </button>
        </div>
      </div>

      {/* Фильтры и сортировка */}
      <div className="admin-filters" style={{ marginBottom: '20px', padding: '15px', background: 'var(--bg-primary)', borderRadius: '16px', border: '1px solid var(--border-color)' }}>
        <h3 style={{ marginTop: 0, marginBottom: '15px', color: 'var(--text-primary)' }}>Фильтры и сортировка</h3>
        
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '15px', marginBottom: '15px' }}>
          <div>
            <label style={{ display: 'block', marginBottom: '5px', fontWeight: 'bold', color: 'var(--text-primary)' }}>ID заказа</label>
            <input
              type="text"
              placeholder="Поиск по ID..."
              value={filterOrderId}
              onChange={(e) => setFilterOrderId(e.target.value)}
              style={{ width: '100%', padding: '8px', borderRadius: '12px', border: '2px solid var(--border-color)', background: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            />
          </div>

          <div>
            <label style={{ display: 'block', marginBottom: '5px', fontWeight: 'bold', color: 'var(--text-primary)' }}>ID пользователя</label>
            <input
              type="text"
              placeholder="Поиск по ID..."
              value={filterUserId}
              onChange={(e) => setFilterUserId(e.target.value)}
              style={{ width: '100%', padding: '8px', borderRadius: '12px', border: '2px solid var(--border-color)', background: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            />
          </div>

          <div>
            <label style={{ display: 'block', marginBottom: '5px', fontWeight: 'bold', color: 'var(--text-primary)' }}>Дата от</label>
            <input
              type="date"
              value={filterDateFrom}
              onChange={(e) => setFilterDateFrom(e.target.value)}
              style={{ width: '100%', padding: '8px', borderRadius: '12px', border: '2px solid var(--border-color)', background: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            />
          </div>

          <div>
            <label style={{ display: 'block', marginBottom: '5px', fontWeight: 'bold', color: 'var(--text-primary)' }}>Дата до</label>
            <input
              type="date"
              value={filterDateTo}
              onChange={(e) => setFilterDateTo(e.target.value)}
              style={{ width: '100%', padding: '8px', borderRadius: '12px', border: '2px solid var(--border-color)', background: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            />
          </div>

          <div>
            <label style={{ display: 'block', marginBottom: '5px', fontWeight: 'bold', color: 'var(--text-primary)' }}>Сумма от (₽)</label>
            <input
              type="number"
              placeholder="Мин. сумма"
              value={filterAmountMin}
              onChange={(e) => setFilterAmountMin(e.target.value)}
              style={{ width: '100%', padding: '8px', borderRadius: '12px', border: '2px solid var(--border-color)', background: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            />
          </div>

          <div>
            <label style={{ display: 'block', marginBottom: '5px', fontWeight: 'bold', color: 'var(--text-primary)' }}>Сумма до (₽)</label>
            <input
              type="number"
              placeholder="Макс. сумма"
              value={filterAmountMax}
              onChange={(e) => setFilterAmountMax(e.target.value)}
              style={{ width: '100%', padding: '8px', borderRadius: '12px', border: '2px solid var(--border-color)', background: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            />
          </div>

          <div>
            <label style={{ display: 'block', marginBottom: '5px', fontWeight: 'bold', color: 'var(--text-primary)' }}>Сортировка</label>
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value)}
              style={{ width: '100%', padding: '8px', borderRadius: '12px', border: '2px solid var(--border-color)', background: 'var(--bg-primary)', color: 'var(--text-primary)' }}
            >
              <option value="date_desc">По дате: сначала новые</option>
              <option value="date_asc">По дате: сначала старые</option>
              <option value="amount_desc">По сумме: сначала большие</option>
              <option value="amount_asc">По сумме: сначала маленькие</option>
              <option value="id_desc">По ID: сначала большие</option>
              <option value="id_asc">По ID: сначала маленькие</option>
              <option value="user_asc">По пользователю: по возрастанию</option>
              <option value="user_desc">По пользователю: по убыванию</option>
            </select>
          </div>
        </div>

        {(filterOrderId || filterUserId || filterDateFrom || filterDateTo || filterAmountMin || filterAmountMax) && (
          <button
            className="btn-secondary"
            onClick={() => {
              setFilterOrderId('');
              setFilterUserId('');
              setFilterDateFrom('');
              setFilterDateTo('');
              setFilterAmountMin('');
              setFilterAmountMax('');
            }}
            style={{ marginTop: '10px' }}
          >
            Сбросить фильтры
          </button>
        )}

        <div style={{ marginTop: '10px', color: 'var(--text-secondary)', fontSize: '14px' }}>
          Найдено заказов: {filteredOrders.length} из {orders.length}
        </div>
      </div>

      <div className="orders-grid">
        {filteredOrders.length === 0 ? (
          <div className="empty-state">
            {orders.length === 0 ? 'Нет заказов' : 'Заказы не найдены по заданным фильтрам'}
          </div>
        ) : (
          filteredOrders.map(order => {
            const details = orderDetails[order.id] || [];
            const totalAmount = details.reduce((sum, item) => sum + (item.price * item.quantity), 0);
            const totalItems = details.reduce((sum, item) => sum + item.quantity, 0);
            
            return (
              <div key={order.id} className="order-card">
                <div className="order-header">
                  <h3>Заказ #{order.id}</h3>
                  <span className="order-date">
                    {new Date(order.order_date).toLocaleString('ru-RU')}
                  </span>
                </div>
                
                <div className="order-info">
                  <p><strong>Пользователь ID:</strong> {order.user_id}</p>
                  <p><strong>Товаров:</strong> {totalItems} шт</p>
                  <p><strong>Сумма:</strong> {totalAmount.toFixed(2)} ₽</p>
                </div>

                {details.length > 0 && (
                  <div className="order-items">
                    <h4>Товары:</h4>
                    {details.map(item => (
                      <div key={item.id} className="order-item">
                        <span>{item.product_name}</span>
                        <span>Размер: {item.size}</span>
                        <span>{item.quantity} шт × {item.price} ₽</span>
                      </div>
                    ))}
                  </div>
                )}

                <button
                  className="btn-view-details"
                  onClick={() => setSelectedOrder(selectedOrder === order.id ? null : order.id)}
                >
                  {selectedOrder === order.id ? 'Скрыть детали' : 'Показать детали'}
                </button>
              </div>
            );
          })
        )}
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

export default OrdersAdmin;



