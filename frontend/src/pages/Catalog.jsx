import React, { useState, useEffect, useCallback } from 'react';
import { useAuth } from '../AuthContext';
import { useNotification } from '../components/Notification';
import { productsAPI, brandsAPI, categoriesAPI, basketAPI, favoritesAPI, reviewsAPI } from '../api';
import Pagination from '../components/Pagination';
import './Catalog.css';

const Catalog = () => {
  const [allProducts, setAllProducts] = useState([]);
  const [products, setProducts] = useState([]);
  const [brands, setBrands] = useState([]);
  const [categories, setCategories] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedBrand, setSelectedBrand] = useState('');
  const [selectedCategory, setSelectedCategory] = useState('');
  const [searchTerm, setSearchTerm] = useState('');
  const [minPrice, setMinPrice] = useState('');
  const [maxPrice, setMaxPrice] = useState('');
  const [selectedSize, setSelectedSize] = useState(''); // Выбранный размер для фильтрации
  const [sortBy, setSortBy] = useState('default'); // 'default', 'price_asc', 'price_desc', 'name_asc', 'name_desc', 'size_asc', 'size_desc'
  // basketItems и favoriteItems загружаются, но используются через API
  const [, setBasketItems] = useState([]);
  const [, setFavoriteItems] = useState([]);
  const [productsWithSizes, setProductsWithSizes] = useState([]); // Товары с информацией о размерах
  const [availableSizesList, setAvailableSizesList] = useState([]); // Список всех доступных размеров
  const [showSizeModal, setShowSizeModal] = useState(false);
  const [selectedProduct, setSelectedProduct] = useState(null);
  const [availableSizes, setAvailableSizes] = useState([]);
  const [actionType, setActionType] = useState(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [limit, setLimit] = useState(12); // 12 товаров на странице для каталога
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(1);
  const { user, isAuthenticated } = useAuth();
  const { showSuccess, showError } = useNotification();

  // Reviews states
  const [showReviewsModal, setShowReviewsModal] = useState(false);
  const [reviewsForProduct, setReviewsForProduct] = useState([]);
  const [showReviewForm, setShowReviewForm] = useState(false);
  const [reviewRating, setReviewRating] = useState(5);
  const [reviewComment, setReviewComment] = useState('');
  const [reviewProductId, setReviewProductId] = useState(null);
  const [existingReview, setExistingReview] = useState(null);
  const [reviewToEdit, setReviewToEdit] = useState(null);
  const [editModeReviewId, setEditModeReviewId] = useState(null);

  // Добавить состояние и компонент модального окна товара в основной компонент
  const [showProductInfoModal, setShowProductInfoModal] = useState(false);
  const [selectedProductForInfo, setSelectedProductForInfo] = useState(null);
  // Состояния для выбранного размера и для подсказки
  const [selectedSizeId, setSelectedSizeId] = useState(null);
  const [actionsWarning, setActionsWarning] = useState('');
  // Состояние для отслеживания открытости фильтров на мобильных
  const [filtersOpen, setFiltersOpen] = useState(false);

  // Блокировка прокрутки страницы при открытых фильтрах
  useEffect(() => {
    if (filtersOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [filtersOpen]);

  // Initialize arrays to avoid null errors
  useEffect(() => {
    if (!products) setProducts([]);
    if (!brands) setBrands([]);
    if (!categories) setCategories([]);
  }, []);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [productsRes, brandsRes, categoriesRes] = await Promise.all([
        productsAPI.getAll(currentPage, limit),
        brandsAPI.getAll(),
        categoriesAPI.getAll(),
      ]);

      // Обработка нового формата ответа с пагинацией
      let productsData = [];
      if (productsRes.data && productsRes.data.data) {
        productsData = productsRes.data.data || [];
        setTotal(productsRes.data.total || 0);
        setTotalPages(productsRes.data.total_pages || 1);
      } else {
        // Fallback для старого формата (массив)
        productsData = productsRes.data || [];
        setTotal(productsData.length || 0);
        setTotalPages(1);
      }

      console.log('📦 Товары с API:', productsData);
      console.log('🏷️ Бренды с API:', brandsRes.data);
      console.log('📁 Категории с API:', categoriesRes.data);

      setAllProducts(productsData);
      setBrands(brandsRes.data);
      setCategories(categoriesRes.data);
      
      // Загружаем размеры для товаров текущей страницы
      const productsWithSizesData = await Promise.all(
        productsData.map(async (product) => {
          try {
            const sizesRes = await productsAPI.getSizes(product.id);
            const sizes = sizesRes.data || [];
            const sizeValues = sizes.map(s => s.size).filter(s => s !== null && s !== undefined);
            // Сохраняем все размеры, которые есть в наличии (quantity > 0)
            const availableSizes = sizes
              .filter(s => s.quantity > 0)
              .map(s => s.size);
            return {
              ...product,
              sizes: availableSizes, // Все доступные размеры
              allSizes: sizeValues, // Все размеры (включая без наличия)
              minSize: sizeValues.length > 0 ? Math.min(...sizeValues) : null,
              maxSize: sizeValues.length > 0 ? Math.max(...sizeValues) : null,
            };
          } catch (error) {
            console.warn(`Не удалось загрузить размеры для товара ${product.id}:`, error);
            return {
              ...product,
              sizes: [],
              allSizes: [],
              minSize: null,
              maxSize: null,
            };
          }
        })
      );
      
      setProductsWithSizes(productsWithSizesData);
      
      // Для фильтра по размерам нужно загрузить все товары (или кешировать список размеров)
      // Пока используем только размеры текущей страницы для фильтра
      const allSizesSet = new Set();
      productsWithSizesData.forEach(product => {
        if (product.sizes && product.sizes.length > 0) {
          product.sizes.forEach(size => allSizesSet.add(size));
        }
      });
      const sortedSizes = Array.from(allSizesSet).sort((a, b) => a - b);
      setAvailableSizesList(sortedSizes);
      
      console.log('✅ Загружено:', {
        товаров: productsData.length,
        брендов: brandsRes.data.length,
        категорий: categoriesRes.data.length,
        страница: currentPage,
        всего: total || productsData.length
      });
    } catch (error) {
      console.error('Ошибка загрузки данных:', error);
      console.error('Детали:', error.response?.data || error.message);
      showError('Не удалось загрузить данные. Проверьте что backend запущен на http://localhost:8080');
    } finally {
      setLoading(false);
    }
  }, [currentPage, limit, total, showError]);

  const loadBasketAndFavorites = useCallback(async () => {
    if (!isAuthenticated() || !user) return;
    
    try {
      const [basketRes, favRes] = await Promise.all([
        basketAPI.getByUserId(user.id),
        favoritesAPI.getByUserId(user.id)
      ]);
      
      setBasketItems(basketRes.data.items || []);
      setFavoriteItems(favRes.data.items || []);
    } catch (error) {
      console.error('Ошибка загрузки корзины и избранного:', error);
    }
  }, [user, isAuthenticated]);

  // Сбрасываем страницу при изменении фильтров
  useEffect(() => {
    if (selectedBrand || selectedCategory || selectedSize || searchTerm || minPrice || maxPrice || sortBy !== 'default') {
      setCurrentPage(1);
    }
  }, [selectedBrand, selectedCategory, selectedSize, searchTerm, minPrice, maxPrice, sortBy]);

  const applyFilters = useCallback(() => {
    // Используем товары с размерами для фильтрации и сортировки
    let filteredProducts = productsWithSizes.length > 0 ? [...productsWithSizes] : [...allProducts];

    if (selectedBrand) {
      filteredProducts = filteredProducts.filter(p => p.brand_id === parseInt(selectedBrand));
    }

    if (selectedCategory) {
      filteredProducts = filteredProducts.filter(p => p.category_id === parseInt(selectedCategory));
    }

    if (searchTerm) {
      filteredProducts = filteredProducts.filter(p => 
        p.name.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    if (minPrice) {
      filteredProducts = filteredProducts.filter(p => p.price >= parseFloat(minPrice));
    }
    if (maxPrice) {
      filteredProducts = filteredProducts.filter(p => p.price <= parseFloat(maxPrice));
    }

    // Фильтр по размеру
    if (selectedSize) {
      const sizeNum = parseInt(selectedSize);
      filteredProducts = filteredProducts.filter(p => {
        // Проверяем, есть ли выбранный размер в доступных размерах товара
        return p.sizes && p.sizes.includes(sizeNum);
      });
    }

    // Применяем сортировку
    if (sortBy !== 'default') {
      filteredProducts.sort((a, b) => {
        switch (sortBy) {
          case 'price_asc':
            return a.price - b.price;
          case 'price_desc':
            return b.price - a.price;
          case 'name_asc':
            return a.name.localeCompare(b.name, 'ru');
          case 'name_desc':
            return b.name.localeCompare(a.name, 'ru');
          case 'size_asc':
            // Сортировка по минимальному размеру (от меньшего к большему)
            const aMinSize = a.minSize !== null && a.minSize !== undefined ? a.minSize : Infinity;
            const bMinSize = b.minSize !== null && b.minSize !== undefined ? b.minSize : Infinity;
            if (aMinSize === Infinity && bMinSize === Infinity) return 0;
            if (aMinSize === Infinity) return 1; // Товары без размеров в конец
            if (bMinSize === Infinity) return -1;
            return aMinSize - bMinSize;
          case 'size_desc':
            // Сортировка по максимальному размеру (от большего к меньшему)
            const aMaxSize = a.maxSize !== null && a.maxSize !== undefined ? a.maxSize : -Infinity;
            const bMaxSize = b.maxSize !== null && b.maxSize !== undefined ? b.maxSize : -Infinity;
            if (aMaxSize === -Infinity && bMaxSize === -Infinity) return 0;
            if (aMaxSize === -Infinity) return 1; // Товары без размеров в конец
            if (bMaxSize === -Infinity) return -1;
            return bMaxSize - aMaxSize;
          default:
            return 0;
        }
      });
    }

    setProducts(filteredProducts);
  }, [allProducts, productsWithSizes, selectedBrand, selectedCategory, searchTerm, minPrice, maxPrice, selectedSize, sortBy]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  useEffect(() => {
    applyFilters();
  }, [applyFilters]);

  useEffect(() => {
    if (isAuthenticated() && user) {
      loadBasketAndFavorites();
    }
  }, [isAuthenticated, user, loadBasketAndFavorites]);

  const openSizeModal = async (productId, action) => {
    if (!isAuthenticated()) {
      showError('Для добавления необходимо войти в систему');
      return;
    }

    try {
      const sizesRes = await productsAPI.getSizes(productId);
      const sizes = sizesRes.data;

      if (!sizes || sizes.length === 0) {
        showError('Нет доступных размеров для этого товара');
        return;
      }

      setSelectedProduct(productId);
      setAvailableSizes(sizes);
      setActionType(action);
      setShowSizeModal(true);
    } catch (error) {
      console.error('Ошибка загрузки размеров:', error);
      showError('Не удалось загрузить размеры товара');
    }
  };

  const handleSizeSelected = async (productSizeId) => {
    try {
      if (actionType === 'basket') {
        await basketAPI.add({
          user_id: user.id,
          product_size_id: productSizeId,
          quantity: 1,
        });
        // Товар добавлен в корзину
      } else if (actionType === 'favorite') {
        await favoritesAPI.add({
          user_id: user.id,
          product_size_id: productSizeId,
        });
        // Товар добавлен в избранное
      }

      await loadBasketAndFavorites();
      setShowSizeModal(false);
      showSuccess(`Товар добавлен в ${actionType === 'basket' ? 'корзину' : 'избранное'}`);
    } catch (error) {
      console.error('Ошибка добавления:', error);
      showError('Не удалось добавить товар');
    }
  };

  const openReviewsModal = async (productId) => {
    setReviewProductId(productId);
    setShowReviewsModal(true);
    setShowReviewForm(false);
    setExistingReview(null);
    setReviewToEdit(null);
    
    try {
      const response = await reviewsAPI.getByProductId(productId);
      // API возвращает пагинированный ответ с полем data
      const reviews = (response.data && response.data.data) ? response.data.data : (Array.isArray(response.data) ? response.data : []);
      setReviewsForProduct(reviews);
      
      // Проверяем, есть ли уже отзыв от текущего пользователя
      if (user && user.id) {
        const myReview = reviews.find(r => r.user_id === user.id);
        if (myReview) {
          setExistingReview(myReview);
        }
      }
    } catch (error) {
      console.error('Ошибка загрузки отзывов:', error);
      setReviewsForProduct([]);
    }
  };

  const handleEditReview = (review) => {
    setReviewToEdit(review);
    setReviewRating(review.rating);
    setReviewComment(review.comment);
    setShowReviewForm(true);
  };

  const handleSubmitReview = async (e) => {
    e.preventDefault();
    
    if (!reviewComment.trim()) {
      showError('Введите комментарий');
      return;
    }

    try {
      if (reviewToEdit) {
        // Обновляем существующий отзыв
        await reviewsAPI.update(reviewToEdit.id, {
          rating: reviewRating,
          comment: reviewComment,
        });
        showSuccess('Отзыв успешно обновлен');
      } else {
        // Создаем новый отзыв
        await reviewsAPI.create({
          product_id: reviewProductId,
          rating: reviewRating,
          comment: reviewComment,
        });
        showSuccess('Отзыв успешно добавлен');
      }
      
      setShowReviewForm(false);
      setReviewComment('');
      setReviewRating(5);
      setReviewToEdit(null);
      setEditModeReviewId(null);
      
      // Обновляем список отзывов
      const response = await reviewsAPI.getByProductId(reviewProductId);
      // API возвращает пагинированный ответ с полем data
      const reviews = (response.data && response.data.data) ? response.data.data : (Array.isArray(response.data) ? response.data : []);
      setReviewsForProduct(reviews);
      
      // Обновляем отзывы в модальном окне товара, если оно открыто
      if (selectedProductForInfo && selectedProductForInfo.id === reviewProductId) {
        setSelectedProductForInfo({
          ...selectedProductForInfo,
          reviews: reviews
        });
      }
      
      // Обновляем existingReview
      if (user && user.id) {
        const myReview = reviews.find(r => r.user_id === user.id);
        setExistingReview(myReview);
      }
    } catch (error) {
      console.error('Ошибка создания/обновления отзыва:', error);
      showError(error.response?.data || 'Не удалось сохранить отзыв');
    }
  };

  const openProductInfoModal = async (product) => {
    let sizes = [];
    let reviews = [];
    try {
      const sizesRes = await productsAPI.getSizes(product.id);
      sizes = Array.isArray(sizesRes.data) ? sizesRes.data : [];
    } catch {}
    try {
      const reviewsRes = await reviewsAPI.getByProductId(product.id);
      console.log('📝 Reviews response:', reviewsRes.data);
      // API возвращает пагинированный ответ с полем data
      reviews = (reviewsRes.data && reviewsRes.data.data) ? reviewsRes.data.data : (Array.isArray(reviewsRes.data) ? reviewsRes.data : []);
      console.log('📝 Extracted reviews:', reviews);
    } catch (error) {
      console.error('❌ Error loading reviews:', error);
    }
    // Убеждаемся, что reviews всегда массив
    const safeReviews = Array.isArray(reviews) ? reviews : [];
    
    // Получаем названия бренда и категории
    const brand = brands.find(b => b.id === product.brand_id);
    const category = categories.find(c => c.id === product.category_id);
    
    // Извлекаем информацию о материале из названия или описания
    const materialKeywords = {
      'кожа': 'Кожа',
      'замша': 'Замша',
      'текстиль': 'Текстиль',
      'синтетика': 'Синтетика',
      'нубук': 'Нубук',
      'mesh': 'Сетка',
      'leather': 'Кожа',
      'suede': 'Замша',
      'textile': 'Текстиль',
      'synthetic': 'Синтетика',
      'nubuck': 'Нубук'
    };
    
    let detectedMaterial = null;
    const productNameLower = (product.name || '').toLowerCase();
    const productDescLower = ((product.description || '') + ' ' + (product.name || '')).toLowerCase();
    
    for (const [keyword, material] of Object.entries(materialKeywords)) {
      if (productNameLower.includes(keyword) || productDescLower.includes(keyword)) {
        detectedMaterial = material;
        break;
      }
    }
    
    setSelectedProductForInfo({
      ...product, 
      sizes, 
      reviews: safeReviews,
      brand_name: brand ? brand.brand_name : null,
      category_name: category ? category.category_name : null,
      material: detectedMaterial
    });
    setShowProductInfoModal(true);
    setSelectedSizeId(null); // Сбрасываем выбранный размер при открытии модалки
    setActionsWarning(''); // Сбрасываем варнинг при открытии модалки
  };

  const closeProductInfoModal = () => {
    setShowProductInfoModal(false);
    setSelectedProductForInfo(null);
    setSelectedSizeId(null); // Сбрасываем выбранный размер при закрытии модалки
    setActionsWarning(''); // Сбрасываем варнинг при закрытии модалки
  };

  if (loading) {
    return (
      <div className="catalog-loading">
        <div className="spinner"></div>
        <p>Загрузка каталога...</p>
      </div>
    );
  }

  return (
    <div className="catalog-container">
      <button 
        className="mobile-filters-toggle"
        onClick={() => setFiltersOpen(!filtersOpen)}
      >
        <span>{filtersOpen ? '✕' : '☰'}</span>
        <span>Фильтры</span>
      </button>
      {filtersOpen && (
        <div 
          className="filters-overlay"
          onClick={() => setFiltersOpen(false)}
        />
      )}
      <div 
        className={`catalog-filters ${filtersOpen ? 'filters-open' : ''}`}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="filters-header">
          <h2>Фильтры</h2>
          <button 
            className="mobile-filters-close"
            onClick={() => setFiltersOpen(false)}
            aria-label="Закрыть фильтры"
          >
            ✕
          </button>
        </div>
        
        <div className="filter-group">
          <label>Поиск</label>
          <input
            type="text"
            placeholder="Название товара..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="search-input"
          />
        </div>

        <div className="filter-group">
          <label>Цена от</label>
          <input
            type="number"
            placeholder="Мин"
            value={minPrice}
            onChange={(e) => setMinPrice(e.target.value)}
            className="price-input"
          />
        </div>

        <div className="filter-group">
          <label>Цена до</label>
          <input
            type="number"
            placeholder="Макс"
            value={maxPrice}
            onChange={(e) => setMaxPrice(e.target.value)}
            className="price-input"
          />
        </div>

        <div className="filter-group">
          <label>Бренд</label>
          <select
            value={selectedBrand}
            onChange={(e) => setSelectedBrand(e.target.value)}
          >
            <option value="">Все бренды</option>
            {brands.map(brand => (
              <option key={brand.id} value={brand.id}>
                {brand.brand_name}
              </option>
            ))}
          </select>
        </div>

        <div className="filter-group">
          <label>Категория</label>
          <select
            value={selectedCategory}
            onChange={(e) => setSelectedCategory(e.target.value)}
          >
            <option value="">Все категории</option>
            {categories.map(category => (
              <option key={category.id} value={category.id}>
                {category.category_name}
              </option>
            ))}
          </select>
        </div>

        <div className="filter-group">
          <label>Размер</label>
          <select
            value={selectedSize}
            onChange={(e) => setSelectedSize(e.target.value)}
          >
            <option value="">Все размеры</option>
            {availableSizesList.map(size => (
              <option key={size} value={size}>
                {size}
              </option>
            ))}
          </select>
        </div>

        <div className="filter-group">
          <label>Сортировка</label>
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value)}
          >
            <option value="default">По умолчанию</option>
            <option value="price_asc">По цене: от дешевых</option>
            <option value="price_desc">По цене: от дорогих</option>
            <option value="name_asc">По названию: А-Я</option>
            <option value="name_desc">По названию: Я-А</option>
            <option value="size_asc">По размеру: от меньшего</option>
            <option value="size_desc">По размеру: от большего</option>
          </select>
        </div>

        {(selectedBrand || selectedCategory || selectedSize || searchTerm || minPrice || maxPrice || sortBy !== 'default') && (
          <button
            className="clear-filters"
            onClick={() => {
              setSelectedBrand('');
              setSelectedCategory('');
              setSelectedSize('');
              setSearchTerm('');
              setMinPrice('');
              setMaxPrice('');
              setSortBy('default');
            }}
          >
            Сбросить все фильтры
          </button>
        )}
      </div>

      <div className="catalog-grid">
        {(!products || products.length === 0) ? (
          <div className="no-products">
            <p>📭 Товары не найдены</p>
            <p style={{fontSize: '0.9em', color: 'var(--text-secondary)', marginTop: '10px'}}>
              {brands.length === 0 && categories.length === 0 
                ? 'База данных пуста. Добавьте товары через админ-панель или Swagger API.'
                : 'Попробуйте сбросить фильтры.'}
            </p>
          </div>
        ) : (
          products.map(product => (
            <div key={product.id} className="product-card" onClick={() => openProductInfoModal(product)} style={{cursor: 'pointer'}}>
              {product.image_url ? (
                <img
                  src={product.image_url.split(',')[0].trim()}
                  alt={product.name}
                  className="product-image"
                />
              ) : (
                  <div className="product-image-placeholder">
                  <span>Нет фото</span>
                </div>
              )}
              <h3 className="product-name">{product.name}</h3>
              {(product.reviews && product.reviews.length > 0) && (
                <div className="product-mini-rating">
                  {'★'.repeat(Math.round(product.reviews.reduce((a,b)=>a+b.rating,0)/product.reviews.length))}
                  <span style={{marginLeft:4, color:'#888'}}>({product.reviews.length})</span>
                </div>
              )}
              <p className="product-price">{product.price} ₽</p>
              {isAuthenticated() && (
                <div className="product-actions">
                  <button className="btn-add-cart" onClick={e => {e.stopPropagation(); openSizeModal(product.id, 'basket')}}>
                    В корзину
                  </button>
                  <button className="btn-add-fav" onClick={e => {e.stopPropagation(); openSizeModal(product.id, 'favorite')}}>
                    В избранное
                  </button>
                </div>
              )}
            </div>
          ))
        )}
      </div>

      {totalPages > 1 && (
        <Pagination
          currentPage={currentPage}
          totalPages={totalPages}
          onPageChange={(page) => {
            setCurrentPage(page);
            window.scrollTo({ top: 0, behavior: 'smooth' });
          }}
          limit={limit}
          total={total}
        />
      )}

      {showSizeModal && (
        <div className="modal-overlay" onClick={() => setShowSizeModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h3>Выберите размер</h3>
            <div className="size-grid">
              {availableSizes.map((size) => (
                <button
                  key={size.id}
                  className="size-button"
                  onClick={() => handleSizeSelected(size.id)}
                  disabled={size.quantity === 0}
                >
                  {size.size}
                  {size.quantity === 0 && ' (нет в наличии)'}
                </button>
              ))}
            </div>
            <button 
              className="modal-close" 
              onClick={() => setShowSizeModal(false)}
            >
              Отмена
            </button>
          </div>
        </div>
      )}

      {/* Reviews Modal */}
      {showReviewsModal && (
        <div className="modal-overlay" onClick={() => setShowReviewsModal(false)}>
          <div className="modal-content modal-reviews" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Отзывы о товаре</h3>
              <button className="modal-close" onClick={() => setShowReviewsModal(false)}>×</button>
            </div>
            
            {showReviewForm ? (
              <form onSubmit={handleSubmitReview}>
                <h4 style={{ marginBottom: '1rem' }}>
                  {reviewToEdit ? 'Изменить отзыв' : 'Оставить отзыв'}
                </h4>
                <div className="review-form-group">
                  <label>Рейтинг</label>
                  <select value={reviewRating} onChange={(e) => setReviewRating(parseInt(e.target.value))}>
                    <option value={1}>⭐ 1</option>
                    <option value={2}>⭐⭐ 2</option>
                    <option value={3}>⭐⭐⭐ 3</option>
                    <option value={4}>⭐⭐⭐⭐ 4</option>
                    <option value={5}>⭐⭐⭐⭐⭐ 5</option>
                  </select>
                </div>

                <div className="review-form-group">
                  <label>Комментарий</label>
                  <textarea
                    value={reviewComment}
                    onChange={(e) => setReviewComment(e.target.value)}
                    placeholder="Ваш отзыв..."
                    required
                    rows={4}
                  />
                </div>

                <div className="modal-actions">
                  <button type="submit" className="btn-primary">
                    {reviewToEdit ? 'Сохранить изменения' : 'Отправить отзыв'}
                  </button>
                  <button 
                    type="button" 
                    className="btn-secondary" 
                    onClick={() => {
                      setShowReviewForm(false);
                      setReviewToEdit(null);
                      setReviewRating(5);
                      setReviewComment('');
                    }}
                  >
                    Отмена
                  </button>
                </div>
              </form>
            ) : (
              <>
                <div className="reviews-list">
                  {(!reviewsForProduct || reviewsForProduct.length === 0) ? (
                    <p className="no-reviews">Пока нет отзывов</p>
                  ) : (
                    reviewsForProduct.map((review) => (
                      <div key={review.id} className="review-item">
                        <div className="review-header">
                          <div className="review-rating">
                            {'⭐'.repeat(review.rating)}
                          </div>
                          <span className="review-date">
                            {new Date(review.date).toLocaleDateString('ru-RU')}
                          </span>
                        </div>
                        <p className="review-comment">{review.comment}</p>
                        {user && review.user_id === user.id && (
                          <button 
                            className="btn-edit-review"
                            onClick={() => handleEditReview(review)}
                          >
                            Изменить
                          </button>
                        )}
                      </div>
                    ))
                  )}
                </div>

                {isAuthenticated() && (
                  existingReview ? (
                    <p className="review-info">Вы уже оставили отзыв на этот товар</p>
                  ) : (
                    <button className="btn-primary" onClick={() => {
                      setReviewToEdit(null);
                      setReviewRating(5);
                      setReviewComment('');
                      setShowReviewForm(true);
                    }}>
                      Оставить отзыв
                    </button>
                  )
                )}
              </>
            )}
          </div>
        </div>
      )}

      {/* Product Info Modal */}
      {showProductInfoModal && selectedProductForInfo && (
        <div className="modal-overlay" onClick={closeProductInfoModal}>
          <div className="modal-content modal-product-info-wide" onClick={e=>e.stopPropagation()} style={{minWidth:860,maxWidth:1120,display:'flex',gap:'2.7rem',padding:'2.7rem 3.4rem',alignItems:'flex-start',position:'relative'}}>
            <button 
              className="modal-close" 
              onClick={closeProductInfoModal}
              style={{
                position:'absolute',
                top:'1.5rem',
                right:'1.5rem',
                width:'36px',
                height:'36px',
                borderRadius:'50%',
                border:'none',
                background:'var(--bg-secondary)',
                color:'var(--text-primary)',
                fontSize:'1.5rem',
                cursor:'pointer',
                display:'flex',
                alignItems:'center',
                justifyContent:'center',
                boxShadow:'0 2px 8px var(--shadow)',
                zIndex:10,
                transition:'background 0.2s'
              }}
              onMouseEnter={(e) => e.target.style.background = 'var(--bg-tertiary)'}
              onMouseLeave={(e) => e.target.style.background = 'var(--bg-secondary)'}
            >
              ×
            </button>
            <div className="modal-picside" style={{flex:'0 0 440px',display:'flex',flexDirection:'column',alignItems:'center',minWidth:320,maxWidth:440}}>
              <div className="modal-main-image-container" style={{width:'100%',maxWidth:420,height:360,minHeight:360,display:'flex',alignItems:'center',justifyContent:'center',borderRadius:14,marginBottom:30,boxShadow:'0 1px 12px var(--shadow)',position:'relative',overflow:'hidden'}}>
                <img 
                  src={selectedProductForInfo.selectedImage||(selectedProductForInfo.image_url?selectedProductForInfo.image_url.split(',')[0].trim():'')} 
                  alt="Product" 
                  style={{
                    width:'100%',
                    height:'100%',
                    objectFit:'contain',
                    display:'block',
                    position:'absolute',
                    top:0,
                    left:0,
                    opacity:0,
                    transition:'opacity 0.3s ease'
                  }}
                  onLoad={(e) => {
                    e.target.style.opacity = '1';
                  }}
                  onError={(e) => {
                    e.target.style.opacity = '0';
                  }}
                />
              </div>
              <div style={{display:'flex',gap:'16px',marginTop:0,marginBottom:24,flexWrap:'wrap',justifyContent:'center'}}>
                {(selectedProductForInfo.image_url ? selectedProductForInfo.image_url.split(',') : []).map((url,idx) => (
                    <div key={url+idx} style={{width:'70px',height:'70px',flexShrink:0,position:'relative'}}>
                      <img 
                        onClick={()=>{
                          const newObj={...selectedProductForInfo,selectedImage:url.trim()};
                          setSelectedProductForInfo(newObj);
                        }}
                        src={url.trim()} 
                        alt="preview"
                        style={{
                          width:'70px',
                          height:'70px',
                          objectFit:'cover',
                          borderRadius:'7px',
                          border:'2px solid '+((selectedProductForInfo.selectedImage?selectedProductForInfo.selectedImage:url.trim())===url.trim()?'#667eea':'#eee'),
                          cursor:'pointer',
                          boxShadow:'0 1px 4px #eee',
                          display:'block',
                          opacity:0,
                          transition:'opacity 0.3s ease',
                          position:'absolute',
                          top:0,
                          left:0
                        }}
                        onLoad={(e) => {
                          e.target.style.opacity = '1';
                        }}
                        onError={(e) => {
                          e.target.style.opacity = '0';
                        }}
                      />
                    </div>
                ))}
              </div>
            </div>
            <div className="modal-infoside" style={{flex:'1 1',minWidth:220,display:'flex',flexDirection:'column',gap:'1.5rem',maxWidth:'420px'}}>
              <h3 style={{marginBottom:'.8rem',fontSize:'1.45rem',color:'var(--text-primary)'}}>{selectedProductForInfo.name}</h3>
              <div style={{display:'flex',alignItems:'center',marginBottom:10,gap:'12px'}}>
                {selectedProductForInfo.brand_name && <span style={{background:'#f3f6fa',color:'#578',fontWeight:600,borderRadius:7,padding:'3px 13px',fontSize:'0.98em'}}>{selectedProductForInfo.brand_name}</span>}
                {selectedProductForInfo.category_name && <span style={{background:'#f2f4ee',color:'#0a534e',fontWeight:500,borderRadius:7,padding:'3px 13px',fontSize:'0.98em'}}>{selectedProductForInfo.category_name}</span>}
              </div>
              <div style={{color:'var(--text-primary)',fontWeight:700,fontSize:'2.1rem',marginBottom:17}}>{selectedProductForInfo.price} ₽</div>
              {selectedProductForInfo.description && <div style={{background:'var(--bg-secondary)',color:'var(--text-secondary)',borderRadius:8,padding:'10px 15px',marginBottom:12,fontSize:'1.06em',lineHeight:'1.6'}}>{selectedProductForInfo.description}</div>}
              {/* Дополнительная информация о товаре */}
              <div style={{background:'var(--bg-secondary)',borderRadius:8,padding:'14px 16px',marginBottom:12,fontSize:'0.98em',lineHeight:'1.7',border:'1px solid var(--border-color)'}}>
                <div style={{display:'flex',flexDirection:'column',gap:'10px',color:'var(--text-secondary)'}}>
                  {selectedProductForInfo.brand_name && (
                    <div style={{display:'flex',alignItems:'center',gap:'10px'}}>
                      <span style={{fontSize:'1.2em',opacity:0.7}}>🏷️</span>
                      <strong style={{color:'var(--text-primary)',minWidth:'90px'}}>Бренд:</strong> 
                      <span style={{color:'var(--text-primary)',fontWeight:500}}>{selectedProductForInfo.brand_name}</span>
                    </div>
                  )}
                  {selectedProductForInfo.category_name && (
                    <div style={{display:'flex',alignItems:'center',gap:'10px'}}>
                      <span style={{fontSize:'1.2em',opacity:0.7}}>📁</span>
                      <strong style={{color:'var(--text-primary)',minWidth:'90px'}}>Категория:</strong> 
                      <span style={{color:'var(--text-primary)',fontWeight:500}}>{selectedProductForInfo.category_name}</span>
                    </div>
                  )}
                  {selectedProductForInfo.material && (
                    <div style={{display:'flex',alignItems:'center',gap:'10px'}}>
                      <span style={{fontSize:'1.2em',opacity:0.7}}>🧵</span>
                      <strong style={{color:'var(--text-primary)',minWidth:'90px'}}>Материал:</strong> 
                      <span style={{color:'var(--text-primary)',fontWeight:500}}>{selectedProductForInfo.material}</span>
                    </div>
                  )}
                  {selectedProductForInfo.reviews && Array.isArray(selectedProductForInfo.reviews) && selectedProductForInfo.reviews.length > 0 && (
                    <div style={{display:'flex',alignItems:'center',gap:'10px',flexWrap:'wrap'}}>
                      <span style={{fontSize:'1.2em',opacity:0.7}}>⭐</span>
                      <strong style={{color:'var(--text-primary)',minWidth:'90px'}}>Рейтинг:</strong> 
                      <span style={{color:'#ffb400',fontSize:'1.1em',fontWeight:700}}>{'★'.repeat(Math.round(selectedProductForInfo.reviews.reduce((a,b)=>a+(b.rating||0),0)/selectedProductForInfo.reviews.length))}</span>
                      <span style={{color:'var(--text-secondary)',marginLeft:4}}>({selectedProductForInfo.reviews.length} {selectedProductForInfo.reviews.length === 1 ? 'отзыв' : selectedProductForInfo.reviews.length < 5 ? 'отзыва' : 'отзывов'})</span>
                    </div>
                  )}
                  {(selectedProductForInfo.sizes && Array.isArray(selectedProductForInfo.sizes) && selectedProductForInfo.sizes.length > 0) && (
                    <div style={{display:'flex',alignItems:'center',gap:'10px',flexWrap:'wrap'}}>
                      <span style={{fontSize:'1.2em',opacity:0.7}}>📏</span>
                      <strong style={{color:'var(--text-primary)',minWidth:'90px'}}>Размеры:</strong> 
                      <span style={{color:'var(--text-primary)',fontWeight:500}}>{selectedProductForInfo.sizes.map(s => s.size || s).filter((v,i,a)=>a.indexOf(v)===i).sort((a,b)=>a-b).join(', ')}</span>
                    </div>
                  )}
                </div>
              </div>
              {(selectedProductForInfo.sizes||[]).length>0 && (
                <div style={{margin:'18px 0 36px'}}>
                  <div style={{marginBottom:7,fontWeight:600,color:'var(--text-primary)'}}>
                    {isAuthenticated() ? 'Выберите размер:' : 'Доступные размеры:'}
                  </div>
                  <div style={{display:'flex',flexWrap:'wrap',gap:'12px'}}>
                  {selectedProductForInfo.sizes.map(size=>(
                    <button key={size.id||size.size} type="button"
                      style={{padding:'11px 18px',borderRadius:9,border:selectedSizeId===size.id?'2.4px solid #667eea':'1.5px solid var(--border-color)',fontWeight:600,fontSize:'1.11em',background:selectedSizeId===size.id?'rgba(102, 126, 234, 0.1)':'var(--bg-secondary)',color:'var(--text-primary)',outline:'none',cursor:!isAuthenticated()||size.quantity===0?'not-allowed':'pointer',opacity:size.quantity===0?0.3:1,minWidth:48,transition:'box-shadow 0.17s'}}
                      onClick={()=>isAuthenticated()&&size.quantity>0&&setSelectedSizeId(size.id)} disabled={!isAuthenticated()||size.quantity===0}>
                      {size.size}
                    </button>
                  ))}
                  </div>
                  {actionsWarning && <div style={{color:'#da3e3e',fontSize:'0.99em',marginTop:'6px'}}>{actionsWarning}</div>}
                      {!isAuthenticated() && (
                    <div style={{color:'var(--primary-color)',fontSize:'0.95em',marginTop:'10px',fontWeight:500}}>
                      Войдите в систему, чтобы добавить товар в корзину или избранное
                    </div>
                  )}
                </div>
              )}
              {isAuthenticated() && (
                <div className="product-info-actions" style={{display:'flex',gap:'1.3rem',margin:'12px 0 30px'}}>
                  <button className="btn-primary" style={{minWidth:160,fontSize:'1.11em',padding:'14px 0'}} onClick={()=>{
                    if (!selectedSizeId) {setActionsWarning('Пожалуйста, выберите размер'); return;}
                    setActionsWarning('');
                    basketAPI.add({user_id:user.id,product_size_id:selectedSizeId,quantity:1}).then(()=>{showSuccess('Добавлено в корзину')});
                  }} 
                  disabled={!selectedSizeId}>
                  В корзину
                  </button>
                  <button className="btn-add-fav" style={{minWidth:130,fontSize:'1.01em',padding:'14px 0'}} onClick={()=>{
                    if (!selectedSizeId) {setActionsWarning('Пожалуйста, выберите размер'); return;}
                    setActionsWarning('');
                    favoritesAPI.add({user_id:user.id,product_size_id:selectedSizeId}).then(()=>{showSuccess('Добавлено в избранное')});
                  }} 
                  disabled={!selectedSizeId}>
                  В избранное
                  </button>
                </div>
              )}
              <div className="product-info-reviews" style={{marginTop:28,paddingTop:18,borderTop:'2px solid var(--border-color)'}}>
                <h4 style={{marginBottom:8,fontWeight:600,color:'var(--text-primary)'}}>Отзывы покупателей</h4>
                  {(!selectedProductForInfo.reviews || !Array.isArray(selectedProductForInfo.reviews) || selectedProductForInfo.reviews.length === 0) ? (
                  <div style={{display:'flex',flexDirection:'column',alignItems:'center',padding:'25px 5px'}}>
                    <div style={{color:'var(--text-secondary)',fontSize:'1.1em'}}>Пока нет ни одного отзыва</div>
                  </div>
                ) : (
                   <div className="reviews-list" style={{maxHeight:175,overflowY:'auto',marginBottom:15,gap:'1.2em',display:'flex',flexDirection:'column',paddingRight:4,scrollbarWidth:'none',msOverflowStyle:'none'}}>
                      {(Array.isArray(selectedProductForInfo.reviews) ? selectedProductForInfo.reviews : []).map(r=>(
                        <div key={r.id} style={{background:'var(--bg-secondary)',borderRadius:9,padding:'16px 17px',boxShadow:'0 2px 7px var(--shadow)',display:'flex',gap:14,alignItems:'flex-start',marginBottom:2,position:'relative'}}>
                          <div style={{marginTop:3}}></div>
                          <div style={{flex:'1 1',minWidth:0}}>
                            <div style={{display:'flex',gap:8,alignItems:'center',marginBottom:5,flexWrap:'wrap'}}>
                              <span style={{color:'#ffb400',fontSize:'1.35em',fontWeight:700}}>{'★'.repeat(r.rating||0)}{'☆'.repeat(5-(r.rating||0))}</span>
                              <span style={{fontSize:'0.96em',color:'var(--text-secondary)',marginLeft:4,marginTop:2}}>{new Date(r.date).toLocaleDateString('ru-RU')}</span>
                              {user && r.user_id === user.id && (
                                <button style={{marginLeft:10,padding:'3px 10px',fontSize:'0.99em',border:'none',background:'#ede8fb',color:'#5a3eb0',borderRadius:6,cursor:'pointer',transition:'background .17s'}} onClick={()=>{
                                  setReviewProductId(selectedProductForInfo.id);
                                  setEditModeReviewId(r.id);
                                  setReviewRating(r.rating);
                                  setReviewComment(r.comment);
                                  setShowReviewForm(true);
                                  setReviewToEdit(r);
                                }}>Изменить</button>
                              )}
                            </div>
                            <div style={{fontSize:'1.09em',color:'var(--text-primary)',wordBreak:'break-word',lineHeight:'1.55'}}>{r.comment}</div>
                          </div>
                        </div>
                      ))}
                   </div>
                )}
                {isAuthenticated() && (
                  <button style={{marginTop:8,background:'var(--bg-tertiary)',color:'var(--text-primary)',padding:'10px 28px',border:'none',borderRadius:'8px',fontWeight:'600',fontSize:'1.07em'}} onClick={()=>{
                    setReviewProductId(selectedProductForInfo.id);
                    setShowReviewForm(true);
                    setReviewToEdit(null);
                    setEditModeReviewId(null);
                    setReviewComment('');
                    setReviewRating(5);
                  }}>Оставить отзыв</button>
                )}
                {showReviewForm && isAuthenticated() && (
                  <form onSubmit={handleSubmitReview} style={{marginTop:15,marginBottom:8,background:'var(--bg-secondary)',padding:'17px',borderRadius:'10px',boxShadow:'0 1px 7px var(--shadow)',display:'flex',flexDirection:'column',gap:'11px'}}>
                    <label style={{fontWeight:500,color:'var(--text-primary)'}}>Рейтинг:
                      <select value={reviewRating} onChange={e=>setReviewRating(parseInt(e.target.value))} style={{marginLeft:'12px',fontSize:'1.18em',borderRadius:'6px',border:'1px solid var(--border-color)',background:'var(--bg-primary)',color:'var(--text-primary)',padding:'4px 20px 4px 10px'}}>
                        {[1,2,3,4,5].map(v=>(<option key={v} value={v}>{'★'.repeat(v)}</option>))}
                      </select>
                    </label>
                    <textarea value={reviewComment} onChange={e=>setReviewComment(e.target.value)} rows={3} required placeholder="Ваш текст отзыва..." style={{width:'100%',background:'var(--bg-primary)',color:'var(--text-primary)',border:'1.5px solid var(--border-color)',borderRadius:'6px',fontSize:'1.09em',padding:'10px'}}/>
                    <div style={{display:'flex',gap:'1.3rem'}}>
                      <button className="btn-primary" type="submit" style={{fontWeight:600,fontSize:'1.09em'}}>{editModeReviewId ? 'Сохранить изменения' : 'Отправить'}</button>
                      <button type="button" className="btn-secondary" style={{fontWeight:600,fontSize:'1.09em'}} onClick={()=>{
                        setShowReviewForm(false);setReviewToEdit(null);setEditModeReviewId(null);setReviewComment('');setReviewRating(5);
                      }}>Отмена</button>
                    </div>
                  </form>
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Catalog;

