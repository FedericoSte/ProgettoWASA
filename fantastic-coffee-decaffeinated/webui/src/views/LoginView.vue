<template>
  <div class="login-container">
    <div class="login-box">
      <h1>WASAText</h1>
      <p>Inserisci il tuo username per accedere</p>
      
      <!-- Il modificatore .prevent blocca il ricaricamento della pagina al submit del form -->
      <form @submit.prevent="doLogin">
        <input 
          v-model="username" 
          type="text" 
          placeholder="Es. MarioRossi" 
          required 
          minlength="3" 
          maxlength="16"
        >
        <button type="submit" :disabled="isLoading">
          {{ isLoading ? 'Accesso in corso...' : 'Entra' }}
        </button>
      </form>
      
      <p v-if="errorMsg" class="error-msg">{{ errorMsg }}</p>
    </div>
  </div>
</template>

<script>
// Importiamo l'istanza di Axios configurata nello Step precedente
import api from '../services/api';

export default {
  name: 'LoginView',
  data() {
    return {
      username: '',
      errorMsg: '',
      isLoading: false
    };
  },
  methods: {
    async doLogin() {
      this.errorMsg = '';
      this.isLoading = true;
      
      try {
        // Chiamata POST al tuo backend Go (endpoint /session)
        const response = await api.post('/session', { name: this.username });
        
        // Il tuo backend Go restituisce un JSON con { "identifier": "il_token" }
        const token = response.data.identifier;
        
        // Salviamo il token e l'username nel LocalStorage del browser
        localStorage.setItem('wasaToken', token);
        localStorage.setItem('wasaUsername', this.username);
        
        // Usiamo il router di Vue per reindirizzare l'utente alla schermata della chat
        this.$router.push('/app');
        
      } catch (error) {
        console.error(error);
        if (error.response && error.response.data && error.response.data.error) {
           this.errorMsg = error.response.data.error;
        } else {
           this.errorMsg = "Errore di connessione al server.";
        }
      } finally {
        this.isLoading = false;
      }
    }
  }
}
</script>

<style scoped>
/* CSS base per centrare il login (verrà isolato solo per questo componente) */
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background-color: #f0f2f5;
}
.login-box {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 4px 6px rgba(0,0,0,0.1);
  text-align: center;
  width: 100%;
  max-width: 400px;
}
input {
  width: 100%;
  padding: 10px;
  margin: 10px 0;
  border: 1px solid #ccc;
  border-radius: 4px;
}
button {
  width: 100%;
  padding: 10px;
  background-color: #0084ff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}
button:disabled {
  background-color: #ccc;
}
.error-msg {
  color: red;
  margin-top: 10px;
}
</style>