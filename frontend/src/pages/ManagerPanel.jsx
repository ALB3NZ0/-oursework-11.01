import React, { useState, useEffect } from 'react';
import { useAuth } from '../AuthContext';
import { useNotification } from '../components/Notification';
import { reportsAPI, productsAPI, categoriesAPI, brandsAPI, ordersAPI } from '../api';
import * as XLSX from 'xlsx';
import './ManagerPanel.css';

const ManagerPanel = () => {
  const { user } = useAuth();
  const { showSuccess, showError } = useNotification();
  const [generating, setGenerating] = useState('');
  const [error, setError] = useState('');
  const [viewerOpen, setViewerOpen] = useState(false);
  const [viewerContent, setViewerContent] = useState(null);
  const [viewerFilename, setViewerFilename] = useState('');
  const [viewerType, setViewerType] = useState(''); // 'pdf', 'text', or 'excel'
  const [excelData, setExcelData] = useState(null); // Данные Excel для просмотра
  
  // Данные для графиков
  const [statsLoading, setStatsLoading] = useState(true);
  const [products, setProducts] = useState([]);
  const [categories, setCategories] = useState([]);
  const [brands, setBrands] = useState([]);
  const [orders, setOrders] = useState([]);

  // Проверка что пользователь менеджер
  const isManager = user && user.role_id === 2;

  // Загрузка данных для графиков
  useEffect(() => {
    if (isManager) {
      loadStatistics();
    }
  }, [isManager]);

  const loadStatistics = async () => {
    setStatsLoading(true);
    try {
      const [productsRes, categoriesRes, brandsRes, ordersRes] = await Promise.all([
        productsAPI.getAll(1, 1000).catch(() => ({ data: { data: [] } })), // Загружаем много товаров для статистики
        categoriesAPI.getAll().catch(() => ({ data: [] })),
        brandsAPI.getAll().catch(() => ({ data: [] })),
        ordersAPI.getAll().catch(() => ({ data: { data: [] } })), // может не работать для менеджера
      ]);

      // Обработка пагинированного ответа для products
      let productsData = [];
      if (productsRes.data && productsRes.data.data) {
        productsData = Array.isArray(productsRes.data.data) ? productsRes.data.data : [];
      } else if (Array.isArray(productsRes.data)) {
        productsData = productsRes.data;
      }
      setProducts(productsData);

      // Обработка пагинированного ответа для orders
      let ordersData = [];
      if (ordersRes.data && ordersRes.data.data) {
        ordersData = Array.isArray(ordersRes.data.data) ? ordersRes.data.data : [];
      } else if (Array.isArray(ordersRes.data)) {
        ordersData = ordersRes.data;
      }
      setOrders(ordersData);

      // Categories и brands обычно возвращают массивы напрямую
      setCategories(Array.isArray(categoriesRes.data) ? categoriesRes.data : []);
      setBrands(Array.isArray(brandsRes.data) ? brandsRes.data : []);
    } catch (error) {
      console.error('Ошибка загрузки статистики:', error);
      setProducts([]);
      setCategories([]);
      setBrands([]);
      setOrders([]);
    } finally {
      setStatsLoading(false);
    }
  };

  const downloadFile = (blob, filename) => {
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  };

  const openViewer = (blob, filename, type) => {
    if (type === 'excel') {
      // Для Excel файлов парсим данные для просмотра
      blob.arrayBuffer().then(buffer => {
        const workbook = XLSX.read(buffer, { type: 'array' });
        const sheetName = workbook.SheetNames[0];
        const worksheet = workbook.Sheets[sheetName];
        const data = XLSX.utils.sheet_to_json(worksheet, { header: 1, defval: '' });
        setExcelData({ sheetName, data });
        setViewerContent(null);
        setViewerFilename(filename);
        setViewerType('excel');
        setViewerOpen(true);
      }).catch(err => {
        console.error('Ошибка парсинга Excel:', err);
        // Если не получилось распарсить, просто скачиваем
        downloadFile(blob, filename);
      });
    } else {
      const url = window.URL.createObjectURL(blob);
      setViewerContent(url);
      setViewerFilename(filename);
      setViewerType(type);
      setViewerOpen(true);
      // Также скачиваем файл автоматически
      downloadFile(blob, filename);
    }
  };

  const closeViewer = () => {
    if (viewerContent) {
      window.URL.revokeObjectURL(viewerContent);
    }
    setViewerOpen(false);
    setViewerContent(null);
    setExcelData(null);
    setViewerFilename('');
    setViewerType('');
  };

  const generatePDF = async (type) => {
    if (!isManager) return;
    
    setGenerating(type);
    setError('');
    try {
      let response;
      let filename;

      switch (type) {
        case 'sales':
          response = await reportsAPI.generateSalesPDF();
          filename = `sales_report_${new Date().toISOString().split('T')[0]}.pdf`;
          break;
        case 'inventory':
          response = await reportsAPI.generateInventoryPDF();
          filename = `inventory_report_${new Date().toISOString().split('T')[0]}.pdf`;
          break;
        case 'customers':
          response = await reportsAPI.generateCustomersPDF();
          filename = `customers_report_${new Date().toISOString().split('T')[0]}.pdf`;
          break;
        case 'categories':
          response = await reportsAPI.generateCategoriesPDF();
          filename = `categories_report_${new Date().toISOString().split('T')[0]}.pdf`;
          break;
        default:
          return;
      }

      // Открываем просмотрщик
      openViewer(response.data, filename, 'pdf');
    } catch (error) {
      console.error('Ошибка генерации PDF:', error);
      const errorMsg = error.response?.data?.error || 'Не удалось сгенерировать отчет';
      setError(errorMsg);
      showError(errorMsg);
    } finally {
      setGenerating('');
    }
  };

  const generateExcel = async (type) => {
    if (!isManager) return;
    
    setGenerating(type);
    setError('');
    try {
      let response;
      let filename;

      switch (type) {
        case 'sales_excel':
          response = await reportsAPI.generateSalesExcel();
          filename = `sales_report_${new Date().toISOString().split('T')[0]}.xlsx`;
          break;
        case 'inventory_excel':
          response = await reportsAPI.generateInventoryExcel();
          filename = `inventory_report_${new Date().toISOString().split('T')[0]}.xlsx`;
          break;
        case 'customers_excel':
          response = await reportsAPI.generateCustomersExcel();
          filename = `customers_report_${new Date().toISOString().split('T')[0]}.xlsx`;
          break;
        case 'categories_excel':
          response = await reportsAPI.generateCategoriesExcel();
          filename = `categories_report_${new Date().toISOString().split('T')[0]}.xlsx`;
          break;
        default:
          return;
      }

      // Открываем просмотрщик для Excel
      openViewer(response.data, filename, 'excel');
    } catch (error) {
      console.error('Ошибка генерации Excel отчета:', error);
      const errorMsg = error.response?.data?.error || 'Не удалось сгенерировать отчет';
      setError(errorMsg);
      showError(errorMsg);
    } finally {
      setGenerating('');
    }
  };

  const generateText = async (type) => {
    if (!isManager) return;
    
    setGenerating(type);
    setError('');
    try {
      let response;
      let filename;

      switch (type) {
        case 'customers_text':
          response = await reportsAPI.generateCustomersText();
          filename = `customers_report_${new Date().toISOString().split('T')[0]}.txt`;
          break;
        case 'inventory_text':
          response = await reportsAPI.generateInventoryText();
          filename = `inventory_report_${new Date().toISOString().split('T')[0]}.txt`;
          break;
        default:
          return;
      }

      // Открываем просмотрщик для текста
      openViewer(response.data, filename, 'text');
    } catch (error) {
      console.error('Ошибка генерации текстового отчета:', error);
      const errorMsg = error.response?.data?.error || 'Не удалось сгенерировать отчет';
      setError(errorMsg);
      showError(errorMsg);
    } finally {
      setGenerating('');
    }
  };

  if (!isManager) {
    return (
      <div className="manager-panel-container">
        <div className="access-denied">
          <h2>Доступ запрещен</h2>
          <p>У вас нет прав для доступа к панели менеджера.</p>
          <p>Требуется роль менеджера.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="manager-panel-container">
      <h1 className="page-title">Панель менеджера</h1>

      {error && (
        <div className="error-message">
          <p>{error}</p>
          <button onClick={() => setError('')}>Закрыть</button>
        </div>
      )}

      <div className="manager-sections">
        {/* Графики и статистика */}
        <section className="manager-section">
          <h2>Статистика и аналитика</h2>
          <p className="section-description">
            Визуализация данных каталога и продаж
          </p>

          {statsLoading ? (
            <div className="loading-text">Загрузка данных...</div>
          ) : (
            <>
              {/* Статистические карточки */}
              <div className="stats-cards">
                <div className="stat-card">
                  <div className="stat-info">
                    <h3>{products.length}</h3>
                    <p>Товаров в каталоге</p>
                  </div>
                </div>
                <div className="stat-card">
                  <div className="stat-info">
                    <h3>{categories.length}</h3>
                    <p>Категорий</p>
                  </div>
                </div>
                <div className="stat-card">
                  <div className="stat-info">
                    <h3>{brands.length}</h3>
                    <p>Брендов</p>
                  </div>
                </div>
                <div className="stat-card">
                  <div className="stat-info">
                    <h3>{orders.length}</h3>
                    <p>Всего заказов</p>
                  </div>
                </div>
              </div>

              {/* Графики */}
              <div className="charts-grid">
                <div className="chart-card">
                  <h3>Распределение по категориям</h3>
                  <CategoryChart products={products} categories={categories} />
                </div>
                
                <div className="chart-card">
                  <h3>Распределение по брендам (круговая)</h3>
                  <BrandPieChart products={products} brands={brands} />
                </div>
                
                <div className="chart-card">
                  <h3>Распределение по категориям (круговая)</h3>
                  <CategoryPieChart products={products} categories={categories} />
                </div>
                
                <div className="chart-card">
                  <h3>Диапазон цен товаров</h3>
                  <PriceChart products={products} />
                </div>

                <div className="chart-card">
                  <h3>Заказы по дням</h3>
                  <OrdersChart orders={orders} />
                </div>
              </div>
            </>
          )}
        </section>

        {/* Генерация PDF отчетов */}
        <section className="manager-section">
          <h2>PDF Отчеты</h2>
          <p className="section-description">
            Генерация отчетов в формате PDF
          </p>
          
          <div className="reports-grid">
            <div className="report-card">
              <h3>Отчет по продажам</h3>
              <p>Статистика продаж и выручка</p>
              <button
                className="btn-generate"
                onClick={() => generatePDF('sales')}
                disabled={generating === 'sales'}
              >
                {generating === 'sales' ? 'Генерация...' : 'Сгенерировать PDF'}
              </button>
            </div>

            <div className="report-card">
              <h3>Отчет по инвентарю</h3>
              <p>Остатки товаров на складе</p>
              <button
                className="btn-generate"
                onClick={() => generatePDF('inventory')}
                disabled={generating === 'inventory'}
              >
                {generating === 'inventory' ? 'Генерация...' : 'Сгенерировать PDF'}
              </button>
            </div>

            <div className="report-card">
              <h3>Отчет по клиентам</h3>
              <p>Информация о покупателях</p>
              <button
                className="btn-generate"
                onClick={() => generatePDF('customers')}
                disabled={generating === 'customers'}
              >
                {generating === 'customers' ? 'Генерация...' : 'Сгенерировать PDF'}
              </button>
            </div>

            <div className="report-card">
              <h3>Отчет по категориям</h3>
              <p>Статистика по категориям товаров</p>
              <button
                className="btn-generate"
                onClick={() => generatePDF('categories')}
                disabled={generating === 'categories'}
              >
                {generating === 'categories' ? 'Генерация...' : 'Сгенерировать PDF'}
              </button>
            </div>
          </div>
        </section>

        {/* Генерация Excel отчетов */}
        <section className="manager-section">
          <h2>Excel Отчеты</h2>
          <p className="section-description">
            Генерация отчетов в формате Excel (XLSX)
          </p>
          
          <div className="reports-grid">
            <div className="report-card">
              <h3>Отчет по продажам</h3>
              <p>Статистика продаж и выручка в Excel</p>
              <button
                className="btn-generate"
                onClick={() => generateExcel('sales_excel')}
                disabled={generating === 'sales_excel'}
              >
                {generating === 'sales_excel' ? 'Генерация...' : 'Сгенерировать Excel'}
              </button>
            </div>

            <div className="report-card">
              <h3>Отчет по инвентарю</h3>
              <p>Остатки товаров на складе в Excel</p>
              <button
                className="btn-generate"
                onClick={() => generateExcel('inventory_excel')}
                disabled={generating === 'inventory_excel'}
              >
                {generating === 'inventory_excel' ? 'Генерация...' : 'Сгенерировать Excel'}
              </button>
            </div>

            <div className="report-card">
              <h3>Отчет по клиентам</h3>
              <p>Информация о покупателях в Excel</p>
              <button
                className="btn-generate"
                onClick={() => generateExcel('customers_excel')}
                disabled={generating === 'customers_excel'}
              >
                {generating === 'customers_excel' ? 'Генерация...' : 'Сгенерировать Excel'}
              </button>
            </div>

            <div className="report-card">
              <h3>Отчет по категориям</h3>
              <p>Статистика по категориям товаров в Excel</p>
              <button
                className="btn-generate"
                onClick={() => generateExcel('categories_excel')}
                disabled={generating === 'categories_excel'}
              >
                {generating === 'categories_excel' ? 'Генерация...' : 'Сгенерировать Excel'}
              </button>
            </div>
          </div>
        </section>

        {/* Генерация текстовых отчетов */}
        <section className="manager-section">
          <h2>Текстовые отчеты</h2>
          <p className="section-description">
            Генерация отчетов в текстовом формате (UTF-8)
          </p>
          
          <div className="reports-grid">
            <div className="report-card">
              <h3>Отчет по клиентам (TXT)</h3>
              <p>Информация о покупателях в текстовом формате</p>
              <button
                className="btn-generate"
                onClick={() => generateText('customers_text')}
                disabled={generating === 'customers_text'}
              >
                {generating === 'customers_text' ? 'Генерация...' : 'Сгенерировать TXT'}
              </button>
            </div>

            <div className="report-card">
              <h3>Отчет по инвентарю (TXT)</h3>
              <p>Остатки товаров в текстовом формате</p>
              <button
                className="btn-generate"
                onClick={() => generateText('inventory_text')}
                disabled={generating === 'inventory_text'}
              >
                {generating === 'inventory_text' ? 'Генерация...' : 'Сгенерировать TXT'}
              </button>
            </div>
          </div>
        </section>
      </div>

      {/* Модальное окно для просмотра отчетов */}
      {viewerOpen && (
        <div className="viewer-overlay" onClick={closeViewer}>
          <div className="viewer-modal" onClick={(e) => e.stopPropagation()}>
            <div className="viewer-header">
              <h3>{viewerFilename}</h3>
              <div className="viewer-actions">
                <button className="btn-close-viewer" onClick={closeViewer}>
                  ✕ Закрыть
                </button>
              </div>
            </div>
            <div className="viewer-content">
              {viewerType === 'pdf' && viewerContent && (
                <iframe
                  src={viewerContent}
                  className="viewer-iframe"
                  title="PDF Viewer"
                />
              )}
              {viewerType === 'text' && viewerContent && (
                <TextContentViewer url={viewerContent} />
              )}
              {viewerType === 'excel' && excelData && (
                <ExcelContentViewer data={excelData} filename={viewerFilename} onDownload={() => {
                  // Функция для повторной загрузки и скачивания файла
                  const type = viewerFilename.includes('sales') ? 'sales_excel' :
                               viewerFilename.includes('inventory') ? 'inventory_excel' :
                               viewerFilename.includes('customers') ? 'customers_excel' :
                               viewerFilename.includes('categories') ? 'categories_excel' : '';
                  if (type) {
                    generateExcel(type);
                  }
                }} />
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

// Компонент для отображения текстового контента
const TextContentViewer = ({ url }) => {
  const [text, setText] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch(url)
      .then(res => res.text())
      .then(content => {
        setText(content);
        setLoading(false);
      })
      .catch(err => {
        console.error('Ошибка загрузки текста:', err);
        setText('Не удалось загрузить содержимое файла');
        setLoading(false);
      });
  }, [url]);

  if (loading) {
    return <div className="viewer-loading">Загрузка...</div>;
  }

  return (
    <pre className="viewer-text-content">{text}</pre>
  );
};

// Компонент для отображения Excel контента
const ExcelContentViewer = ({ data, filename, onDownload }) => {
  if (!data || !data.data || data.data.length === 0) {
    return <div className="viewer-loading">Нет данных для отображения</div>;
  }

  const [header, ...rows] = data.data;
  const maxCols = Math.max(...data.data.map(row => row.length), 0);

  return (
    <div className="excel-viewer">
      <div className="excel-viewer-header" style={{ marginBottom: '10px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h4 style={{ margin: 0, color: 'var(--text-primary)' }}>{data.sheetName}</h4>
        <button 
          className="btn-secondary" 
          onClick={onDownload}
          style={{ padding: '8px 16px', fontSize: '14px' }}
        >
          Скачать файл
        </button>
      </div>
      <div className="excel-table-container" style={{ overflow: 'auto', maxHeight: '70vh', border: '1px solid var(--border-color)', borderRadius: '4px' }}>
        <table className="excel-table" style={{ width: '100%', borderCollapse: 'collapse', backgroundColor: 'var(--bg-primary)' }}>
          {header && header.length > 0 && (
            <thead>
              <tr>
                {header.map((cell, index) => (
                  <th
                    key={index}
                    style={{
                      padding: '10px',
                      textAlign: 'left',
                      backgroundColor: '#4472C4',
                      color: '#fff',
                      fontWeight: 'bold',
                      border: '1px solid var(--border-color)',
                      position: 'sticky',
                      top: 0,
                      zIndex: 10
                    }}
                  >
                    {cell || ''}
                  </th>
                ))}
                {/* Заполняем пустые ячейки если нужно */}
                {Array.from({ length: Math.max(0, maxCols - header.length) }).map((_, index) => (
                  <th
                    key={`empty-${index}`}
                    style={{
                      padding: '10px',
                      backgroundColor: '#4472C4',
                      border: '1px solid var(--border-color)'
                    }}
                  ></th>
                ))}
              </tr>
            </thead>
          )}
          <tbody>
            {rows.map((row, rowIndex) => (
              <tr key={rowIndex} style={{ backgroundColor: rowIndex % 2 === 0 ? 'var(--bg-primary)' : 'var(--bg-secondary)' }}>
                {row.map((cell, cellIndex) => (
                  <td
                    key={cellIndex}
                    style={{
                      padding: '8px 10px',
                      border: '1px solid var(--border-color)',
                      whiteSpace: 'nowrap',
                      color: 'var(--text-primary)'
                    }}
                  >
                    {cell !== null && cell !== undefined ? String(cell) : ''}
                  </td>
                ))}
                {/* Заполняем пустые ячейки если нужно */}
                {Array.from({ length: Math.max(0, maxCols - row.length) }).map((_, index) => (
                  <td
                    key={`empty-${index}`}
                    style={{
                      padding: '8px 10px',
                      border: '1px solid var(--border-color)',
                      color: 'var(--text-primary)'
                    }}
                  ></td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

// Компоненты графиков

// Вертикальная столбчатая диаграмма по категориям
const CategoryChartVertical = ({ products, categories }) => {
  // Убеждаемся, что products - это массив
  if (!products || !Array.isArray(products)) {
    return <div className="chart-empty">Нет данных</div>;
  }

  const categoryCounts = {};
  
  products.forEach(product => {
    const categoryId = product.category_id;
    categoryCounts[categoryId] = (categoryCounts[categoryId] || 0) + 1;
  });

  const categoryData = categories.map(cat => ({
    name: cat.category_name || `Категория ${cat.id}`,
    count: categoryCounts[cat.id] || 0,
    color: getColorForIndex(cat.id)
  })).filter(item => item.count > 0).slice(0, 6);

  const maxCount = Math.max(...categoryData.map(d => d.count), 1);
  const chartHeight = 180;
  const barWidth = Math.max(35, (300 - categoryData.length * 50) / categoryData.length);
  const maxBarHeight = chartHeight - 50;

  return (
    <div className="vertical-bar-chart">
      {categoryData.length === 0 ? (
        <div className="chart-empty">Нет данных</div>
      ) : (
        <svg viewBox={`0 0 ${categoryData.length * 60} ${chartHeight}`} className="vertical-bar-svg">
          {categoryData.map((item, index) => {
            const barHeight = (item.count / maxCount) * maxBarHeight;
            const x = index * 60 + 10;
            const y = chartHeight - 40 - barHeight;
            
            return (
              <g key={index}>
                {/* Столбец */}
                <rect
                  x={x}
                  y={y}
                  width={barWidth}
                  height={barHeight}
                  fill={item.color}
                  rx="4"
                  style={{ transition: 'all 0.3s' }}
                />
                {/* Значение на столбце */}
                <text
                  x={x + barWidth / 2}
                  y={y - 5}
                  textAnchor="middle"
                  fontSize="12"
                  fontWeight="600"
                  fill="#212529"
                >
                  {item.count}
                </text>
                {/* Название категории */}
                <text
                  x={x + barWidth / 2}
                  y={chartHeight - 10}
                  textAnchor="middle"
                  fontSize="11"
                  fill="#495057"
                >
                  {item.name.length > 12 ? item.name.substring(0, 10) + '...' : item.name}
                </text>
              </g>
            );
          })}
        </svg>
      )}
    </div>
  );
};

// Круговая диаграмма по брендам
const BrandPieChart = ({ products, brands }) => {
  // Убеждаемся, что products - это массив
  if (!products || !Array.isArray(products)) {
    return <div className="chart-empty">Нет данных</div>;
  }

  const brandCounts = {};
  
  products.forEach(product => {
    const brandId = product.brand_id;
    brandCounts[brandId] = (brandCounts[brandId] || 0) + 1;
  });

  const brandData = brands.map(brand => ({
    name: brand.brand_name || `Бренд ${brand.id}`,
    count: brandCounts[brand.id] || 0,
    color: getColorForIndex(brand.id + 10)
  })).filter(item => item.count > 0).slice(0, 8);

  const total = brandData.reduce((sum, item) => sum + item.count, 0);
  
  if (total === 0 || brandData.length === 0) {
    return <div className="chart-empty">Нет данных</div>;
  }

  let currentAngle = -90;
  const centerX = 150;
  const centerY = 150;
  const radius = 100;

  return (
    <div className="pie-chart">
      <svg viewBox="0 0 300 300" className="pie-chart-svg">
        {brandData.map((item, index) => {
          const percentage = (item.count / total) * 100;
          const angle = (item.count / total) * 360;
          const startAngle = currentAngle;
          const endAngle = currentAngle + angle;
          
          const startRad = (startAngle * Math.PI) / 180;
          const endRad = (endAngle * Math.PI) / 180;
          
          const x1 = centerX + radius * Math.cos(startRad);
          const y1 = centerY + radius * Math.sin(startRad);
          const x2 = centerX + radius * Math.cos(endRad);
          const y2 = centerY + radius * Math.sin(endRad);
          
          const largeArcFlag = angle > 180 ? 1 : 0;
          
          const pathData = [
            `M ${centerX} ${centerY}`,
            `L ${x1} ${y1}`,
            `A ${radius} ${radius} 0 ${largeArcFlag} 1 ${x2} ${y2}`,
            'Z'
          ].join(' ');
          
          const labelAngle = (startAngle + endAngle) / 2;
          const labelRadius = radius * 0.7;
          const labelX = centerX + labelRadius * Math.cos((labelAngle * Math.PI) / 180);
          const labelY = centerY + labelRadius * Math.sin((labelAngle * Math.PI) / 180);
          
          currentAngle = endAngle;
          
          return (
            <g key={index}>
              <path
                d={pathData}
                fill={item.color}
                stroke="white"
                strokeWidth="2"
              />
              {percentage > 5 && (
                <text
                  x={labelX}
                  y={labelY}
                  textAnchor="middle"
                  fontSize="11"
                  fontWeight="600"
                  fill="white"
                  style={{ textShadow: '0 1px 2px rgba(0,0,0,0.3)' }}
                >
                  {percentage.toFixed(1)}%
                </text>
              )}
            </g>
          );
        })}
      </svg>
      <div className="pie-legend">
        {brandData.map((item, index) => (
          <div key={index} className="pie-legend-item">
            <span className="pie-legend-color" style={{ backgroundColor: item.color }}></span>
            <span className="pie-legend-name">{item.name}</span>
            <span className="pie-legend-value">({item.count})</span>
          </div>
        ))}
      </div>
    </div>
  );
};

// Круговая диаграмма по категориям
const CategoryPieChart = ({ products, categories }) => {
  // Убеждаемся, что products - это массив
  if (!products || !Array.isArray(products)) {
    return <div className="chart-empty">Нет данных</div>;
  }

  const categoryCounts = {};
  
  products.forEach(product => {
    const categoryId = product.category_id;
    categoryCounts[categoryId] = (categoryCounts[categoryId] || 0) + 1;
  });

  const categoryData = categories.map(cat => ({
    name: cat.category_name || `Категория ${cat.id}`,
    count: categoryCounts[cat.id] || 0,
    color: getColorForIndex(cat.id)
  })).filter(item => item.count > 0).slice(0, 8);

  const total = categoryData.reduce((sum, item) => sum + item.count, 0);
  
  if (total === 0 || categoryData.length === 0) {
    return <div className="chart-empty">Нет данных</div>;
  }

  let currentAngle = -90;
  const centerX = 150;
  const centerY = 150;
  const radius = 100;

  return (
    <div className="pie-chart">
      <svg viewBox="0 0 300 300" className="pie-chart-svg">
        {categoryData.map((item, index) => {
          const percentage = (item.count / total) * 100;
          const angle = (item.count / total) * 360;
          const startAngle = currentAngle;
          const endAngle = currentAngle + angle;
          
          const startRad = (startAngle * Math.PI) / 180;
          const endRad = (endAngle * Math.PI) / 180;
          
          const x1 = centerX + radius * Math.cos(startRad);
          const y1 = centerY + radius * Math.sin(startRad);
          const x2 = centerX + radius * Math.cos(endRad);
          const y2 = centerY + radius * Math.sin(endRad);
          
          const largeArcFlag = angle > 180 ? 1 : 0;
          
          const pathData = [
            `M ${centerX} ${centerY}`,
            `L ${x1} ${y1}`,
            `A ${radius} ${radius} 0 ${largeArcFlag} 1 ${x2} ${y2}`,
            'Z'
          ].join(' ');
          
          const labelAngle = (startAngle + endAngle) / 2;
          const labelRadius = radius * 0.7;
          const labelX = centerX + labelRadius * Math.cos((labelAngle * Math.PI) / 180);
          const labelY = centerY + labelRadius * Math.sin((labelAngle * Math.PI) / 180);
          
          currentAngle = endAngle;
          
          return (
            <g key={index}>
              <path
                d={pathData}
                fill={item.color}
                stroke="white"
                strokeWidth="2"
              />
              {percentage > 5 && (
                <text
                  x={labelX}
                  y={labelY}
                  textAnchor="middle"
                  fontSize="11"
                  fontWeight="600"
                  fill="white"
                  style={{ textShadow: '0 1px 2px rgba(0,0,0,0.3)' }}
                >
                  {percentage.toFixed(1)}%
                </text>
              )}
            </g>
          );
        })}
      </svg>
      <div className="pie-legend">
        {categoryData.map((item, index) => (
          <div key={index} className="pie-legend-item">
            <span className="pie-legend-color" style={{ backgroundColor: item.color }}></span>
            <span className="pie-legend-name">{item.name}</span>
            <span className="pie-legend-value">({item.count})</span>
          </div>
        ))}
      </div>
    </div>
  );
};

// Вертикальная столбчатая диаграмма цен
const PriceChartVertical = ({ products }) => {
  if (!products || !Array.isArray(products) || products.length === 0) {
    return <div className="chart-empty">Нет данных</div>;
  }

  const prices = products.map(p => parseFloat(p.price) || 0).filter(p => p > 0);
  if (prices.length === 0) {
    return <div className="chart-empty">Нет данных о ценах</div>;
  }

  // Создаем диапазоны цен
  const ranges = [
    { label: '0-2000₽', min: 0, max: 2000, count: 0 },
    { label: '2000-5000₽', min: 2000, max: 5000, count: 0 },
    { label: '5000-10000₽', min: 5000, max: 10000, count: 0 },
    { label: '10000+₽', min: 10000, max: Infinity, count: 0 }
  ];

  prices.forEach(price => {
    for (const range of ranges) {
      if (price >= range.min && price < range.max) {
        range.count++;
        break;
      }
    }
  });

  const maxCount = Math.max(...ranges.map(r => r.count), 1);
  const chartHeight = 180;
  const barWidth = 45;
  const maxBarHeight = chartHeight - 50;
  const spacing = 25;

  return (
    <div className="vertical-bar-chart">
      <svg viewBox={`0 0 ${ranges.length * (barWidth + spacing) + spacing} ${chartHeight}`} className="vertical-bar-svg">
        {ranges.map((range, index) => {
          const barHeight = (range.count / maxCount) * maxBarHeight;
          const x = index * (barWidth + spacing) + spacing;
          const y = chartHeight - 40 - barHeight;
          const color = getColorForIndex(index + 20);
          
          return (
            <g key={index}>
              {/* Столбец */}
              <rect
                x={x}
                y={y}
                width={barWidth}
                height={barHeight}
                fill={color}
                rx="4"
                style={{ transition: 'all 0.3s' }}
              />
              {/* Значение на столбце */}
              <text
                x={x + barWidth / 2}
                y={y - 5}
                textAnchor="middle"
                fontSize="12"
                fontWeight="600"
                fill="#212529"
              >
                {range.count}
              </text>
              {/* Название диапазона */}
              <text
                x={x + barWidth / 2}
                y={chartHeight - 10}
                textAnchor="middle"
                fontSize="11"
                fill="#495057"
              >
                {range.label}
              </text>
            </g>
          );
        })}
      </svg>
      <div className="price-stats">
        <div>Мин: {Math.min(...prices).toFixed(0)}₽</div>
        <div>Средняя: {(prices.reduce((a, b) => a + b, 0) / prices.length).toFixed(0)}₽</div>
        <div>Макс: {Math.max(...prices).toFixed(0)}₽</div>
      </div>
    </div>
  );
};

// График распределения по категориям (старая горизонтальная версия)
const CategoryChart = ({ products, categories }) => {
  // Убеждаемся, что products - это массив
  if (!products || !Array.isArray(products)) {
    return <div className="chart-empty">Нет данных</div>;
  }

  const categoryCounts = {};
  
  products.forEach(product => {
    const categoryId = product.category_id;
    categoryCounts[categoryId] = (categoryCounts[categoryId] || 0) + 1;
  });

  const categoryData = categories.map(cat => ({
    name: cat.category_name || `Категория ${cat.id}`,
    count: categoryCounts[cat.id] || 0,
    color: getColorForIndex(cat.id)
  })).filter(item => item.count > 0).slice(0, 6);

  const maxCount = Math.max(...categoryData.map(d => d.count), 1);

  return (
    <div className="bar-chart">
      {categoryData.length === 0 ? (
        <div className="chart-empty">Нет данных</div>
      ) : (
        categoryData.map((item, index) => (
          <div key={index} className="bar-item">
            <div className="bar-label">{item.name}</div>
            <div className="bar-container">
              <div 
                className="bar" 
                style={{ 
                  width: `${(item.count / maxCount) * 100}%`,
                  backgroundColor: item.color
                }}
              >
                <span className="bar-value">{item.count}</span>
              </div>
            </div>
          </div>
        ))
      )}
    </div>
  );
};

// График распределения по брендам
const BrandChart = ({ products, brands }) => {
  // Убеждаемся, что products - это массив
  if (!products || !Array.isArray(products)) {
    return <div className="chart-empty">Нет данных</div>;
  }

  const brandCounts = {};
  
  products.forEach(product => {
    const brandId = product.brand_id;
    brandCounts[brandId] = (brandCounts[brandId] || 0) + 1;
  });

  const brandData = brands.map(brand => ({
    name: brand.brand_name || `Бренд ${brand.id}`,
    count: brandCounts[brand.id] || 0,
    color: getColorForIndex(brand.id + 10)
  })).filter(item => item.count > 0).slice(0, 6);

  const maxCount = Math.max(...brandData.map(d => d.count), 1);

  return (
    <div className="bar-chart">
      {brandData.length === 0 ? (
        <div className="chart-empty">Нет данных</div>
      ) : (
        brandData.map((item, index) => (
          <div key={index} className="bar-item">
            <div className="bar-label">{item.name}</div>
            <div className="bar-container">
              <div 
                className="bar" 
                style={{ 
                  width: `${(item.count / maxCount) * 100}%`,
                  backgroundColor: item.color
                }}
              >
                <span className="bar-value">{item.count}</span>
              </div>
            </div>
          </div>
        ))
      )}
    </div>
  );
};

// График цен
const PriceChart = ({ products }) => {
  if (!products || !Array.isArray(products) || products.length === 0) {
    return <div className="chart-empty">Нет данных</div>;
  }

  const prices = products.map(p => parseFloat(p.price) || 0).filter(p => p > 0);
  if (prices.length === 0) {
    return <div className="chart-empty">Нет данных о ценах</div>;
  }

  const minPrice = Math.min(...prices);
  const maxPrice = Math.max(...prices);
  const avgPrice = prices.reduce((a, b) => a + b, 0) / prices.length;

  // Создаем диапазоны цен
  const ranges = [
    { label: '0-2000₽', min: 0, max: 2000, count: 0 },
    { label: '2000-5000₽', min: 2000, max: 5000, count: 0 },
    { label: '5000-10000₽', min: 5000, max: 10000, count: 0 },
    { label: '10000+₽', min: 10000, max: Infinity, count: 0 }
  ];

  prices.forEach(price => {
    for (const range of ranges) {
      if (price >= range.min && price < range.max) {
        range.count++;
        break;
      }
    }
  });

  const maxCount = Math.max(...ranges.map(r => r.count), 1);

  return (
    <div className="bar-chart">
      {ranges.map((range, index) => (
        <div key={index} className="bar-item">
          <div className="bar-label">{range.label}</div>
          <div className="bar-container">
            <div 
              className="bar" 
              style={{ 
                width: `${(range.count / maxCount) * 100}%`,
                backgroundColor: getColorForIndex(index + 20)
              }}
            >
              <span className="bar-value">{range.count}</span>
            </div>
          </div>
        </div>
      ))}
      <div className="price-stats">
        <div>Мин: {minPrice.toFixed(0)}₽</div>
        <div>Средняя: {avgPrice.toFixed(0)}₽</div>
        <div>Макс: {maxPrice.toFixed(0)}₽</div>
      </div>
    </div>
  );
};

// График заказов по дням
const OrdersChart = ({ orders }) => {
  const last7Days = [];
  const today = new Date();
  
  for (let i = 6; i >= 0; i--) {
    const date = new Date(today);
    date.setDate(date.getDate() - i);
    const dateStr = date.toISOString().split('T')[0];
    last7Days.push({ date: dateStr, count: 0 });
  }

  orders.forEach(order => {
    const orderDate = new Date(order.order_date).toISOString().split('T')[0];
    const day = last7Days.find(d => d.date === orderDate);
    if (day) {
      day.count++;
    }
  });

  const maxCount = Math.max(...last7Days.map(d => d.count), 1);

  return (
    <div className="line-chart">
      {last7Days.length === 0 ? (
        <div className="chart-empty">Нет данных</div>
      ) : (
        <svg viewBox="0 0 400 200" className="line-chart-svg">
          {/* Сетка */}
          <defs>
            <linearGradient id="lineGradient" x1="0%" y1="0%" x2="0%" y2="100%">
              <stop offset="0%" stopColor="#667eea" stopOpacity="0.3" />
              <stop offset="100%" stopColor="#667eea" stopOpacity="0.05" />
            </linearGradient>
          </defs>
          
          {/* Область графика */}
          <polygon
            fill="url(#lineGradient)"
            points={`50,150 ${last7Days.map((_, i) => `${50 + (i * 50)},${150 - (last7Days[i].count / maxCount) * 100}`).join(' ')} 350,150`}
          />
          
          {/* Линия */}
          <polyline
            fill="none"
            stroke="#667eea"
            strokeWidth="3"
            points={last7Days.map((_, i) => `${50 + (i * 50)},${150 - (last7Days[i].count / maxCount) * 100}`).join(' ')}
          />
          
          {/* Точки */}
          {last7Days.map((day, i) => (
            <g key={i}>
              <circle
                cx={50 + (i * 50)}
                cy={150 - (day.count / maxCount) * 100}
                r="5"
                fill="#667eea"
                stroke="white"
                strokeWidth="2"
              />
              <text
                x={50 + (i * 50)}
                y={180}
                fontSize="10"
                textAnchor="middle"
                fill="#6c757d"
              >
                {new Date(day.date).toLocaleDateString('ru-RU', { weekday: 'short' })}
              </text>
              <text
                x={50 + (i * 50)}
                y={150 - (day.count / maxCount) * 100 - 10}
                fontSize="11"
                textAnchor="middle"
                fill="#212529"
                fontWeight="600"
              >
                {day.count}
              </text>
            </g>
          ))}
        </svg>
      )}
    </div>
  );
};

// Вспомогательная функция для цветов
const getColorForIndex = (index) => {
  const colors = [
    '#667eea', '#764ba2', '#f093fb', '#4facfe',
    '#43e97b', '#fa709a', '#fee140', '#30cfd0',
    '#a8edea', '#fed6e3', '#d299c2', '#fef9d7'
  ];
  return colors[index % colors.length];
};

export default ManagerPanel;
