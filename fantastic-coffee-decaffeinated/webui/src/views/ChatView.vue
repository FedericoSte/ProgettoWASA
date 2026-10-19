<template>
  <div class="chat-layout" @click="closeAllMenus">
    
    <!-- ================= SIDEBAR SINISTRA (Lista Chat) ================= -->
    <aside class="sidebar">
      <div class="profile-header">
        <!-- L'intera area del profilo ora è cliccabile -->
        <div class="profile-info clickable-profile" @click.stop="openModal('myProfile')" title="Apri Profilo">
          <img :src="getMyAvatar()" class="my-avatar" alt="Me">
          <h2 class="clickable-title">{{ myUsername }}</h2>
        </div>
        <button @click="doLogout" class="btn-logout">Esci</button>
      </div>

      <div class="conversation-list">
        <p v-if="conversations.length === 0" class="empty-text">Nessuna chat.</p>
        <p v-else-if="filteredConversations.length === 0" class="empty-text">Nessuna corrispondenza.</p>
        
        <ul>
          <li 
            v-for="conv in filteredConversations" 
            :key="conv.id"
            :class="{ active: selectedConversation && selectedConversation.id === conv.id }"
            @click="selectConversation(conv)"
            @contextmenu.prevent="openChatMenu($event, conv)"
            class="conversation-item"
          >
            <img :src="getChatAvatar(conv)" class="list-avatar" alt="Avatar">
            <div class="list-text-container">
              <div class="conv-name-row">
                <div class="conv-name">{{ getDisplayConversationName(conv) }}</div>
                <div class="unread-badge" v-if="conv.unreadCount > 0">{{ conv.unreadCount }}</div>
              </div>
              <div class="conv-snippet" v-if="conv.lastMessage">{{ conv.lastMessage }}</div>
            </div>
          </li>
        </ul>
      </div>

      <div class="sidebar-footer">
        <input v-model="searchQuery" type="text" class="search-input" placeholder="Cerca chat...">
        <div class="fab-container">
          <div v-if="showNewMenu" class="fab-menu">
            <button @click.stop="openModal('chat')">👤 Nuova chat</button>
            <button @click.stop="openModal('group')">👥 Nuovo gruppo</button>
          </div>
          <button @click.stop="toggleNewMenu" class="fab-button">
            {{ showNewMenu ? '✕' : '＋' }}
          </button>
        </div>
      </div>
    </aside>

    <!-- ================= AREA CHAT CENTRALE ================= -->
    <main class="chat-area">
      <div v-if="!selectedConversation" class="placeholder telegram-bg">
        <div class="badge-notice">Seleziona una chat per iniziare</div>
      </div>
      
      <div v-else class="active-chat telegram-bg">
        
        <div class="chat-header clickable" @click.stop="toggleInfoSidebar">
          <div class="header-left">
            <img :src="getChatAvatar(selectedConversation)" class="header-avatar" alt="Avatar">
            <div class="header-info">
              <h3>{{ getDisplayConversationName(selectedConversation) }}</h3>
              <span class="header-status">{{ isPrivateChat(selectedConversation) ? 'ultimo accesso di recente' : currentGroupMembers.length + ' membri' }}</span>
            </div>
          </div>
        </div>
        
        <div class="messages-list" ref="messagesContainer">
          <template v-for="msg in selectedChatDetails.messages || []" :key="msg.id">
            
            <!-- MESSAGGIO SPECIALE DI SISTEMA (Stile WhatsApp) -->
            <div v-if="msg.type === 'system'" class="system-message-row">
              <div class="system-message-pill">
                {{ formatSystemMessage(msg) }}
              </div>
            </div>

            <!-- MESSAGGIO NORMALE -->
            <div 
              v-else
              :class="['message-row', isMine(msg) ? 'mine' : 'theirs']"
            >
              <img 
                v-if="!isMine(msg) && !isPrivateChat(selectedConversation)" 
                :src="getAvatarFallback(getSenderName(msg))" 
                class="msg-avatar" 
                alt="User"
              >

              <div class="message-bubble" @contextmenu.prevent="openMessageMenu($event, msg)">
                <div class="message-sender" v-if="!isMine(msg) && !isPrivateChat(selectedConversation)">
                  {{ getSenderName(msg) }}
                </div>
                
                <div class="message-body">
                  <span class="message-content">{{ msg.content }}</span>
                  <span class="message-meta">
                    <span class="message-time">{{ formatTime(msg.timestamp) }}</span>
                    <span v-if="isMine(msg)" class="message-ticks" :class="getTickClass(msg)">
                      {{ getTickIcon(msg) }}
                    </span>
                  </span>
                </div>
                
                <div class="reactions-container" v-if="msg.reactions && msg.reactions.length > 0">
                  <span 
                    v-for="reaction in msg.reactions" 
                    :key="reaction.id" 
                    @click.stop="removeReaction(msg, reaction)"
                    class="reaction-badge"
                    title="Clicca per rimuovere"
                  >
                    {{ reaction.content }}
                  </span>
                </div>
              </div>
            </div>

          </template>
        </div>
        
        <div class="message-input">
           <input 
             type="text" 
             v-model="messageInput" 
             @keyup.enter="sendMessage"
             placeholder="Scrivi un messaggio..."
           >
           <button class="btn-send-tg" @click="sendMessage" :disabled="!messageInput.trim()">➤</button>
        </div>
      </div>
    </main>

    <!-- ================= SIDEBAR DESTRA (Group Info) ================= -->
    <aside class="info-sidebar" v-if="showInfoSidebar && selectedConversation" @click.stop>
      
      <div class="info-header">
        <div class="info-header-left">
          <button class="btn-close-info" @click="showInfoSidebar = false">✕</button>
          <span class="info-title">{{ isPrivateChat(selectedConversation) ? 'User Info' : 'Group Info' }}</span>
        </div>
      </div>

      <div class="info-content-scroll">
        <div class="info-profile-section">
          <div class="avatar-edit-container">
            <img :src="getChatAvatar(selectedConversation)" class="info-large-avatar" alt="Avatar">
            <button v-if="!isPrivateChat(selectedConversation) && isGroupOwner" class="btn-edit-avatar" @click.stop="openModal('changePhoto')" title="Cambia foto">📷</button>
          </div>
          
          <div class="name-edit-container">
            <h2 class="info-name">{{ getDisplayConversationName(selectedConversation) }}</h2>
            <button v-if="!isPrivateChat(selectedConversation) && isGroupOwner" class="btn-edit-name" @click.stop="openModal('renameGroup')" title="Rinomina gruppo">✏️</button>
          </div>
          <p class="info-subtitle">{{ isPrivateChat(selectedConversation) ? 'In linea' : currentGroupMembers.length + ' membri' }}</p>
        </div>

        <div class="info-tabs" v-if="!isPrivateChat(selectedConversation)">
          <div class="tab active">Membri</div>
          <div class="tab">Media</div>
        </div>

        <div class="members-list" v-if="!isPrivateChat(selectedConversation)">
           <div class="member-item" v-for="(member, index) in currentGroupMembers" :key="index">
             <img :src="member === myUsername ? getMyAvatar() : getAvatarFallback(member)" class="member-avatar">
             <div class="member-info">
               <p class="member-name">{{ member }} <span class="me-label" v-if="member === myUsername">(tu)</span></p>
               <p class="member-status">in linea</p>
             </div>
             <span class="member-badge" v-if="member === groupOwner">Proprietario</span>
             <button 
               v-if="isGroupOwner && member !== myUsername" 
               class="btn-remove-member" 
               @click.stop="removeMemberFromGroup(member)"
               title="Rimuovi dal gruppo"
             >
               ✕
             </button>
           </div>
        </div>
      </div>

      <button 
        v-if="!isPrivateChat(selectedConversation)" 
        class="fab-add-member" 
        @click="openModal('addMember')"
        title="Aggiungi Membro"
      >
        <svg viewBox="0 0 24 24" width="28" height="28" fill="white"><path d="M15 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm-9-2V7H4v3H1v2h3v3h2v-3h3v-2H6zm9 4c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/></svg>
      </button>
    </aside>

    <!-- ================= MODALE MULTIUSO ================= -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal-content">
        <h3>{{ modalConfig.title }}</h3>
        <p v-if="modalConfig.label" class="modal-subtitle">{{ modalConfig.label }}</p>
        
        <!-- 1. Modale Profilo Personale -->
        <div v-if="modalType === 'myProfile'" class="profile-edit-section">
           <label class="input-label">Nome Utente:</label>
           <input type="text" v-model="profileEditName" class="modal-input">
           
           <label class="input-label">Foto Profilo (Sfoglia o scatta):</label>
           <input type="file" accept="image/*" @change="onFileSelected" class="modal-input file-input">
           
           <div v-if="selectedFileBase64" class="preview-img-container">
             <img :src="selectedFileBase64" class="preview-img" alt="Preview">
           </div>
        </div>

        <!-- 2. Modale Foto Gruppo -->
        <div v-else-if="modalType === 'changePhoto'" class="profile-edit-section">
           <input type="file" accept="image/*" @change="onFileSelected" class="modal-input file-input">
           <div v-if="selectedFileBase64" class="preview-img-container">
             <img :src="selectedFileBase64" class="preview-img" alt="Preview">
           </div>
        </div>

        <!-- 3. Modale Inoltro -->
        <select v-else-if="modalType === 'forward'" v-model="modalSelectValue" class="modal-input">
          <option disabled value="">Seleziona una chat...</option>
          <option v-for="c in conversations" :key="c.id" :value="c.id">{{ getDisplayConversationName(c) }}</option>
        </select>

        <!-- 4. Modali Standard (Testo) -->
        <input 
          v-else
          type="text" 
          v-model="modalInputValue" 
          @keyup.enter="confirmModal"
          autofocus
          class="modal-input"
          :placeholder="modalConfig.placeholder"
        >
        
        <div class="modal-actions">
          <button @click="closeModal" class="btn-cancel">Annulla</button>
          <button @click="confirmModal" class="btn-confirm" :disabled="isConfirmDisabled">Salva</button>
        </div>
      </div>
    </div>

    <!-- ================= MENU CONTESTUALI ================= -->
    <div v-if="chatMenu.show && chatMenu.conversation" class="custom-context-menu" :style="{ top: chatMenu.y + 'px', left: chatMenu.x + 'px' }">
      <button 
        v-if="isPrivateChat(chatMenu.conversation)" 
        @click.stop="deleteChat(chatMenu.conversation)" 
        class="btn-delete-action"
      >
        🗑️ Elimina Chat
      </button>
      <button 
        v-else 
        @click.stop="leaveToGroup(chatMenu.conversation)" 
        class="btn-delete-action"
      >
        🚪 Lascia il gruppo
      </button>
    </div>

    <div v-if="msgMenu.show" class="custom-context-menu msg-menu" :style="{ top: msgMenu.y + 'px', left: msgMenu.x + 'px' }">
      <div class="emoji-bar">
        <span @click.stop="reactToMessage('👍')">👍</span>
        <span @click.stop="reactToMessage('❤️')">❤️</span>
        <span @click.stop="reactToMessage('😂')">😂</span>
        <span @click.stop="reactToMessage('😡')">😡</span>
      </div>
      <button @click.stop="openModal('forward')" class="btn-action">➡️ Inoltra Messaggio</button>
      <button v-if="isMine(msgMenu.message)" @click.stop="deleteMessage(msgMenu.message)" class="btn-delete-action">🗑️ Elimina</button>
    </div>

  </div>
