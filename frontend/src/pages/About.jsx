import React, { useEffect } from 'react';
import './About.css';

const About = () => {
  useEffect(() => {
    // Проверяем, не загружен ли уже скрипт
    let scriptExists = document.querySelector('script[src*="api-maps.yandex.ru"]');
    
    const initMapWithYmaps = () => {
      if (window.ymaps && window.ymaps.ready) {
        window.ymaps.ready(() => {
          setTimeout(initMap, 100); // Небольшая задержка для гарантии полной загрузки
        });
      } else {
        setTimeout(initMapWithYmaps, 100);
      }
    };

    if (!scriptExists) {
      // Создаем и загружаем скрипт
      const script = document.createElement('script');
      script.src = 'https://api-maps.yandex.ru/2.1/?apikey=5e967f67-2f2d-4529-8eea-d53f72dc4301&lang=ru_RU';
      script.async = true;
      
      script.onload = () => {
        console.log('Яндекс карты загружены');
        initMapWithYmaps();
      };
      
      script.onerror = () => {
        console.error('Ошибка загрузки Яндекс карт');
        const mapElement = document.getElementById('map');
        if (mapElement) {
          mapElement.innerHTML = '<p style="text-align: center; padding: 20px; color: #666;">Не удалось загрузить карту. Пожалуйста, обновите страницу.</p>';
        }
      };
      
      document.head.appendChild(script);
    } else {
      // Скрипт уже загружен, просто инициализируем карту
      initMapWithYmaps();
    }

    // Cleanup function
    return () => {
      const mapElement = document.getElementById('map');
      if (mapElement) {
        mapElement.innerHTML = '';
      }
    };
  }, []);

  const initMap = () => {
    const mapElement = document.getElementById('map');
    if (!mapElement) {
      console.error('Элемент map не найден');
      return;
    }

    // Проверяем, не инициализирована ли уже карта
    if (mapElement._yandexMap) {
      return;
    }

    if (!window.ymaps) {
      console.error('Yandex Maps API не загружена');
      return;
    }

    if (typeof window.ymaps.Map !== 'function') {
      console.error('window.ymaps.Map не является функцией');
      return;
    }

    try {
      const map = new window.ymaps.Map('map', {
        center: [55.751574, 37.573856], // Координаты центра (Москва)
        zoom: 12
      });

      // Сохраняем ссылку на карту для предотвращения повторной инициализации
      mapElement._yandexMap = map;

      const locations = [
        { coords: [55.752220, 37.615560], name: 'Магазин на Красной площади', address: 'Красная площадь, 1' },
        { coords: [55.767900, 37.636900], name: 'Магазин на Чистых прудах', address: 'Чистопрудный бульвар, 12' },
        { coords: [55.730600, 37.635800], name: 'Магазин у Третьяковской', address: 'Лаврушинский переулок, 8' },
        { coords: [55.760200, 37.618300], name: 'Магазин на Патриарших', address: 'Малая Бронная, 15' },
        { coords: [55.748700, 37.581500], name: 'Магазин на Киевской', address: 'Киевская улица, 22' }
      ];

      locations.forEach(location => {
        const placemark = new window.ymaps.Placemark(
          location.coords,
          {
            balloonContentHeader: location.name,
            balloonContentBody: `
              <div style="padding: 5px 0;">
                <p style="margin: 5px 0;"><strong>Адрес:</strong> ${location.address}</p>
                <p style="margin: 5px 0;"><strong>Телефон:</strong> +7 (495) 123-45-67</p>
                <p style="margin: 5px 0;"><strong>Режим работы:</strong> Ежедневно с 10:00 до 22:00</p>
              </div>
            `,
            balloonContentFooter: '',
            hintContent: location.name
          },
          {
            preset: 'islands#blueShoppingIcon'
          }
        );
        map.geoObjects.add(placemark);
      });

      // Добавляем элементы управления
      map.controls.add('zoomControl', {
        position: { top: 10, right: 10 }
      });
      map.controls.add('typeSelector', {
        position: { top: 200, right: 10 }
      });
      map.controls.add('fullscreenControl', {
        position: { top: 250, right: 10 }
      });

      console.log('Карта успешно инициализирована');
    } catch (error) {
      console.error('Ошибка инициализации карты:', error);
      if (mapElement) {
        mapElement.innerHTML = '<p style="text-align: center; padding: 20px; color: #666;">Ошибка загрузки карты. Пожалуйста, обновите страницу.</p>';
      }
    }
  };

  return (
    <div className="about-container">
      <div className="store-container">
        <h1 className="store-title">Добро пожаловать в ShoesStore!</h1>
        <p className="store-description">
          Ваш идеальный магазин кроссовок. У нас вы найдете широкий ассортимент обуви от ведущих брендов, 
          включая Nike, Adidas, Puma, Reebok и многие другие.
        </p>
        <p className="store-description">
          Мы предлагаем стильные и удобные кроссовки для спорта, повседневной жизни и особых случаев. 
          В ShoesStore каждый найдет идеальную пару!
        </p>

        <div className="store-benefits">
          <div className="benefit-item">
            <span className="benefit-icon">✅</span>
            <p>Оригинальная продукция от мировых брендов</p>
          </div>
          <div className="benefit-item">
            <span className="benefit-icon">💰</span>
            <p>Доступные цены и частые акции</p>
          </div>
          <div className="benefit-item">
            <span className="benefit-icon">🚚</span>
            <p>Быстрая доставка по всей России</p>
          </div>
          <div className="benefit-item">
            <span className="benefit-icon">🛡️</span>
            <p>Гарантия качества и удобный возврат</p>
          </div>
        </div>

        <div className="map-section">
          <h2 className="store-locations">Наши магазины в Москве</h2>
          <p className="map-description">
            Мы гордимся тем, что у нас есть несколько точек продаж в разных районах столицы. 
            Выберите ближайший магазин на карте!
          </p>
          <div id="map"></div>
        </div>

        <div className="contact-info">
          <h2>Контакты</h2>
          <div className="contact-grid">
            <div className="contact-item">
              <span className="contact-icon">📞</span>
              <div>
                <strong>Телефон:</strong>
                <p>+7 (495) 123-45-67</p>
              </div>
            </div>
            <div className="contact-item">
              <span className="contact-icon">✉️</span>
              <div>
                <strong>Email:</strong>
                <p>info@shoesstore.ru</p>
              </div>
            </div>
            <div className="contact-item">
              <span className="contact-icon">⏰</span>
              <div>
                <strong>Режим работы:</strong>
                <p>Ежедневно с 10:00 до 22:00</p>
              </div>
            </div>
            <div className="contact-item">
              <span className="contact-icon">📍</span>
              <div>
                <strong>Главный офис:</strong>
                <p>г. Москва, Красная площадь, 1</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default About;

