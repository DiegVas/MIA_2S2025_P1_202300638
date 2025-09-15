package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"main/Analyzer"
	"main/provider/reports"
	"main/structs"
	"main/utils"
	"net/http"
)

func handleRep(w http.ResponseWriter, command string) {

	// Parsear el comando REP
	repParams, err := Analyzer.ParseRep(command)
	if err != nil {
		fmt.Printf("ERROR parseando comando REP: %v\n", err)
		utils.SendJSONResponse(w, structs.Response{
			Message: "Error al parsear comando REP: " + err.Error(),
			Type:    "error",
		}, http.StatusBadRequest)
		return
	}

	reportPath, err := reports.GenerateReport(repParams)
	if err != nil {
		fmt.Printf("ERROR generando reporte: %v\n", err)
		utils.SendJSONResponse(w, structs.Response{
			Message: "Error al generar reporte: " + err.Error(),
			Type:    "error",
		}, http.StatusInternalServerError)
		return
	}

	fmt.Printf("Reporte generado exitosamente en: %s\n", reportPath)

	// Respuesta exitosa
	utils.SendJSONResponse(w, structs.Response{
		Message: fmt.Sprintf("Reporte %s generado exitosamente en: %s", repParams.Name, reportPath),
		Type:    "success",
		Data: map[string]interface{}{
			"report_type":  repParams.Name,
			"report_path":  reportPath,
			"partition_id": repParams.ID,
		},
	}, http.StatusOK)
}

// HandleRep maneja peticiones HTTP directas (conservado para compatibilidad)
func HandleRep(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Método HTTP: %s\n", r.Method)
	fmt.Printf("URL: %s\n", r.URL.String())
	fmt.Printf("Headers: %v\n", r.Header)
	fmt.Printf("Content-Length: %d\n", r.ContentLength)

	// Verificar que el método sea POST
	if r.Method != http.MethodPost {
		fmt.Println("ERROR: Método no permitido")
		utils.SendJSONResponse(w, structs.Response{
			Message: "Método no permitido. Use POST.",
			Type:    "error",
		}, http.StatusMethodNotAllowed)
		return
	}

	// Leer el cuerpo completo para debug
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("ERROR leyendo body: %v\n", err)
		utils.SendJSONResponse(w, structs.Response{
			Message: "Error leyendo el cuerpo de la petición: " + err.Error(),
			Type:    "error",
		}, http.StatusBadRequest)
		return
	}

	fmt.Printf("Body raw (longitud %d): %s\n", len(body), string(body))

	// Verificar si el body está vacío
	if len(body) == 0 {
		fmt.Println("ERROR: Body vacío")
		utils.SendJSONResponse(w, structs.Response{
			Message: "El cuerpo de la petición está vacío. Debe enviar un JSON con el campo 'command'",
			Type:    "error",
		}, http.StatusBadRequest)
		return
	}

	// Decodificar el JSON
	var request structs.Command
	if err := json.Unmarshal(body, &request); err != nil {
		fmt.Printf("ERROR parseando JSON: %v\n", err)
		utils.SendJSONResponse(w, structs.Response{
			Message: "Error al decodificar JSON: " + err.Error(),
			Type:    "error",
		}, http.StatusBadRequest)
		return
	}

	fmt.Printf("Comando parseado: %+v\n", request)
	fmt.Printf("Comando string: '%s'\n", request.Command)

	// Parsear el comando REP
	repParams, err := Analyzer.ParseRep(request.Command)
	if err != nil {
		fmt.Printf("ERROR parseando comando REP: %v\n", err)
		utils.SendJSONResponse(w, structs.Response{
			Message: "Error al parsear comando REP: " + err.Error(),
			Type:    "error",
		}, http.StatusBadRequest)
		return
	}

	reportPath, err := reports.GenerateReport(repParams)
	if err != nil {
		fmt.Printf("ERROR generando reporte: %v\n", err)
		utils.SendJSONResponse(w, structs.Response{
			Message: "Error al generar reporte: " + err.Error(),
			Type:    "error",
		}, http.StatusInternalServerError)
		return
	}

	// Respuesta exitosa
	utils.SendJSONResponse(w, structs.Response{
		Message: fmt.Sprintf("Reporte %s generado exitosamente en: %s", repParams.Name, reportPath),
		Type:    "success",
		Data: map[string]interface{}{
			"report_type":  repParams.Name,
			"report_path":  reportPath,
			"partition_id": repParams.ID,
		},
	}, http.StatusOK)
}
