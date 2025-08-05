<script setup>
import ChatInput from "../common/ChatInput.vue";
import { ref, nextTick } from "vue";
import "../../styles/chat-animations.css";
import { sendCommand } from "../../services/send_commands.ts";

const messages = ref([]);
const messagesContainer = ref(null); // Referencia al contenedor de mensajes

const scrollToBottom = async () => {
  await nextTick(); // Esperar a que el DOM se actualice
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight;
  }
};

const handleSendMessage = async (message) => {
  console.log("Mensaje enviado:", message);
  messages.value.push({
    id: crypto.randomUUID(),
    text: message,
    timestamp: new Date(),
    type: "user",
  });

  await scrollToBottom();

  // Crear el objeto comando con el formato correcto
  const commandObj = {
    command: message,
    timestamp: new Date().toISOString(),
    adjunt_file: null,
  };

  const system_response = await sendCommand(commandObj);
  console.log("Respuesta del sistema:", system_response);

  messages.value.push({
    id: crypto.randomUUID(),
    ...system_response,
  });

  await scrollToBottom();
};

const handleInputChange = (value) => {};

const getMessageIcon = (type) => {
  switch (type) {
    case "user":
      return "👤";
    case "system":
      return "🤖";
    case "system-error":
      return "❌";
    default:
      return "💬";
  }
};

const getMessageClass = (type) => {
  return `message-item message-${type}`;
};
</script>

<template>
  <div class="body text-white p-5">
    <div class="content-wrapper w-[70%] mx-auto flex flex-col h-full">
      <nav
        v-if="messages.length != 0"
        class="messages-clean bg-green-500 hover:bg-green-600 px-4 py-2 rounded-t-lg cursor-pointer transition-all duration-300"
        @click="messages = []"
      >
        <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14m-7 7V5" />
        </svg>
        <span class="text-white font-medium">Nuevo chat</span>
      </nav>
      <div class="messages-area flex-1 mb-8 min-h-[300px]">
        <div
          v-if="messages.length === 0"
          class="welcome-message flex flex-col items-center justify-center h-full text-center"
        >
          <h2 class="text-2xl font-bold mb-4 text-gray-200">¡Bienvenido!</h2>
          <p class="text-gray-400">Escribe un mensaje para comenzar...</p>
        </div>

        <div
          v-else
          ref="messagesContainer"
          class="messages-list flex flex-col gap-4 overflow-y-auto max-h-full p-4 bg-gray-800/20 rounded-lg border border-gray-700"
        >
          <div
            v-for="message in messages"
            :key="message.id"
            :class="getMessageClass(message.type)"
            class="p-4 rounded-lg transition-all duration-300 hover:translate-x-1"
          >
            <div class="message-header flex justify-between items-center mb-2">
              <div class="message-type-indicator flex items-center gap-2">
                <span class="message-icon text-lg">{{ getMessageIcon(message.type) }}</span>
                <span class="message-author text-sm font-semibold uppercase tracking-wide">{{
                  message.type === "user"
                    ? "Usuario"
                    : message.type === "system"
                    ? "Sistema"
                    : message.type === "system-error"
                    ? "Error del sistema"
                    : "Desconocido"
                }}</span>
              </div>
              <div class="message-time text-xs text-gray-400">
                {{ message.timestamp.toLocaleTimeString() }}
              </div>
            </div>
            <div class="message-content text-gray-200 leading-relaxed">
              {{ message.text }}
            </div>
          </div>
        </div>
      </div>

      <!-- Input de chat -->
      <div class="sticky bottom-0 pt-4 border-t border-gray-600">
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
/* Estilos personalizados que no se pueden hacer fácilmente con Tailwind */
.messages-clean {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.3rem;
  width: fit-content;
  margin-left: 5px;
}

.message-item {
  background: rgba(255, 255, 255, 0.05);
  border-left: 3px solid var(--color-primary);
}

.message-user {
  border-left-color: var(--color-primary);
}

.message-system {
  border-left-color: var(--color-secondary);
}

.message-system-error {
  border-left-color: var(--color-error);
}

.message-content {
  font-family: var(--font-family-primary);
  word-wrap: break-word;
}

/* Scrollbar personalizada */
.overflow-y-auto::-webkit-scrollbar {
  width: 6px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 3px;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background: rgba(195, 221, 73, 0.3);
  border-radius: 3px;
}

.overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: rgba(195, 221, 73, 0.5);
}
</style>
