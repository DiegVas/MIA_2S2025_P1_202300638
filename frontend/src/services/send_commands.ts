interface SendCommand {
  command: string;
  timestamp: string;
  adjunt_file: string | null;
}

export async function sendCommand(
  command: SendCommand
): Promise<{ type: string; text: string; timestamp: Date } | undefined> {
  const response = await fetch("http://localhost:8080/command", {
    method: "POST",
    body: JSON.stringify(command),
    headers: {
      "Content-Type": "application/json",
    },
  });

  console.log("Sending command:", command);
  console.log("Response status:", response.status);

  if (response.ok) {
    const data = await response.json();
    console.log("Command sent successfully:", data);
    return {
      type: data.type,
      text: data.message,
      timestamp: new Date(),
    };
  } else {
    console.error("Error sending command:", response.statusText);
    return {
      type: "system-error",
      text: "Error al enviar el comando",
      timestamp: new Date(),
    };
  }
}
