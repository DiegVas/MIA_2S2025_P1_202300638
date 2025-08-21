<script setup>
import { ref, nextTick, watch } from "vue";
import { useFileLoader } from "../../composables/useFileLoader.js";

// Props
const props = defineProps({
  placeholder: {
    type: String,
    default: "Escribe tu mensaje aquí...",
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  maxRows: {
    type: Number,
    default: 6,
  },
  showSendButton: {
    type: Boolean,
    default: true,
  },
});

// Emits
const emit = defineEmits(["send", "input", "focus", "blur", "fileLoaded"]);

// Reactive data
const message = ref("");
const textareaRef = ref(null);
const isFocused = ref(false);

// Composables
const { loadSmiaFile, isLoading: fileLoading, error: fileError } = useFileLoader();

// Methods
const handleInput = () => {
  emit("input", message.value);
  adjustHeight();
};

const handleSend = () => {
  if (message.value.trim() && !props.disabled) {
    emit("send", message.value.trim());
    message.value = "";
    adjustHeight();
  }
};

const handleKeydown = (event) => {
  if ((event.ctrlKey || event.metaKey) && event.key === "Enter") {
    event.preventDefault();
    handleSend();
  } else if (event.key === "Enter" && !event.shiftKey) {
    event.preventDefault();
    handleSend();
  }
};

const handleFocus = () => {
  isFocused.value = true;
  emit("focus");
};

const handleBlur = () => {
  isFocused.value = false;
  emit("blur");
};

const adjustHeight = async () => {
  await nextTick();
  if (textareaRef.value) {
    const textarea = textareaRef.value;
    textarea.style.height = "auto";

    const lineHeight = 24;
    const padding = 24;
    const maxHeight = lineHeight * props.maxRows + padding;
    const scrollHeight = textarea.scrollHeight;

    textarea.style.height = Math.min(scrollHeight, maxHeight) + "px";
  }
};

const handleFileLoad = async () => {
  try {
    const content = await loadSmiaFile();
    emit("fileLoaded", content);
  } catch (error) {
    console.error("Error al cargar archivo:", error);
  }
};

watch(message, () => {
  adjustHeight();
});

const focus = () => {
  textareaRef.value?.focus();
};

defineExpose({
  focus,
  clear: () => {
    message.value = "";
    adjustHeight();
  },
});
</script>

<template>
  <div
    class="chat-input-container"
    :class="{
      focused: isFocused,
      disabled: disabled,
      'has-content': message.length > 0,
    }"
  >
    <div class="input-wrapper">
      <button
        v-if="showSendButton"
        class="file-button"
        :class="{
          active: !fileLoading && !props.disabled,
          loading: fileLoading,
        }"
        :disabled="fileLoading || disabled"
        @click="handleFileLoad"
        type="button"
        title="Cargar archivo .smia"
      >
        <svg
          v-if="!fileLoading"
          class="file-icon"
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
            d="M7 8v8a5 5 0 1 0 10 0V6.5a3.5 3.5 0 1 0-7 0V15a2 2 0 0 0 4 0V8"
          />
        </svg>
        <svg
          v-else
          class="loading-icon"
          aria-hidden="true"
          xmlns="http://www.w3.org/2000/svg"
          width="24"
          height="24"
          fill="none"
          viewBox="0 0 24 24"
        >
          <path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6l4 2" />
          <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2" fill="none" />
        </svg>
      </button>
      <textarea
        ref="textareaRef"
        v-model="message"
        :placeholder="placeholder"
        :disabled="disabled"
        class="chat-textarea"
        rows="1"
        @input="handleInput"
        @keydown="handleKeydown"
        @focus="handleFocus"
        @blur="handleBlur"
      />

      <button
        v-if="showSendButton"
        class="send-button"
        :class="{ active: message.trim().length > 0 && !disabled }"
        :disabled="!message.trim() || disabled"
        @click="handleSend"
        type="button"
      >
        <svg class="send-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path
            d="M7 11L12 6L17 11M12 18V7"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
      </button>
    </div>

    <div v-if="message.length > 0" class="char-counter">{{ message.length }} caracteres</div>
  </div>
