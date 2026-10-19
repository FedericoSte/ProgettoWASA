import axios from 'axios';

const api = axios.create({
    // Usa dinamicamente l'IP/hostname del Raspberry Pi (o localhost se aperto in locale)
    baseURL: `http://${window.location.hostname}:3000`, 
    timeout: 5000,
});

api.interceptors.request.use((config) => {
    const token = localStorage.getItem('wasaToken');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

export default api;