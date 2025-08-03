package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"main/structs"
	"net/http"
)

func HandlePostUser(w http.ResponseWriter, r *http.Request) {
    print("HandlePostUser called")
    
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

    // Procesar el comando (aquí solo lo imprimimos)
    fmt.Printf("Comando recibido: %+v\n", command)

    // Responder con JSON
    response := structs.Response{
        Message: "Comando procesado exitosamente",
        Data:    command,
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