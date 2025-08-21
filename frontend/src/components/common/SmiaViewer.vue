<script setup>
import { computed } from "vue";

const props = defineProps({
  content: {
    type: String,
    required: true,
  },
  filename: {
    type: String,
    default: "archivo.smia",
  },
});

const emit = defineEmits(["executeCommand", "executeAll"]);

// Procesar el contenido en líneas
const lines = computed(() => {
  if (!props.content) return [];

  return props.content.split("\n").map((line, index) => {
    const trimmed = line.trim();
    return {
      number: index + 1,
      content: line,
      trimmed,
      type: trimmed === "" ? "blank" : trimmed.startsWith("#") ? "comment" : "command",
    };
  });
});

// Extraer solo comandos válidos
const commands = computed(() => {
  return lines.value.filter((line) => line.type === "command").map((line) => line.trimmed);
});

const handleExecuteCommand = (command) => {
  emit("executeCommand", command);
};

const handleExecuteAll = () => {
  emit("executeAll", commands.value);
};
</script>

<template>
  <div class="smia-viewer">
    <div class="smia-header">
      <div class="file-info">
        <svg class="file-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path
            d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
        <span class="filename">{{ filename }}</span>
        <span class="file-stats">{{ lines.length }} líneas, {{ commands.length }} comandos</span>
      </div>
      <div class="actions">
        <button
          v-if="commands.length > 0"
          @click="handleExecuteAll"
          class="execute-all-btn"
          title="Ejecutar todos los comandos"
        >
          <svg class="icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path
              d="M8 5v14l11-7z"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
          Ejecutar Todo
        </button>
      </div>
    </div>

    <div class="smia-content">
      <div class="line-numbers">
        <div
          v-for="line in lines"
          :key="line.number"
          class="line-number"
          :class="{ 'command-line': line.type === 'command' }"
        >
          {{ line.number }}
        </div>
      </div>

      <div class="code-content">
        <div
          v-for="line in lines"
          :key="line.number"
          class="code-line"
          :class="{
            comment: line.type === 'comment',
            command: line.type === 'command',
            blank: line.type === 'blank',
          }"
        >
          <span class="line-text">{{ line.content || " " }}</span>
          <button
            v-if="line.type === 'command'"
            @click="handleExecuteCommand(line.trimmed)"
            class="execute-btn"
            title="Ejecutar este comando"
          >
            <svg class="icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path
                d="M8 5v14l11-7z"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.smia-viewer {
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  overflow: hidden;
  font-family: "Fira Code", "Consolas", monospace;
  font-size: 14px;
}

.smia-header {
  background: rgba(255, 255, 255, 0.05);
  padding: 12px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.file-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.file-icon {
  width: 20px;
  height: 20px;
  color: var(--color-primary);
}

.filename {
  font-weight: 600;
  color: var(--color-primary);
}

.file-stats {
  color: rgba(255, 255, 255, 0.6);
  font-size: 12px;
}

.actions {
  display: flex;
  gap: 8px;
}

.execute-all-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  background: var(--color-primary);
  color: var(--color-background);
  border: none;
  border-radius: 6px;
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.execute-all-btn:hover {
  background: #b8d142;
  transform: translateY(-1px);
}

.smia-content {
  display: flex;
  max-height: 400px;
  overflow: auto;
}

.line-numbers {
  background: rgba(0, 0, 0, 0.2);
  padding: 16px 8px;
  border-right: 1px solid rgba(255, 255, 255, 0.1);
  user-select: none;
}

.line-number {
  height: 20px;
  line-height: 20px;
  text-align: right;
  color: rgba(255, 255, 255, 0.4);
  font-size: 12px;
  min-width: 30px;
}

.line-number.command-line {
  color: var(--color-primary);
  font-weight: 600;
}

.code-content {
  flex: 1;
  padding: 16px 0;
}

.code-line {
  height: 20px;
  line-height: 20px;
  display: flex;
  align-items: center;
  padding: 0 16px;
  position: relative;
  transition: background-color var(--transition-fast);
}

.code-line:hover {
  background: rgba(255, 255, 255, 0.05);
}

.line-text {
  flex: 1;
  white-space: pre;
}

.comment .line-text {
  color: #6a9955;
  font-style: italic;
}

.command .line-text {
  color: #dcdcaa;
}

.blank .line-text {
  height: 20px;
}

.execute-btn {
  opacity: 0;
  background: var(--color-primary);
  border: none;
  border-radius: 4px;
  width: 24px;
  height: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all var(--transition-fast);
  margin-left: 8px;
}

.code-line:hover .execute-btn {
  opacity: 1;
}

.execute-btn:hover {
  background: #b8d142;
  transform: scale(1.1);
}

.execute-btn .icon {
  width: 12px;
  height: 12px;
  color: var(--color-background);
}

/* Scrollbar personalizada */
.smia-content::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.smia-content::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
}

.smia-content::-webkit-scrollbar-thumb {
  background: rgba(195, 221, 73, 0.3);
  border-radius: 3px;
}

.smia-content::-webkit-scrollbar-thumb:hover {
  background: rgba(195, 221, 73, 0.5);
}
</style>