</template>

<script>
import api from '../services/api';

export default {
  name: 'ChatView',
  data() {
    return {
      myUsername: localStorage.getItem('wasaUsername') || 'Utente',
      myTokenId: localStorage.getItem('wasaToken'),
      myPhoto: localStorage.getItem('wasaPhoto') || '', // Foto salvata in locale
      conversations: [],           
      searchQuery: '',             
      selectedConversation: null,
      selectedChatDetails: {},     
      messageInput: '',            
      pollingInterval: null,       
      
      showNewMenu: false,
      showInfoSidebar: false,
      
      showModal: false,
      modalType: 'chat',           
      modalInputValue: '',
      modalSelectValue: '', 
      
      // Nuovi campi per l'immagine
      profileEditName: '',
      selectedFileBase64: '',

      chatMenu: { show: false, x: 0, y: 0, conversation: null },
      msgMenu: { show: false, x: 0, y: 0, message: null },
    };
  },
  computed: {
    filteredConversations() {
      if (!this.searchQuery.trim()) return this.conversations; 
      const lowerQuery = this.searchQuery.toLowerCase();
      return this.conversations.filter(conv => {
        return this.getDisplayConversationName(conv).toLowerCase().includes(lowerQuery);
      });
    },
    // Legge la lista reale dei membri restituita dal backend (tabella participants)
    currentGroupMembers() {
      if (
        this.selectedChatDetails &&
        Array.isArray(this.selectedChatDetails.members) &&
        this.selectedChatDetails.members.length > 0
      ) {
        return this.selectedChatDetails.members;
      }
      return [this.myUsername];
    },
    // Restituisce lo username del creatore/proprietario del gruppo
    groupOwner() {
      if (!this.selectedConversation) return '';
      return this.selectedChatDetails.owner || this.selectedConversation.owner || '';
    },
    // Verifica se l'utente loggato è il proprietario del gruppo
    isGroupOwner() {
      if (!this.selectedConversation || this.isPrivateChat(this.selectedConversation)) return false;
      return this.groupOwner === this.myUsername;
    },
    modalConfig() {
      const configs = {
        'chat': { title: 'Nuova Chat', label: "Username dell'utente:", placeholder: 'Es. marco' },
        'group': { title: 'Nuovo Gruppo', label: "Nome del gruppo:", placeholder: 'Es. Amici' },
        'addMember': { title: 'Aggiungi Membro', label: "Username da aggiungere:", placeholder: 'Es. luigi' },
        'renameGroup': { title: 'Rinomina Gruppo', label: "Nuovo nome:", placeholder: 'Nuovo nome...' },
        'changePhoto': { title: 'Modifica Foto Gruppo', label: "Scegli un'immagine dal tuo PC:" },
        'myProfile': { title: 'Il Mio Profilo', label: "" },
        'forward': { title: 'Inoltra Messaggio', label: "Seleziona la chat di destinazione:" }
      };
      return configs[this.modalType] || configs['chat'];
    },
    // Controlla se il tasto "Salva" deve essere disabilitato
    isConfirmDisabled() {
      if (this.modalType === 'forward') return !this.modalSelectValue;
      if (this.modalType === 'changePhoto') return !this.selectedFileBase64;
      if (this.modalType === 'myProfile') {
        return !this.profileEditName.trim() && !this.selectedFileBase64;
      }
      return !this.modalInputValue.trim();
    }
  },
  async mounted() {
    await this.fetchConversations();
    this.startPolling();
  },
  beforeUnmount() {
    this.stopPolling();
  },
  methods: {
    // ----------------- GESTIONE FILE (BASE64) -----------------
    onFileSelected(event) {
      const file = event.target.files[0];
      if (!file) return;

      // Legge il file locale e lo trasforma in una stringa Base64
      const reader = new FileReader();
      reader.onload = (e) => {
        this.selectedFileBase64 = e.target.result;
      };
      reader.readAsDataURL(file);
    },

    // ----------------- FORMATTAZIONE MESSAGGI DI SISTEMA -----------------
    formatSystemMessage(msg) {
      if (!msg || !msg.content) return '';
      const content = msg.content.trim();
      if (this.isMine(msg)) {
        if (content.startsWith('ha ')) {
          return 'Hai ' + content.substring(3);
        }
        return content;
      } else {
        if (content.startsWith('ha ')) {
          return `${this.getSenderName(msg)} ${content}`;
        }
        return content;
      }
    },

    // ----------------- GESTIONE SPUNTE E AVATAR -----------------
    getTickIcon(msg) {
      const status = msg.status || 'read'; 
      if (status === 'sent') return '✓';
      if (status === 'received' || status === 'read') return '✓✓';
      return '✓';
    },
    getTickClass(msg) {
      const status = msg.status || 'read';
      return status === 'read' ? 'tick-blue' : 'tick-gray';
    },
    isMine(msg) {
      return msg.sender === this.myTokenId || msg.sender === this.myUsername;
    },
    getSenderName(msg) {
      if (this.isMine(msg)) return this.myUsername;
      if (this.isPrivateChat(this.selectedConversation)) {
        return this.getDisplayConversationName(this.selectedConversation);
      }
      return msg.sender;
    },
    getDisplayConversationName(conv) {
      if (!conv || !conv.name) return 'Chat ' + (conv ? conv.id : '');
      if (conv.name.includes('|')) {
        const names = conv.name.split('|');
        return names[0] === this.myUsername ? names[1] : names[0];
      }
      return conv.name;
    },
    isPrivateChat(conv) {
      return conv && conv.name && conv.name.includes('|');
    },
    getMyAvatar() {
      return this.myPhoto || this.getAvatarFallback(this.myUsername);
    },
    getChatAvatar(conv) {
      if (conv.photo) return conv.photo;
      return this.getAvatarFallback(this.getDisplayConversationName(conv));
    },
    getAvatarFallback(name) {
      const cleanName = (name && name.length > 30) ? name.substring(0, 2) : name;
      return `https://ui-avatars.com/api/?name=${encodeURIComponent(cleanName)}&background=random&color=fff&size=128&bold=true`;
    },
    formatTime(timestamp) {
      if (!timestamp) return '';
      const date = new Date(timestamp);
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    },
    
    // ----------------- GESTIONE UI -----------------
    toggleInfoSidebar() { this.showInfoSidebar = !this.showInfoSidebar; },
    closeAllMenus() {
      this.showNewMenu = false;
      this.chatMenu.show = false;
      this.msgMenu.show = false;
    },

    // ----------------- LOGICA CHAT E MESSAGGI -----------------
    async selectConversation(conv) {
      this.selectedConversation = conv;
      this.showInfoSidebar = false; 
      this.closeAllMenus();
      await this.loadCurrentChat();
      this.scrollToBottom();
    },

    async loadCurrentChat() {
      if (!this.selectedConversation) return;
      const convId = this.selectedConversation.id;
      try {
        const res = await api.get(`/conversazioni/${convId}`);
        if (!this.selectedConversation || this.selectedConversation.id !== convId) return;

        const oldLength = this.selectedChatDetails.messages ? this.selectedChatDetails.messages.length : 0;
        this.selectedChatDetails = res.data;
        if (res.data.owner) {
          this.selectedConversation.owner = res.data.owner;
        }
        if (this.selectedChatDetails.messages && this.selectedChatDetails.messages.length > oldLength) {
          this.scrollToBottom();
        }
        await api.patch(`/conversazioni/${convId}/status`, { status: 'read' }).catch(()=>{});
        if (this.selectedConversation && this.selectedConversation.id === convId) {
          this.selectedConversation.unreadCount = 0;
        }
      } catch (e) {
        // Se il gruppo è stato cancellato dal proprietario o siamo stati rimossi (404 Not Found), chiudiamo la vista attiva
        if (e.response && e.response.status === 404) {
          if (this.selectedConversation && this.selectedConversation.id === convId) {
            this.selectedConversation = null;
            this.selectedChatDetails = {};
            this.showInfoSidebar = false;
          }
          await this.fetchConversations();
        } else {
          console.error(e);
        }
      }
    },

    async sendMessage() {
      if (!this.messageInput.trim() || !this.selectedConversation) return;
      const content = this.messageInput.trim();
      this.messageInput = ''; 
      try {
        await api.post(`/conversazioni/${this.selectedConversation.id}/messaggi`, { content: content, type: 'text' });
        await this.loadCurrentChat(); 
        this.scrollToBottom();
        await this.fetchConversations(); 
      } catch (e) { console.error(e); }
    },

    startPolling() {
      this.stopPolling();
      this.pollingInterval = setInterval(async () => {
        await this.fetchConversations();
        if (this.selectedConversation) {
          await this.loadCurrentChat();
        }
      }, 3000);
    },
    stopPolling() {
      if (this.pollingInterval) {
        clearInterval(this.pollingInterval);
        this.pollingInterval = null;
      }
    },
    scrollToBottom() {
      this.$nextTick(() => {
        const container = this.$refs.messagesContainer;
        if (container) container.scrollTop = container.scrollHeight;
      });
    },

    // ----------------- AZIONI -----------------
    async reactToMessage(emoji) {
      const msg = this.msgMenu.message;
      this.closeAllMenus();
      try {
        await api.post(`/conversazioni/${this.selectedConversation.id}/messaggi/${msg.id}/commenti`, { content: emoji });
        await this.loadCurrentChat();
      } catch (e) { console.error(e); }
    },
    async removeReaction(msg, reaction) {
      try {
        await api.delete(`/conversazioni/${this.selectedConversation.id}/messaggi/${msg.id}/commenti/${reaction.id}`);
        await this.loadCurrentChat();
      } catch (e) { console.error(e); }
    },
    async deleteMessage(msg) {
      this.closeAllMenus();
      if (!confirm("Vuoi eliminare questo messaggio per tutti?")) return;
      try {
        await api.delete(`/conversazioni/${this.selectedConversation.id}/messaggi/${msg.id}`);
        await this.loadCurrentChat();
      } catch (e) { console.error(e); }
    },

    openChatMenu(event, conv) {
      this.closeAllMenus();
      this.chatMenu.x = event.clientX;
      this.chatMenu.y = event.clientY;
      this.chatMenu.conversation = conv;
      this.chatMenu.show = true;
    },
    openMessageMenu(event, msg) {
      this.closeAllMenus();
      this.msgMenu.x = event.clientX;
      this.msgMenu.y = event.clientY;
      this.msgMenu.message = msg;
      this.msgMenu.show = true;
    },

    // Elimina una chat privata tra due utenti
    async deleteChat(conv) {
      this.closeAllMenus();
      if (!conv) return;
      if (confirm(`Sei sicuro di voler eliminare la chat con ${this.getDisplayConversationName(conv)}?`)) {
        try {
          await api.delete(`/conversazioni/${conv.id}/partecipanti/me`);
          if (this.selectedConversation && this.selectedConversation.id === conv.id) {
            this.selectedConversation = null;
            this.selectedChatDetails = {};
            this.showInfoSidebar = false;
          }
          await this.fetchConversations();
        } catch (e) {
          console.error(e);
          alert("Errore durante l'eliminazione della chat.");
        }
      }
    },

    // Lascia un gruppo (se chi esce è il proprietario, il gruppo viene eliminato per tutti)
    async leaveToGroup(conv) {
      this.closeAllMenus();
      if (!conv) return;

      const isOwner =
        (conv.owner && conv.owner === this.myUsername) ||
        (this.selectedConversation && this.selectedConversation.id === conv.id && this.isGroupOwner);

      const confirmMsg = isOwner
        ? `Sei il proprietario del gruppo "${this.getDisplayConversationName(conv)}". Uscendo, il gruppo verrà eliminato per tutti i partecipanti. Vuoi continuare?`
        : `Sei sicuro di voler lasciare il gruppo "${this.getDisplayConversationName(conv)}"?`;

      if (confirm(confirmMsg)) {
        try {
          await api.delete(`/conversazioni/${conv.id}/partecipanti/me`);
          if (this.selectedConversation && this.selectedConversation.id === conv.id) {
            this.selectedConversation = null;
            this.selectedChatDetails = {};
            this.showInfoSidebar = false;
          }
          await this.fetchConversations();
        } catch (e) {
          console.error(e);
          alert("Errore durante l'uscita dal gruppo.");
        }
      }
    },

    // Rimuove un membro dal gruppo (solo proprietario)
    async removeMemberFromGroup(memberName) {
      if (!this.selectedConversation || !this.isGroupOwner) return;
      if (confirm(`Vuoi rimuovere ${memberName} dal gruppo?`)) {
        try {
          await api.delete(`/conversazioni/${this.selectedConversation.id}/partecipanti/${encodeURIComponent(memberName)}`);
          await this.loadCurrentChat();
          await this.fetchConversations();
        } catch (e) {
          console.error(e);
          alert("Errore durante la rimozione dell'utente dal gruppo.");
        }
      }
    },

    // ----------------- MODALI E API -----------------
    toggleNewMenu() { this.showNewMenu = !this.showNewMenu; },
    openModal(type) {
      this.closeAllMenus();
      this.modalType = type;
      this.modalInputValue = '';
      this.modalSelectValue = '';
      this.selectedFileBase64 = ''; // Reset file
      this.profileEditName = this.myUsername; // Precompila il nome
      this.showModal = true;
    },
    closeModal() { this.showModal = false; },
    
    async confirmModal() {
      const input = this.modalInputValue.trim();
      const targetId = this.modalSelectValue;
      const mType = this.modalType;
      
      this.closeModal();

      try {
        if (mType === 'chat' && input) {
          await this.createConversationAPI(`${this.myUsername}|${input}`, input);
        } else if (mType === 'group' && input) {
          await this.createConversationAPI(input, null);
        } else if (mType === 'addMember' && input) {
          await api.post(`/conversazioni/${this.selectedConversation.id}/partecipanti`, { userId: input });
          await this.loadCurrentChat();
          await this.fetchConversations();
        } else if (mType === 'renameGroup' && input) {
          await api.patch(`/conversazioni/${this.selectedConversation.id}/name`, { name: input });
          this.selectedConversation.name = input;
          await this.loadCurrentChat();
          await this.fetchConversations();
        } else if (mType === 'changePhoto' && this.selectedFileBase64) {
          // Invio dell'immagine base64 per il gruppo
          await api.patch(`/conversazioni/${this.selectedConversation.id}/photo`, { photo: this.selectedFileBase64 });
          this.selectedConversation.photo = this.selectedFileBase64;
          await this.loadCurrentChat();
          await this.fetchConversations();
        } else if (mType === 'myProfile') {
          // Aggiornamento Nome Utente
          const newName = this.profileEditName.trim();
          if (newName && newName !== this.myUsername) {
            await api.patch(`/utenti/me/username`, { name: newName });
            localStorage.setItem('wasaUsername', newName);
            this.myUsername = newName;
          }
          // Aggiornamento Foto Utente
          if (this.selectedFileBase64) {
            await api.patch(`/utenti/me/photo`, { photo: this.selectedFileBase64 });
            localStorage.setItem('wasaPhoto', this.selectedFileBase64);
            this.myPhoto = this.selectedFileBase64;
          }
          await this.fetchConversations();
          if (this.selectedConversation) {
            await this.loadCurrentChat();
          }
        } else if (mType === 'forward' && targetId) {
          await api.post(`/conversazioni/${targetId}/forward`, { forwardedMessageId: this.msgMenu.message.id });
          alert("Messaggio inoltrato!");
        }
      } catch (e) {
        console.error(e);
        alert("Azione fallita. Potresti non avere i permessi o il nome è già in uso.");
      }
    },
    
    async createConversationAPI(name, userToAdd) {
      const response = await api.post('/conversazioni', { name: name });
      const newGroup = response.data;
      if (userToAdd) {
        await api.post(`/conversazioni/${newGroup.id}/partecipanti`, { userId: userToAdd });
      }
      await this.fetchConversations();
      this.selectConversation(newGroup);
    },

    async fetchConversations() {
      try {
        const response = await api.get('/conversazioni');
        this.conversations = response.data || [];
        
        // Se la conversazione attualmente selezionata non esiste più (es. gruppo eliminato o siamo stati rimossi), chiudila
        if (this.selectedConversation) {
          const updatedConv = this.conversations.find(c => c.id === this.selectedConversation.id);
          if (!updatedConv) {
            this.selectedConversation = null;
            this.selectedChatDetails = {};
            this.showInfoSidebar = false;
          } else {
            this.selectedConversation.name = updatedConv.name;
            this.selectedConversation.photo = updatedConv.photo;
            this.selectedConversation.owner = updatedConv.owner;
          }
        }

        for (let conv of this.conversations) {
          if (conv.unreadCount > 0) {
             api.patch(`/conversazioni/${conv.id}/status`, { status: 'received' }).catch(()=>{});
          }
        }
      } catch (error) {
        if(error.response && error.response.status === 401) this.doLogout();
      }
    },
    doLogout() {
      this.stopPolling();
      localStorage.removeItem('wasaToken');
      localStorage.removeItem('wasaUsername');
      localStorage.removeItem('wasaPhoto');
      this.$router.push('/login');
    }
  }
}
</script>

