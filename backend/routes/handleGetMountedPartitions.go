package routes

import (
	"encoding/json"
	"fmt"
	"main/structs"
	"net/http"
)

// HandleGetMountedPartitions devuelve todas las particiones montadas
func HandleGetMountedPartitions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Obtener el gestor de montajes
	mountManager := structs.GetMountManager()
	
	// Obtener todas las particiones montadas
	mountedPartitions := mountManager.GetMountedPartitions()
	
	// Construir un mensaje explicativo
	var detailMessage string
	if len(mountedPartitions) == 0 {
		detailMessage = "No hay particiones montadas actualmente"
	} else {
		detailMessage = "Particiones montadas:\n"
		for i, partition := range mountedPartitions {
			detailMessage += fmt.Sprintf("%d. ID: %s - Disco: %s, Partición: %s, Correlativo: %d, Letra: %c\n",
				i+1, partition.ID, partition.Path, partition.Name, partition.Correlative, partition.Letter)
		}
		
		detailMessage += "\nFormato de ID: [38 (últimos 2 dígitos del carnet 202300638)][Número correlativo][Letra del disco]"
	}
	
	// Responder con JSON de éxito
	response := structs.Response{
		Message: "Particiones montadas",
		Data: map[string]interface{}{
			"partitions": mountedPartitions,
			"details": detailMessage,
		},
		Type: "system",
	}
	
	// Configurar headers de respuesta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