</template>

<style scoped>
.chat-input-container {
  position: relative;
  width: 100%;
  max-width: 768px;
  margin: 0 auto;
}

.input-wrapper {
  position: relative;
  background: var(--color-background);
  border: 2px solid rgba(255, 255, 255, 0.1);
  border-radius: 24px;
  padding: 4px;
  transition: all var(--transition-normal);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.focused .input-wrapper {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(195, 221, 73, 0.1);
}

.disabled .input-wrapper {
  opacity: 0.6;
  cursor: not-allowed;
}

.chat-textarea {
  width: 100%;
  background: transparent;
  border: none;
  outline: none;
  resize: none;
  font-family: var(--font-family-primary);
  font-size: 16px;
  line-height: 24px;
  color: var(--color-text);
  padding: 12px 60px 12px 60px;
  min-height: 48px;
  max-height: 200px;
  overflow-y: auto;
  border-radius: 20px;
  transition: all var(--transition-fast);
}

.chat-textarea::placeholder {
  color: rgba(253, 253, 255, 0.5);
  font-style: italic;
}

.chat-textarea:disabled {
  cursor: not-allowed;
}

/* Scrollbar personalizado */
.chat-textarea::-webkit-scrollbar {
  width: 6px;
}

.chat-textarea::-webkit-scrollbar-track {
  background: transparent;
}

.chat-textarea::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
}

.chat-textarea::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}

.file-button {
  position: absolute;
  left: 8px;
  bottom: 11px;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  cursor: pointer;
  transition: all var(--transition-fast);
  color: rgba(253, 253, 255, 0.5);
  background: transparent;
  border: none;
}

.file-button:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.15);
  transform: scale(1.05);
}

.file-button.active {
  color: var(--color-primary);
}

.file-button.active:hover {
  background: rgba(195, 221, 73, 0.1);
  transform: scale(1.05);
}

.file-button.loading {
  color: var(--color-primary);
  cursor: not-allowed;
}

.file-button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.file-icon,
.loading-icon {
  width: 20px;
  height: 20px;
  transition: transform var(--transition-fast);
}

.loading-icon {
  animation: spin 1s linear infinite;
}

.send-button {
  position: absolute;
  right: 8px;
  bottom: 11px;
  width: 40px;
  height: 40px;
  background: rgba(255, 255, 255, 0.1);
  border: none;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all var(--transition-fast);
  color: rgba(253, 253, 255, 0.5);
}

.send-button:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.15);
  transform: scale(1.05);
}

.send-button.active {
  background: var(--color-primary);
  color: var(--color-background);
  transform: scale(1);
}

.send-button.active:hover {
  background: #b8d142;
  transform: scale(1.05);
}

.send-button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.send-icon {
  width: 20px;
  height: 20px;
  transition: transform var(--transition-fast);
}

.send-button.active .send-icon {
  transform: translateY(-1px);
}

.char-counter {
  position: absolute;
  bottom: -24px;
  right: 8px;
  font-size: 12px;
  color: rgba(253, 253, 255, 0.4);
  font-family: var(--font-family-primary);
}

/* Animaciones */
@keyframes pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(195, 221, 73, 0.4);
  }
  70% {
    box-shadow: 0 0 0 8px rgba(195, 221, 73, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(195, 221, 73, 0);
  }
}

.send-button.active:focus {
  animation: pulse 1s infinite;
}

/* Responsive */
@media (max-width: 768px) {
  .chat-input-container {
    max-width: 100%;
    padding: 0 16px;
  }

  .chat-textarea {
    font-size: 16px; /* Evita zoom en iOS */
  }
}

/* Estados de carga */
.loading .input-wrapper {
  pointer-events: none;
}

.loading .send-button {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
