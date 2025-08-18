interface SendCommand {
  command: string;
  timestamp: string;
  adjunt_file: string | null;
}

export async function sendCommand(
  command: SendCommand
): Promise<{ type: string; text: string; timestamp: Date } | undefined> {
  try {
    const response = await fetch("http://localhost:8080/command", {
      method: "POST",
      body: JSON.stringify(command),
      headers: {
        "Content-Type": "application/json",
      },
    });

    console.log("Sending command:", command);
    console.log("Response status:", response.status);

    // Leer la respuesta JSON independientemente del status
    const data = await response.json();
    console.log("Response data:", data);

    if (response.ok) {
      console.log("Command sent successfully:", data);
      return {
        type: data.type,
        text: data.message,
        timestamp: new Date(),
      };
    } else {
      // Manejar errores HTTP pero con respuesta JSON válida
      console.error("Error sending command:", response.statusText);
      return {
        type: data.type || "system-error",
        text: data.message || "Error al enviar el comando",
        timestamp: new Date(),
      };
    }
  } catch (error) {
    const error_parse = error instanceof Error ? error.message : "Unknown error";
    // Manejar errores de red o parsing
    console.error("Network or parsing error:", error);
    return {
      type: "system-error",
      text: error_parse,
      timestamp: new Date(),
    };
  }
}
