class GoodwillAPI {
    constructor(baseUrl = '/api/v1') {
        this.baseUrl = baseUrl;
        this.csrfToken = document.querySelector('meta[name="csrf-token"]')?.content;
    }

    async request(method, endpoint, data = null) {
        const url = `${this.baseUrl}${endpoint}`;
        const options = {
            method: method,
            headers: {
                'Content-Type': 'application/json',
            }
        };

        // Add CSRF token if available
        if (this.csrfToken) {
            options.headers['X-CSRF-Token'] = this.csrfToken;
        }

        if (data) {
            options.body = JSON.stringify(data);
        }

        try {
            const response = await fetch(url, options);

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'API request failed');
            }

            return response.json();
        } catch (error) {
            console.error('API Error:', error);
            throw error;
        }
    }

    // Search methods
    async getSearches(params = {}) {
        const query = new URLSearchParams(params);
        return this.request('GET', `/searches?${query}`);
    }

    async createSearch(search) {
        return this.request('POST', '/searches', search);
    }

    async getSearch(id) {
        return this.request('GET', `/searches/${id}`);
    }

    async updateSearch(id, search) {
        return this.request('PUT', `/searches/${id}`, search);
    }

    async deleteSearch(id) {
        return this.request('DELETE', `/searches/${id}`);
    }

    async executeSearch(id) {
        return this.request('POST', `/searches/${id}/execute`);
    }

    // Item methods
    async getItems(params = {}) {
        const query = new URLSearchParams(params);
        return this.request('GET', `/items?${query}`);
    }

    async getItem(id) {
        return this.request('GET', `/items/${id}`);
    }

    async getItemHistory(id) {
        return this.request('GET', `/items/${id}/history`);
    }

    // Notification methods
    async getNotifications(params = {}) {
        const query = new URLSearchParams(params);
        return this.request('GET', `/notifications?${query}`);
    }

    async testNotification(notification) {
        return this.request('POST', '/notifications/test', notification);
    }

    // Configuration methods
    async getConfig() {
        return this.request('GET', '/config');
    }

    async updateConfig(config) {
        return this.request('PUT', '/config', config);
    }

    // System methods
    async getSystemStatus() {
        return this.request('GET', '/system/status');
    }
}

// Initialize API client
const api = new GoodwillAPI();