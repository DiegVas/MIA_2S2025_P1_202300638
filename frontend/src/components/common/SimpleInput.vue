<script setup>
import { ref, computed } from "vue";

// Props
const props = defineProps({
  modelValue: {
    type: String,
    default: "",
  },
  placeholder: {
    type: String,
    default: "Escribe aquí...",
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  variant: {
    type: String,
    default: "default", // 'default', 'minimal', 'rounded'
    validator: (value) => ["default", "minimal", "rounded"].includes(value),
  },
});

// Emits
const emit = defineEmits(["update:modelValue", "enter", "focus", "blur"]);

// Reactive data
const inputRef = ref(null);
const isFocused = ref(false);

// Computed
const inputValue = computed({
  get: () => props.modelValue,
  set: (value) => emit("update:modelValue", value),
});

// Methods
const handleKeydown = (event) => {
  if (event.key === "Enter" && !event.shiftKey) {
    event.preventDefault();
    emit("enter", inputValue.value);
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

const focus = () => {
  inputRef.value?.focus();
};

// Expose methods
defineExpose({
  focus,
});
</script>

<template>
  <div
    class="simple-input-container"
    :class="[
      `variant-${variant}`,
      {
        focused: isFocused,
        disabled: disabled,
        'has-content': inputValue.length > 0,
      },
    ]"
  >
    <input
      ref="inputRef"
      v-model="inputValue"
      type="text"
      :placeholder="placeholder"
      :disabled="disabled"
      class="simple-input"
      @keydown="handleKeydown"
      @focus="handleFocus"
      @blur="handleBlur"
    />

    <!-- Slot para iconos o botones -->
    <div v-if="$slots.append" class="input-append">
      <slot name="append" />
    </div>
  </div>
</template>

<style scoped>
.simple-input-container {
  position: relative;
  width: 100%;
  display: flex;
  align-items: center;
}

/* Variante default */
.variant-default {
  background: var(--color-background);
  border: 2px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 4px;
  transition: all var(--transition-normal);
}

.variant-default.focused {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(195, 221, 73, 0.1);
}

/* Variante minimal */
.variant-minimal {
  background: transparent;
  border-bottom: 2px solid rgba(255, 255, 255, 0.1);
  border-radius: 0;
  padding: 8px 0;
  transition: all var(--transition-normal);
}

.variant-minimal.focused {
  border-bottom-color: var(--color-primary);
}

/* Variante rounded */
.variant-rounded {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 25px;
  padding: 4px 16px;
  transition: all var(--transition-normal);
}

.variant-rounded.focused {
  background: rgba(255, 255, 255, 0.08);
  border-color: var(--color-primary);
  box-shadow: 0 4px 12px rgba(195, 221, 73, 0.15);
}

.simple-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  font-family: var(--font-family-primary);
  font-size: 16px;
  line-height: 24px;
  color: var(--color-text);
  padding: 12px 16px;
  transition: all var(--transition-fast);
}

.variant-minimal .simple-input {
  padding: 8px 0;
}

.variant-rounded .simple-input {
  padding: 12px 8px;
}

.simple-input::placeholder {
  color: rgba(253, 253, 255, 0.5);
  font-style: italic;
}

.simple-input:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.input-append {
  display: flex;
  align-items: center;
  padding-right: 8px;
}

/* Efectos hover */
.simple-input-container:not(.disabled):hover {
  transform: translateY(-1px);
}

.variant-default:not(.disabled):hover {
  border-color: rgba(255, 255, 255, 0.2);
}

.variant-minimal:not(.disabled):hover {
  border-bottom-color: rgba(255, 255, 255, 0.3);
}

.variant-rounded:not(.disabled):hover {
  background: rgba(255, 255, 255, 0.08);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* Responsive */
@media (max-width: 768px) {
  .simple-input {
    font-size: 16px; /* Evita zoom en iOS */
  }
}
</style>
