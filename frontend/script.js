// frontend/script.js
const API_BASE_URL = 'http://127.0.0.1:8080';

// DOM elements
const createForm = document.getElementById('createForm');
const loadDataBtn = document.getElementById('loadDataBtn');
const healthCheckBtn = document.getElementById('healthCheckBtn');
const createResult = document.getElementById('createResult');
const dataList = document.getElementById('dataList');
const healthStatus = document.getElementById('healthStatus');

// Helper function for API calls
async function apiCall(endpoint, method = 'GET', data = null) {
    const options = {
        method: method,
        headers: {
            'Content-Type': 'application/json',
        },
    };
    
    if (data) {
        options.body = JSON.stringify(data);
    }
    
    try {
        const response = await fetch(`${API_BASE_URL}${endpoint}`, options);
        const result = await response.json();
        return { ok: response.ok, status: response.status, data: result };
    } catch (error) {
        console.error('API call failed:', error);
        return { ok: false, error: error.message };
    }
}

// Create new data
createForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const keyName = document.getElementById('keyName').value;
    const dataValue = document.getElementById('dataValue').value;
    
    let parsedData;
    try {
        parsedData = JSON.parse(dataValue);
    } catch (e) {
        showResult(createResult, 'داده باید به فرمت JSON معتبر باشد', 'error');
        return;
    }
    
    const result = await apiCall('/api/data', 'POST', {
        name: keyName,
        value: parsedData
    });
    
    if (result.ok) {
        showResult(createResult, `✅ داده با موفقیت ذخیره شد! ID: ${result.data.data.id}`, 'success');
        createForm.reset();
        loadAllData();
    } else {
        showResult(createResult, `❌ خطا در ذخیره داده: ${result.data.error || 'خطای ناشناخته'}`, 'error');
    }
});

// Load all data
async function loadAllData() {
    showLoading(dataList);
    
    const result = await apiCall('/api/data', 'GET');
    
    if (result.ok && result.data.data) {
        displayDataList(result.data.data);
    } else if (result.ok) {
        dataList.innerHTML = '<p class="info">هیچ داده‌ای یافت نشد</p>';
    } else {
        dataList.innerHTML = `<p class="error">❌ خطا در بارگیری داده‌ها: ${result.data?.error || 'خطای ناشناخته'}</p>`;
    }
}

// Display data list
function displayDataList(data) {
    if (Object.keys(data).length === 0) {
        dataList.innerHTML = '<p class="info">هیچ داده‌ای یافت نشد</p>';
        return;
    }
    
    let html = '';
    for (const [key, value] of Object.entries(data)) {
        html += `
            <div class="data-item" data-key="${key}">
                <div class="data-key">🔑 ${key}</div>
                <div class="data-value">${JSON.stringify(value, null, 2)}</div>
                <button class="delete-btn" onclick="deleteData('${key}')">🗑️ حذف</button>
            </div>
        `;
    }
    dataList.innerHTML = html;
}

// Delete data
window.deleteData = async (id) => {
    if (!confirm(`آیا از حذف داده "${id}" اطمینان دارید؟`)) {
        return;
    }
    
    const result = await apiCall(`/api/data/${id}`, 'DELETE');
    
    if (result.ok) {
        showResult(createResult, `✅ داده با موفقیت حذف شد`, 'success');
        loadAllData();
    } else {
        showResult(createResult, `❌ خطا در حذف داده: ${result.data?.error || 'خطای ناشناخته'}`, 'error');
    }
};

// Health check
async function checkHealth() {
    showLoading(healthStatus);
    
    const result = await apiCall('/health', 'GET');
    
    if (result.ok) {
        healthStatus.innerHTML = `
            <div class="success">
                <strong>✅ سرویس سالم است</strong><br>
                وضعیت: ${result.data.data.status}<br>
                زمان: ${result.data.data.time}
            </div>
        `;
    } else {
        healthStatus.innerHTML = `
            <div class="error">
                <strong>❌ سرویس در دسترس نیست</strong><br>
                خطا: ${result.error || 'عدم پاسخگویی سرور'}
            </div>
        `;
    }
}

// Helper functions
function showResult(element, message, type) {
    element.innerHTML = `<div class="${type}">${message}</div>`;
    setTimeout(() => {
        if (element.innerHTML.includes(message)) {
            element.innerHTML = '';
        }
    }, 5000);
}

function showLoading(element) {
    element.innerHTML = '<div class="loading">در حال بارگیری...</div>';
}

// Event listeners
loadDataBtn.addEventListener('click', loadAllData);
healthCheckBtn.addEventListener('click', checkHealth);

// Load data on page load
document.addEventListener('DOMContentLoaded', () => {
    loadAllData();
    checkHealth();
});