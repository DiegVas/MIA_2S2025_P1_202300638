import { ref } from "vue";

export function useFileLoader() {
  const isLoading = ref(false);
  const error = ref(null);

  const readSmiaFile = async (file) => {
    return new Promise((resolve, reject) => {
      if (!file.name.endsWith(".smia")) {
        reject(new Error("El archivo debe tener extensión .smia"));
        return;
      }

      const reader = new FileReader();

      reader.onload = (event) => {
        try {
          const content = event.target.result;
          const processedContent = processSmiaContent(content);
          resolve(processedContent);
        } catch (err) {
          reject(new Error("Error al procesar el archivo: " + err.message));
        }
      };

      reader.onerror = () => {
        reject(new Error("Error al leer el archivo"));
      };

      reader.readAsText(file, "UTF-8");
    });
  };

  const processSmiaContent = (content) => {
    if (!content || typeof content !== "string") {
      throw new Error("Contenido de archivo inválido");
    }

    const lines = content.split("\n");

    const processedLines = lines.map((line) => {
      const trimmedLine = line.trim();

      if (trimmedLine === "") {
        return line;
      }

      if (trimmedLine.startsWith("#")) {
        return line;
      }

      return line;
    });

    return processedLines.join("\n");
  };

  const loadSmiaFile = async () => {
    isLoading.value = true;
    error.value = null;

    try {
      const input = document.createElement("input");
      input.type = "file";
      input.accept = ".smia";
      input.style.display = "none";

      return new Promise((resolve, reject) => {
        input.onchange = async (event) => {
          try {
            const file = event.target.files[0];
            if (!file) {
              reject(new Error("No se seleccionó ningún archivo"));
              return;
            }

            const content = await readSmiaFile(file);
            resolve(content);
          } catch (err) {
            error.value = err.message;
            reject(err);
          } finally {
            isLoading.value = false;
            // Limpiar el input temporal
            document.body.removeChild(input);
          }
        };

        input.oncancel = () => {
          isLoading.value = false;
          document.body.removeChild(input);
          reject(new Error("Carga de archivo cancelada"));
        };

        document.body.appendChild(input);
        input.click();
      });
    } catch (err) {
      isLoading.value = false;
      error.value = err.message;
      throw err;
    }
  };

  const isComment = (line) => {
    return line.trim().startsWith("#");
  };

  const isBlankLine = (line) => {
    return line.trim() === "";
  };

  const extractCommands = (content) => {
    if (!content) return [];

    return content
      .split("\n")
      .map((line) => line.trim())
      .filter((line) => line && !line.startsWith("#"));
  };

  return {
    isLoading,
    error,
    loadSmiaFile,
    readSmiaFile,
    processSmiaContent,
    isComment,
    isBlankLine,
    extractCommands,
  };
}
