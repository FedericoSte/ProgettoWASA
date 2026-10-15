<template>
  <div class="chat-layout">
    
    <!-- Barra laterale (Conversazioni) -->
    <aside class="sidebar">
      <div class="profile-header">
        <h2>Ciao, {{ myUsername }}</h2>
        <button @click="doLogout">Esci</button>
      </div>
      
      <div class="conversation-list">
        <h3>Le tue chat</h3>
        <p v-if="conversations.length === 0">Nessuna conversazione trovata.</p>
        
        <ul>
          <li v-for="conv in conversations" :key="conv.id">
            <!-- Qui in seguito aggiungeremo il click per aprire la singola chat -->
            {{ conv.name || 'Gruppo senza nome' }}
          </li>
        </ul>
      </div>
    </aside>

    <!-- Area principale (Messaggi) -->
    <main class="chat-area">
      <div class="placeholder">
        <p>Seleziona una conversazione per iniziare</p>
      </div>
    </main>

  </div>
</template>

<script>
import api from '../services/api';

export default {
  name: 'ChatView',
  data() {
    return {
      myUsername: localStorage.getItem('wasaUsername') || 'Utente',
      conversations: [],
    };
  },
  async mounted() {
    // Quando la pagina viene caricata, chiama subito questa funzione
    await this.fetchConversations();
  },
  methods: {
    async fetchConversations() {
      try {
        // Chiamata GET al tuo backend (grazie all'interceptor, il token viene inviato in automatico)
        const response = await api.get('/conversazioni');
        this.conversations = response.data || [];
      } catch (error) {
        console.error("Errore nel recupero delle conversazioni:", error);
        // Se il server risponde 401 (token scaduto o non valido), facciamo il logout
        if(error.response && error.response.status === 401){
            this.doLogout();
        }
      }
    },
    doLogout() {
      // Pulisce il localStorage e torna al login
      localStorage.removeItem('wasaToken');
      localStorage.removeItem('wasaUsername');
      this.$router.push('/login');
    }
  }
}
</script>

<style scoped>
.chat-layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}
.sidebar {
  width: 300px;
  background-color: #ffffff;
  border-right: 1px solid #ddd;
  display: flex;
  flex-direction: column;
}
.profile-header {
  padding: 15px;
  background-color: #f0f2f5;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.conversation-list {
  padding: 15px;
  overflow-y: auto;
}
.chat-area {
  flex: 1;
  background-color: #e5e5f7;
  display: flex;
  justify-content: center;
  align-items: center;
}
</style>