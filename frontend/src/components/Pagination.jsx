import React from 'react';
import './Pagination.css';

const Pagination = ({ currentPage, totalPages, onPageChange, limit, total }) => {
  if (totalPages <= 1) return null;

  const pages = [];
  const maxVisible = 7; // Максимальное количество видимых страниц

  // Вычисляем диапазон страниц для отображения
  let startPage = Math.max(1, currentPage - Math.floor(maxVisible / 2));
  let endPage = Math.min(totalPages, startPage + maxVisible - 1);

  // Корректируем начало, если достигли конца
  if (endPage - startPage < maxVisible - 1) {
    startPage = Math.max(1, endPage - maxVisible + 1);
  }

  // Добавляем первую страницу и многоточие
  if (startPage > 1) {
    pages.push(1);
    if (startPage > 2) {
      pages.push('...');
    }
  }

  // Добавляем видимые страницы
  for (let i = startPage; i <= endPage; i++) {
    pages.push(i);
  }

  // Добавляем многоточие и последнюю страницу
  if (endPage < totalPages) {
    if (endPage < totalPages - 1) {
      pages.push('...');
    }
    pages.push(totalPages);
  }

  return (
    <div className="pagination">
      <div className="pagination-info">
        Показано {((currentPage - 1) * limit) + 1}-{Math.min(currentPage * limit, total)} из {total}
      </div>
      <div className="pagination-controls">
        <button
          className="pagination-btn"
          onClick={() => onPageChange(1)}
          disabled={currentPage === 1}
          title="Первая страница"
        >
          ⏮️
        </button>
        <button
          className="pagination-btn"
          onClick={() => onPageChange(currentPage - 1)}
          disabled={currentPage === 1}
          title="Предыдущая страница"
        >
          ◀️
        </button>

        {pages.map((page, index) => {
          if (page === '...') {
            return <span key={`ellipsis-${index}`} className="pagination-ellipsis">...</span>;
          }
          return (
            <button
              key={page}
              className={`pagination-btn ${currentPage === page ? 'active' : ''}`}
              onClick={() => onPageChange(page)}
            >
              {page}
            </button>
          );
        })}

        <button
          className="pagination-btn"
          onClick={() => onPageChange(currentPage + 1)}
          disabled={currentPage === totalPages}
          title="Следующая страница"
        >
          ▶️
        </button>
        <button
          className="pagination-btn"
          onClick={() => onPageChange(totalPages)}
          disabled={currentPage === totalPages}
          title="Последняя страница"
        >
          ⏭️
        </button>
      </div>
    </div>
  );
};

export default Pagination;














