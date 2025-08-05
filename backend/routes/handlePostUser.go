package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"main/structs"
	"main/utils"
	"net/http"
)

func HandlePostUser(w http.ResponseWriter, r *http.Request) {
    
    if r.Method != http.MethodPost {
        http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
        return
    }

    // Leer el body de la petición
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Error leyendo el body", http.StatusBadRequest)
        return
    }

    // Decodificar JSON para comando
    var command structs.Command
    err = json.Unmarshal(body, &command)
    if err != nil {
        http.Error(w, "JSON inválido", http.StatusBadRequest)
        return
    }

    params, err := utils.ParseMkdisk(command.Command)
    if err != nil {
        fmt.Printf("HandlePostUser called - error: %s\n", err.Error()) 

        // !  Si hay error en el parsing, enviar mensaje de error
        response := structs.Response{
            Message: fmt.Sprintf("Error en el comando: %s", err.Error()),
            Data:    nil,
            Type:    "system-error",
        }
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(response)
        return
    }

    err = utils.CreateDisk(params)
    if err != nil {
    // Si hay error al crear el disco
    response := structs.Response{
        Message: fmt.Sprintf("Error al crear disco: %s", err.Error()),
        Data:    nil,
        Type:    "system-error",
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusInternalServerError)
    json.NewEncoder(w).Encode(response)
    return
}

    // Si el comando es válido y el disco se creó correctamente
    confirmationMessage := fmt.Sprintf("Disco creado exitosamente - Size: %d, Unit: %c, Fit: %s, Path: %s", 
    params.Size, params.Unit, params.Fit, params.Path)

    // Responder con JSON de éxito
    response := structs.Response{
    Message: confirmationMessage,
    Data:    params,
    Type:    "system",
}

    // Configurar headers de respuesta
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(response)
}

func HandleGetUsers(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
        return
    }

    // Datos de ejemplo


    response := structs.Response{
        Message: "Usuarios obtenidos exitosamente",
       
    }

    // Enviar respuesta JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}