import axios from 'axios';

// Crea un'istanza di Axios pre-configurata
const api = axios.create({
    // Inserisci qui l'URL su cui gira il tuo backend Go (di solito localhost:3000 o 8080)
    baseURL: 'http://localhost:3000', 
    timeout: 5000,
});

// Interceptor: prima di ogni richiesta, controlla se abbiamo un token salvato e lo aggiunge
api.interceptors.request.use((config) => {
    const token = localStorage.getItem('wasaToken');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

export default api;