<style scoped>
/* ================= STILI BASE E LAYOUT ================= */
.chat-layout { display: flex; height: 100vh; overflow: hidden; position: relative; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; }
.sidebar { width: 350px; background-color: #ffffff; border-right: 1px solid #ddd; display: flex; flex-direction: column; z-index: 10; }
.chat-area { flex: 1; display: flex; flex-direction: column; position: relative; min-width: 300px; }

.telegram-bg {
  background-color: #aedab8; 
  background-image: url('https://web.telegram.org/a/chat-bg-pattern-light.png'); 
  background-size: 500px;
  background-blend-mode: overlay;
}

/* ================= HEADER SIDEBAR (Profilo Cliccabile) ================= */
.profile-header { padding: 12px 15px; background-color: #fff; display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid #eee;}
.clickable-profile { display: flex; align-items: center; gap: 10px; cursor: pointer; padding: 5px; border-radius: 8px; transition: background 0.2s;}
.clickable-profile:hover { background-color: #f4f4f5; }
.my-avatar { width: 38px; height: 38px; border-radius: 50%; object-fit: cover;}
.clickable-title { margin: 0; font-size: 1.1rem; color: #111; font-weight: 600;}
.btn-logout { padding: 6px 12px; background-color: transparent; border: none; color: #888; cursor: pointer; font-weight: bold; }
.btn-logout:hover { color: #ff3b30; }

/* ================= LISTA CONVERSAZIONI ================= */
.conversation-list { padding: 0; overflow-y: auto; flex: 1; background: #fff;}
.empty-text { padding: 20px; font-size: 0.95rem; color: #888; text-align: center; }
.conversation-list ul { list-style: none; padding: 0; margin: 0; }
.conversation-item { padding: 10px 15px; display: flex; align-items: center; gap: 12px; cursor: pointer; transition: background-color 0.2s; user-select: none; }
.conversation-item:hover { background-color: #f4f4f5; }
.conversation-item.active { background-color: #4c9ee3; color: white; }
.list-avatar { width: 48px; height: 48px; border-radius: 50%; object-fit: cover; }
.list-text-container { flex: 1; overflow: hidden; display: flex; flex-direction: column; justify-content: center;}

.conv-name-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 3px; }
.conv-name { font-weight: 600; font-size: 1rem; white-space: nowrap; text-overflow: ellipsis; overflow: hidden;}
.unread-badge { background-color: #25d366; color: white; font-size: 0.75rem; font-weight: bold; border-radius: 10px; padding: 2px 7px; min-width: 10px; text-align: center; line-height: 1.2;}
.conv-snippet { font-size: 0.9rem; color: #888; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.active .conv-name, .active .conv-snippet { color: white; }

/* ================= FOOTER SIDEBAR SINISTRA ================= */
.sidebar-footer { padding: 12px 15px; background-color: #fff; border-top: 1px solid #eee; display: flex; align-items: center; gap: 10px; }
.search-input { flex: 1; padding: 10px 15px; border: none; background: #f4f4f5; border-radius: 20px; outline: none; }
.fab-container { position: relative; }
.fab-button { width: 42px; height: 42px; border-radius: 50%; background-color: #4c9ee3; color: white; border: none; font-size: 22px; cursor: pointer; box-shadow: 0 2px 5px rgba(0,0,0,0.2); transition: 0.2s; }
.fab-button:hover { background-color: #3b88c3; transform: scale(1.05); }
.fab-menu { position: absolute; bottom: 55px; right: 0; background: white; border-radius: 8px; box-shadow: 0 4px 12px rgba(0,0,0,0.15); display: flex; flex-direction: column; overflow: hidden; width: 170px; border: 1px solid #ddd; }
.fab-menu button { padding: 12px 15px; background: none; border: none; text-align: left; cursor: pointer; font-size: 14px; }
.fab-menu button:hover { background-color: #f4f4f5; }

/* ================= CHAT AREA ================= */
.placeholder { display: flex; justify-content: center; align-items: center; height: 100%; }
.badge-notice { background: rgba(0,0,0,0.2); color: white; padding: 5px 12px; border-radius: 15px; font-size: 0.9rem; }
.active-chat { display: flex; flex-direction: column; height: 100%; }

.chat-header { 
  padding: 10px 20px; background-color: #fff; box-shadow: 0 1px 2px rgba(0,0,0,0.1); 
  display: flex; justify-content: space-between; align-items: center; z-index: 5;
  cursor: pointer; transition: background 0.2s;
}
.chat-header:hover { background-color: #f9f9f9; }
.header-left { display: flex; align-items: center; gap: 12px; }
.header-avatar { width: 42px; height: 42px; border-radius: 50%; object-fit: cover; }
.header-info h3 { margin: 0; font-size: 1.1rem; color: #111;}
.header-status { font-size: 0.8rem; color: #888; }

/* ================= MESSAGGI SPECIALI DI SISTEMA (Stile WhatsApp) ================= */
.system-message-row {
  display: flex;
  justify-content: center;
  width: 100%;
  margin: 4px 0;
}
.system-message-pill {
  background-color: #ffffff;
  color: #54656f;
  font-size: 0.85rem;
  padding: 6px 14px;
  border-radius: 8px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.12);
  text-align: center;
  max-width: 80%;
  user-select: none;
}

/* ================= MESSAGGI E SPUNTE ================= */
.messages-list { flex: 1; padding: 20px; overflow-y: auto; display: flex; flex-direction: column; gap: 8px; }
.message-row { display: flex; width: 100%; align-items: flex-end; margin-bottom: 5px; gap: 8px;}
.message-row.mine { justify-content: flex-end; }
.message-row.theirs { justify-content: flex-start; }
.msg-avatar { width: 35px; height: 35px; border-radius: 50%; object-fit: cover; margin-bottom: 2px;}

.message-bubble { 
  max-width: 65%; padding: 6px 10px 8px 12px; position: relative;
  box-shadow: 0 1px 1px rgba(0,0,0,0.15); user-select: none;
  display: flex; flex-direction: column;
}
.mine .message-bubble { background-color: #e3ffc2; border-radius: 12px 12px 0 12px; }
.theirs .message-bubble { background-color: #ffffff; border-radius: 12px 12px 12px 0; }

.message-sender { font-size: 0.85rem; font-weight: 600; color: #4c9ee3; margin-bottom: 2px; }
.message-body { display: flex; flex-wrap: wrap; align-items: flex-end; justify-content: space-between; gap: 8px; }
.message-content { font-size: 0.95rem; color: #111; word-wrap: break-word; line-height: 1.4; margin-right: 15px;}

.message-meta { display: flex; align-items: center; gap: 4px; float: right; margin-bottom: -3px; margin-left: 10px;}
.message-time { font-size: 0.7rem; color: #73a971; white-space: nowrap; }
.theirs .message-time { color: #a0aab2; }

.message-ticks { font-size: 0.75rem; font-weight: bold; letter-spacing: -2px; margin-right: 2px;}
.tick-gray { color: #888; }
.mine .tick-blue { color: #34b7f1; }

.reactions-container { display: flex; flex-wrap: wrap; gap: 4px; margin-top: 5px; }
.reaction-badge { background: #f0f2f5; border: 1px solid #ddd; border-radius: 12px; padding: 2px 6px; font-size: 0.8rem; cursor: pointer; }
.mine .reaction-badge { background: #d2f0b4; border-color: #badc9a; }
.reaction-badge:hover { filter: brightness(0.9); }

/* ================= INPUT MESSAGGIO ================= */
.message-input { padding: 15px 20px; background-color: #fff; display: flex; align-items: center; gap: 15px; box-shadow: 0 -1px 2px rgba(0,0,0,0.05);}
.message-input input { flex: 1; padding: 12px 15px; border: none; background: #f4f4f5; border-radius: 20px; outline: none; font-size: 1rem; }
.btn-send-tg { background-color: #4c9ee3; color: white; border: none; width: 45px; height: 45px; border-radius: 50%; cursor: pointer; font-size: 1.2rem; display: flex; justify-content: center; align-items: center; transition: 0.2s;}
.btn-send-tg:disabled { background-color: #a0cfff; cursor: not-allowed; }
.btn-send-tg:hover:not(:disabled) { background-color: #3b88c3; }

/* ================= SIDEBAR DESTRA (INFO) ================= */
.info-sidebar { width: 330px; background-color: #f4f4f5; border-left: 1px solid #ddd; display: flex; flex-direction: column; position: relative; animation: slideIn 0.3s ease-out; }
@keyframes slideIn { from { margin-right: -330px; } to { margin-right: 0; } }

.info-header { height: 62px; background-color: #fff; border-bottom: 1px solid #eee; display: flex; justify-content: space-between; align-items: center; padding: 0 15px; }
.info-header-left { display: flex; align-items: center; gap: 15px; }
.btn-close-info { background: none; border: none; font-size: 1.2rem; cursor: pointer; color: #666; }
.info-title { font-weight: 600; font-size: 1.1rem; color: #111; }

.info-content-scroll { flex: 1; overflow-y: auto; padding: 15px 0; }
.info-profile-section { background: #fff; text-align: center; padding: 20px 0; margin-bottom: 10px; }

.avatar-edit-container { position: relative; display: inline-block; margin-bottom: 15px; }
.info-large-avatar { width: 100px; height: 100px; border-radius: 50%; object-fit: cover; display: block;}
.btn-edit-avatar { position: absolute; bottom: 0; right: 0; width: 32px; height: 32px; border-radius: 50%; background-color: #4c9ee3; color: white; border: 2px solid white; display: flex; justify-content: center; align-items: center; cursor: pointer; font-size: 0.9rem; transition: 0.2s; }
.btn-edit-avatar:hover { background-color: #3b88c3; transform: scale(1.1); }

.name-edit-container { display: flex; justify-content: center; align-items: center; gap: 8px; margin-bottom: 5px; }
.info-name { margin: 0; font-size: 1.3rem; color: #111; }
.btn-edit-name { background: none; border: none; font-size: 1.1rem; color: #888; cursor: pointer; padding: 0; transition: color 0.2s; }
.btn-edit-name:hover { color: #4c9ee3; }

.info-subtitle { margin: 0; color: #888; font-size: 0.95rem; }

.info-tabs { display: flex; background: #fff; border-bottom: 1px solid #eee; margin-top: 10px;}
.tab { flex: 1; text-align: center; padding: 12px 0; color: #888; font-weight: 500; font-size: 0.95rem; cursor: pointer; }
.tab.active { color: #4c9ee3; border-bottom: 2px solid #4c9ee3; }

.members-list { background: #fff; padding: 10px 0; }
.member-item { display: flex; align-items: center; padding: 10px 20px; gap: 15px; }
.member-avatar { width: 45px; height: 45px; border-radius: 50%; }
.member-info { flex: 1; }
.member-name { margin: 0 0 3px 0; font-weight: 600; color: #111; }
.me-label { color: #888; font-weight: normal; font-size: 0.85rem;}
.member-status { margin: 0; font-size: 0.85rem; color: #4c9ee3; }
.member-badge { font-size: 0.75rem; background: #e5f1ff; color: #4c9ee3; padding: 3px 8px; border-radius: 12px; font-weight: 600; }
.btn-remove-member { background: none; border: none; color: #ccc; font-size: 1rem; cursor: pointer; padding: 4px 8px; border-radius: 50%; transition: 0.2s; }
.btn-remove-member:hover { color: #ff3b30; background-color: #fff0f0; }

.fab-add-member { position: absolute; bottom: 20px; right: 20px; width: 55px; height: 55px; border-radius: 50%; background-color: #4c9ee3; border: none; cursor: pointer; display: flex; justify-content: center; align-items: center; box-shadow: 0 4px 10px rgba(0,0,0,0.3); transition: 0.2s; }
.fab-add-member:hover { transform: scale(1.05); background-color: #3b88c3; }

/* ================= MODALE & CONTEXT MENU ================= */
.modal-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); display: flex; justify-content: center; align-items: center; z-index: 1000; }
.modal-content { background: white; padding: 25px; border-radius: 12px; width: 340px; box-shadow: 0 5px 15px rgba(0,0,0,0.3); }
.modal-content h3 { margin-top: 0; margin-bottom: 8px; color: #111; }
.modal-subtitle { font-size: 0.9rem; color: #555; margin-bottom: 15px; }

/* Stili Modale Profilo */
.profile-edit-section { display: flex; flex-direction: column; gap: 10px; margin-bottom: 20px;}
.input-label { font-size: 0.85rem; font-weight: 600; color: #555; }
.file-input { padding: 8px 0; border: none; }
.preview-img-container { display: flex; justify-content: center; margin-top: 10px;}
.preview-img { width: 100px; height: 100px; border-radius: 50%; object-fit: cover; border: 2px solid #eee;}

.modal-input { width: 100%; padding: 12px; margin-bottom: 20px; border: 1px solid #ccc; border-radius: 8px; outline: none; box-sizing: border-box; font-size: 1rem;}
.modal-actions { display: flex; justify-content: flex-end; gap: 10px; }
.btn-cancel, .btn-confirm { border: none; padding: 10px 18px; border-radius: 8px; cursor: pointer; font-weight: 600; font-size: 0.95rem;}
.btn-cancel { background: transparent; color: #0084ff; }
.btn-confirm { background: #4c9ee3; color: white; }
.btn-confirm:disabled { background: #a0cfff; cursor: not-allowed; }

.custom-context-menu { position: absolute; background-color: white; border-radius: 8px; box-shadow: 0 4px 15px rgba(0,0,0,0.2); border: 1px solid #ddd; z-index: 2000; overflow: hidden; min-width: 150px; }
.btn-action, .btn-delete-action { width: 100%; padding: 12px 15px; background: none; border: none; text-align: left; cursor: pointer; font-size: 14px; font-weight: 500; border-bottom: 1px solid #eee;}
.btn-action:hover { background-color: #f5f6f8; }
.btn-delete-action { color: #ff3b30; border-bottom: none;}
.btn-delete-action:hover { background-color: #fff0f0; }
.msg-menu { min-width: 180px; }
.emoji-bar { display: flex; justify-content: space-around; padding: 10px; border-bottom: 1px solid #eee; background: #fafafa; }
.emoji-bar span { font-size: 1.5rem; cursor: pointer; transition: transform 0.2s; }
.emoji-bar span:hover { transform: scale(1.3); }
</style>