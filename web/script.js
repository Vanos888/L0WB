class OrderSearch {
    constructor() {
        this.initializeEventListeners();
        this.loadOrderFromURL(); // Загружаем заказ из URL при загрузке страницы
    }

    initializeEventListeners() {
        const searchBtn = document.getElementById('searchBtn');
        const orderIdInput = document.getElementById('orderIdInput');

        searchBtn.addEventListener('click', () => this.searchOrder());
        orderIdInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.searchOrder();
            }
        });

        // Слушаем изменения URL (нажатие назад/вперед в браузере)
        window.addEventListener('popstate', (e) => {
            this.loadOrderFromURL();
        });
    }

    // Загружаем заказ из URL
    loadOrderFromURL() {
        const pathParts = window.location.pathname.split('/');
        if (pathParts.length >= 3 && pathParts[1] === 'order' && pathParts[2]) {
            const orderId = pathParts[2];
            document.getElementById('orderIdInput').value = orderId;
            this.searchOrder();
        }
    }

    async searchOrder() {
        const orderId = document.getElementById('orderIdInput').value.trim();
        console.log('Searching for order ID:', orderId);
        if (!orderId) {
            this.showError('Пожалуйста, введите Order ID');
            this.updateURL(''); // Очищаем URL если поле пустое
            return;
        }

        // Обновляем URL в браузере
        this.updateURL(orderId);

        this.showLoading();
        this.hideError();
        this.hideResult();

        const startTime = performance.now();

        try {
            const response = await fetch(`/order/get-order/${encodeURIComponent(orderId)}`, {
                method: 'GET',
                headers: {
                    'Accept': 'application/json'
                }
            });

            const responseTime = Math.round(performance.now() - startTime);

            if (!response.ok) {
                throw new Error(`Order not found (status: ${response.status})`);
            }

            const data = await response.json();

            if (data.success && data.data) {
                this.displayOrder(data.data, responseTime, data.cached);
            } else {
                throw new Error('Invalid response format');
            }

        } catch (error) {
            this.showError(error.message || 'Произошла ошибка при получении данных');
        } finally {
            this.hideLoading();
        }
    }

    // Обновляем URL в браузере
    updateURL(orderId) {
        if (orderId) {
            const newURL = `/order/${orderId}`;
            window.history.pushState({ orderId }, '', newURL);
        } else {
            window.history.pushState({}, '', '/order');
        }
    }

    displayOrder(order, responseTime, fromCache = false) {
        console.log('Order data received:', order);

        // Заполняем основную информацию (PascalCase from Go!)
        document.getElementById('orderId').textContent = order.ID;
        document.getElementById('trackNumber').textContent = order.TrackNumber;
        document.getElementById('entry').textContent = order.Entry;
        document.getElementById('locale').textContent = order.Locale;
        document.getElementById('customerId').textContent = order.CustumerID; // Опечатка в структуре!
        document.getElementById('deliveryService').textContent = order.DeliveryService;
        document.getElementById('shardKey').textContent = order.ShardKey;
        document.getElementById('smId').textContent = order.SmID;
        document.getElementById('dateCreated').textContent = new Date(order.DateCreated).toLocaleString();
        document.getElementById('oofShard').textContent = order.OofShard;

        // Заполняем информацию о доставке (PascalCase!)
        if (order.Delivery) {
            document.getElementById('deliveryName').textContent = order.Delivery.Name;
            document.getElementById('deliveryPhone').textContent = order.Delivery.Phone;
            document.getElementById('deliveryEmail').textContent = order.Delivery.Email;
            document.getElementById('deliveryZip').textContent = order.Delivery.Zip;
            document.getElementById('deliveryCity').textContent = order.Delivery.City;
            document.getElementById('deliveryAddress').textContent = order.Delivery.Address;
            document.getElementById('deliveryRegion').textContent = order.Delivery.Region;
        }

        // Заполняем информацию об оплате (PascalCase!)
        if (order.Payment) {
            document.getElementById('paymentTransaction').textContent = order.Payment.Transaction;
            document.getElementById('paymentRequestId').textContent = order.Payment.RequestID || 'N/A';
            document.getElementById('paymentCurrency').textContent = order.Payment.Currency;
            document.getElementById('paymentProvider').textContent = order.Payment.Provider;
            document.getElementById('paymentAmount').textContent = order.Payment.Amount;
            document.getElementById('paymentDt').textContent = order.Payment.PaymentDt;
            document.getElementById('paymentBank').textContent = order.Payment.Bank;
            document.getElementById('paymentDeliveryCost').textContent = order.Payment.DeliveryCost;
            document.getElementById('paymentGoodsTotal').textContent = order.Payment.GoodsTotal;
            document.getElementById('paymentCustomFee').textContent = order.Payment.CustomFee;
        }

        // Заполняем информацию о товарах
        this.displayItems(order.Items);

        // Показываем информацию о ответе
        document.getElementById('responseTime').textContent = responseTime;
        document.getElementById('dataSource').textContent = fromCache ? 'кеш' : 'база данных';
        document.getElementById('dataSource').style.color = fromCache ? '#28a745' : '#dc3545';

        this.showResult();
    }

    displayItems(items) {
        const itemsContainer = document.getElementById('itemsList');
        itemsContainer.innerHTML = '';

        if (!items || items.length === 0) {
            itemsContainer.innerHTML = '<p class="no-items">Нет товаров</p>';
            return;
        }

        items.forEach((item, index) => {
            const itemCard = document.createElement('div');
            itemCard.className = 'item-card';
            itemCard.innerHTML = `
            <h4>Товар ${index + 1}: ${item.Name}</h4>
            <div class="info-grid">
                <div class="info-item">
                    <label>Brand:</label>
                    <span>${item.Brand}</span>
                </div>
                <div class="info-item">
                    <label>Price:</label>
                    <span>${item.Price} RUB</span>
                </div>
                <div class="info-item">
                    <label>Sale:</label>
                    <span>${item.Sale}%</span>
                </div>
                <div class="info-item">
                    <label>Total Price:</label>
                    <span>${item.TotalPrice}</span>
                </div>
                <div class="info-item">
                    <label>Size:</label>
                    <span>${item.Size}</span>
                </div>
                <div class="info-item">
                    <label>Status:</label>
                    <span>${item.Status}</span>
                </div>
                <div class="info-item">
                    <label>Track Number:</label>
                    <span>${item.TrackNumber}</span>
                </div>
                <div class="info-item">
                    <label>Chart ID:</label>
                    <span>${item.ChartID}</span>
                </div>
                <div class="info-item">
                    <label>RID:</label>
                    <span>${item.RID}</span>
                </div>
                <div class="info-item">
                    <label>Nm ID:</label>
                    <span>${item.NmID}</span>
                </div>
            </div>
        `;
            itemsContainer.appendChild(itemCard);
        });
    }

    switchTab(clickedTab) {
        // Деактивируем все табы
        document.querySelectorAll('.tab').forEach(tab => {
            tab.classList.remove('active');
        });

        // Скрываем все содержимое табов
        document.querySelectorAll('.tab-pane').forEach(pane => {
            pane.classList.remove('active');
        });

        // Активируем выбранный таб
        clickedTab.classList.add('active');
        const tabName = clickedTab.getAttribute('data-tab');
        document.getElementById(`tab-${tabName}`).classList.add('active');
    }

    showLoading() {
        document.getElementById('loading').classList.remove('hidden');
        document.getElementById('searchBtn').disabled = true;
    }

    hideLoading() {
        document.getElementById('loading').classList.add('hidden');
        document.getElementById('searchBtn').disabled = false;
    }

    showError(message) {
        const errorElement = document.getElementById('error');
        const errorMessage = document.getElementById('errorMessage');

        errorMessage.textContent = message;
        errorElement.classList.remove('hidden');
    }

    hideError() {
        document.getElementById('error').classList.add('hidden');
    }

    showResult() {
        document.getElementById('result').classList.remove('hidden');
    }

    hideResult() {
        document.getElementById('result').classList.add('hidden');
    }
}

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    new OrderSearch();
});

function switchTab(clickedTab) {
    // Деактивируем все табы
    document.querySelectorAll('.tab').forEach(tab => {
        tab.classList.remove('active');
    });

    // Скрываем все содержимое табов
    document.querySelectorAll('.tab-pane').forEach(pane => {
        pane.classList.remove('active');
    });

    // Активируем выбранный таб
    clickedTab.classList.add('active');
    const tabName = clickedTab.getAttribute('data-tab');
    const tabPane = document.getElementById(`tab-${tabName}`);
    if (tabPane) {
        tabPane.classList.add('active');
    }
}

// Инициализация табов при загрузке страницы
document.addEventListener('DOMContentLoaded', function() {
    // Инициализируем обработчики для табов
    document.querySelectorAll('.tab').forEach(tab => {
        tab.addEventListener('click', function(e) {
            e.preventDefault();
            switchTab(this);
        });
    });

    // Активируем первую вкладку по умолчанию
    const firstTab = document.querySelector('.tab');
    if (firstTab) {
        switchTab(firstTab);
    }

    // Инициализируем поиск заказов
    new OrderSearch();
});