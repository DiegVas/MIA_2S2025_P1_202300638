package main

import (
	"fmt"
	"log"
	"main/routes"
	"main/utils"
	"net/http"
)

func main() {
    http.HandleFunc("/command", utils.CORSMiddleware(routes.HandlePostUser))     // POST

    // Iniciar servidor
    fmt.Println("Servidor iniciado en http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}