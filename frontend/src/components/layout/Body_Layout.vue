<script setup>
import ChatInput from "../common/ChatInput.vue";
import { ref } from "vue";

const messages = ref([]);

const handleSendMessage = (message) => {
  console.log("Mensaje enviado:", message);
  messages.value.push({
    id: Date.now(),
    text: message,
    timestamp: new Date(),
  });
};

const handleInputChange = (value) => {
  console.log("Input cambiado:", value);
};
</script>

<template>
  <div class="body">
    <div class="content-wrapper">
      <nav v-if="messages.length != 0" class="messages-clean">
        <svg
          class="w-6 h-6 text-gray-800"
          aria-hidden="true"
          xmlns="http://www.w3.org/2000/svg"
          width="24"
          height="24"
          fill="none"
          viewBox="0 0 24 24"
        >
          <path
            stroke="currentColor"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M5 12h14m-7 7V5"
          />
        </svg>
        <span>Nuevo chat</span>
      </nav>
      <div class="messages-area">
        <div v-if="messages.length === 0" class="welcome-message">
          <h2>¡Bienvenido!</h2>
          <p>Escribe un mensaje para comenzar...</p>
        </div>

        <div v-else class="messages-list">
          <div v-for="message in messages" :key="message.id" class="message-item">
            <div class="message-content">
              {{ message.text }}
            </div>
            <div class="message-time">
              {{ message.timestamp.toLocaleTimeString() }}
            </div>
          </div>
        </div>
      </div>

      <!-- Input de chat -->
      <div class="chat-input-section">
        <ChatInput
          placeholder="Escribe tu mensaje aquí... (Enter para enviar, Shift+Enter para nueva línea)"
          @send="handleSendMessage"
          @input="handleInputChange"
          :disabled="false"
          :show-send-button="true"
          :max-rows="4"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.body {
  background: var(--color-background);
  color: var(--color-text);
  display: flex;
  flex-direction: column;
  padding: 20px;
}

.content-wrapper {
  max-width: 80%;
  margin: 0 auto;
  width: 100%;
  display: flex;
  flex-direction: column;
  height: 100%;
  flex: 1;
}

h1 {
  font-family: var(--font-family-display);
  font-size: 3rem;
  font-weight: bold;
  color: var(--color-primary);
  text-align: center;
  margin-bottom: 2rem;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}

.messages-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  margin-bottom: 2rem;
  min-height: 300px;
}

.welcome-message {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  text-align: center;
  opacity: 0.9;
}

.welcome-message h2 {
  font-family: var(--font-family-primary);
  font-size: 1.5rem;
  margin-bottom: 1rem;
  color: var(--color-text);
}

.welcome-message p {
  font-family: var(--font-family-primary);
  font-size: 1rem;
  color: rgba(253, 253, 255, 0.6);
}

.messages-clean {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: center;
  gap: 0.3rem;
  width: 12%;
  padding: 10px 15px;
  margin-left: 5px;
  background-color: var(--color-primary);
  border-radius: 8px 8px 0px 0px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94);
  position: relative;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.messages-clean span {
  color: var(--color-background);
}

.messages-clean:hover {
  background-color: color-mix(in srgb, var(--color-primary) 85%, white 15%);
  transform: translateY(-1px) scale(1.02);
  box-shadow: 0 4px 8px rgba(195, 221, 73, 0.25);
}

.messages-clean:hover span {
  animation: pulse 0.5s ease-in-out;
}

.messages-clean:hover svg {
  animation: rotate 0.5s ease-in-out;
}

@keyframes pulse {
  0%,
  100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.05);
  }
}

@keyframes rotate {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(90deg);
  }
}

.messages-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  overflow-y: auto;
  max-height: 100%;
  padding: 1rem;
  background: rgba(255, 255, 255, 0.02);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.message-item {
  background: rgba(255, 255, 255, 0.05);
  padding: 1rem;
  border-radius: 5px;
  border-left: 3px solid var(--color-primary);
  transition: all var(--transition-fast);
}

.message-item:hover {
  background: rgba(255, 255, 255, 0.08);
  transform: translateX(4px);
}

.message-content {
  font-family: var(--font-family-primary);
  font-size: 1rem;
  line-height: 1.5;
  margin-bottom: 0.5rem;
  word-wrap: break-word;
}

.message-time {
  font-size: 0.75rem;
  color: rgba(253, 253, 255, 0.4);
  font-family: var(--font-family-primary);
  text-align: right;
}

.chat-input-section {
  position: sticky;
  bottom: 0;
  background: var(--color-background);
  padding: 1rem 0;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
}

/* Scrollbar personalizado para mensajes */
.messages-list::-webkit-scrollbar {
  width: 8px;
}

.messages-list::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 4px;
}

.messages-list::-webkit-scrollbar-thumb {
  background: rgba(195, 221, 73, 0.3);
  border-radius: 4px;
}

.messages-list::-webkit-scrollbar-thumb:hover {
  background: rgba(195, 221, 73, 0.5);
}

/* Responsive */
@media (max-width: 768px) {
  .body {
    padding: 16px;
  }

  h1 {
    font-size: 2rem;
    margin-bottom: 1rem;
  }

  .messages-list {
    max-height: 300px;
    padding: 0.75rem;
  }

  .message-item {
    padding: 0.75rem;
  }

  .chat-input-section {
    padding: 0.75rem 0;
  }
}
</style>
