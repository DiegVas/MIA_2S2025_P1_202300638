package utils

import "net/http"

// EnableCORS agrega los headers CORS necesarios para permitir peticiones cross-origin
func EnableCORS(w http.ResponseWriter, r *http.Request) {
    // Permitir el origen del frontend
    w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
    
    // Permitir métodos HTTP específicos
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    
    // Permitir headers específicos
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
    
    // Permitir credentials si es necesario
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    
    // Tiempo de cache para preflight requests
    w.Header().Set("Access-Control-Max-Age", "86400")
}

// CORSMiddleware es un middleware que maneja CORS para todas las rutas
func CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Aplicar headers CORS
        EnableCORS(w, r)
        
        // Si es una petición OPTIONS (preflight), responder inmediatamente
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        // Continuar con el handler siguiente
        next(w, r)
    }
}
