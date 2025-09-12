# Explicación Detallada de la Estructura y Conexión Frontend-Backend

## 1. Arquitectura General del Sistema

El sistema está dividido en dos módulos principales que trabajan en conjunto para simular un sistema de archivos EXT2:

- **Frontend**: Aplicación Vue.js que proporciona una interfaz de usuario tipo terminal
- **Backend**: Servidor HTTP en Go que procesa comandos y gestiona el sistema de archivos

### 1.1 Comunicación entre Módulos

La comunicación entre frontend y backend se realiza mediante **API REST** usando protocolo HTTP, con intercambio de datos en formato JSON.

**Flujo de comunicación:**
```
Usuario → Frontend (Vue.js) → HTTP Request → Backend (Go) → Filesystem Operations → HTTP Response → Frontend → Usuario
```

## 2. Módulo Frontend (Vue.js + TypeScript)

### 2.1 Estructura del Frontend

```
frontend/
├── src/
│   ├── App.vue                    # Componente raíz
│   ├── main.ts                    # Punto de entrada de la aplicación
│   ├── components/
│   │   ├── layout/
│   │   │   ├── Header_Layout.vue  # Header de la aplicación
│   │   │   └── Body_Layout.vue    # Cuerpo principal con chat
│   │   └── common/
│   │       ├── ChatInput.vue      # Input para comandos
│   │       └── SmiaViewer.vue     # Visor de archivos .smia
│   ├── services/
│   │   └── send_commands.ts       # Servicio de comunicación con backend
│   ├── styles/                    # Estilos CSS
│   └── types/                     # Definiciones de tipos TypeScript
├── package.json                   # Dependencias y scripts
└── vite.config.ts                # Configuración de Vite
```

### 2.2 Componentes Clave del Frontend

#### 2.2.1 Body_Layout.vue - Interfaz Principal
```vue
<script setup>
import { sendCommand } from "../../services/send_commands.ts";

const handleSendMessage = async (message) => {
  // 1. Agregar mensaje del usuario a la interfaz
  messages.value.push({
    id: crypto.randomUUID(),
    text: message,
    timestamp: new Date(),
    type: "user",
  });

  // 2. Crear objeto de comando para enviar al backend
  const commandObj = {
    command: message,
    timestamp: new Date().toISOString(),
    adjunt_file: null,
  };

  // 3. Enviar comando al backend y recibir respuesta
  const system_response = await sendCommand(commandObj);
  
  // 4. Mostrar respuesta del sistema
  messages.value.push({
    id: crypto.randomUUID(),
    ...system_response,
  });
};
</script>
```

**Funciones principales:**
- **Gestión de mensajes**: Mantiene historial de comandos y respuestas
- **Interfaz tipo chat**: Simula una terminal con interfaz moderna
- **Manejo de archivos .smia**: Permite cargar y ejecutar scripts de comandos
- **Categorización de respuestas**: Diferencia entre mensajes de usuario, sistema y errores

#### 2.2.2 send_commands.ts - Servicio de Comunicación
```typescript
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

    const data = await response.json();
    
    if (response.ok) {
      return {
        type: data.type,
        text: data.message,
        timestamp: new Date(),
      };
    } else {
      return {
        type: data.type || "system-error",
        text: data.message || "Error al enviar el comando",
        timestamp: new Date(),
      };
    }
  } catch (error) {
    // Manejo de errores de red
  }
}
```

**Características:**
- **Comunicación HTTP**: Realiza peticiones POST al endpoint `/command`
- **Manejo de errores**: Gestiona errores de red y respuestas HTTP
- **Tipado fuerte**: Utiliza TypeScript para validación de tipos
- **Configuración CORS**: Permite comunicación cross-origin

### 2.3 Tecnologías del Frontend

- **Vue.js 3**: Framework reactivo con Composition API
- **TypeScript**: Tipado estático para mayor robustez
- **Vite**: Herramienta de desarrollo y construcción
- **Tailwind CSS**: Framework de utilidades CSS para diseño

## 3. Módulo Backend (Go)

### 3.1 Estructura del Backend

```
backend/
├── main.go                    # Punto de entrada del servidor
├── routes/                    # Manejadores de rutas HTTP
│   ├── handlePostUser.go      # Router principal de comandos
│   ├── handleMkdisk.go        # Creación de discos
│   ├── handleFdisk.go         # Particionado
│   ├── handleMount.go         # Montaje de particiones
│   ├── handleMkfs.go          # Formateo EXT2
│   ├── handleLogin.go         # Autenticación
│   ├── handleUser.go          # Gestión de usuarios
│   ├── handleGroup.go         # Gestión de grupos
│   ├── handleMkdir.go         # Creación de directorios
│   ├── handleMkfile.go        # Creación de archivos
│   ├── handleCat.go           # Lectura de archivos
│   └── handleRep.go           # Generación de reportes
├── Analyzer/                  # Parsers de comandos
├── provider/                  # Lógica de negocio
├── structs/                   # Estructuras de datos
└── utils/                     # Utilidades (CORS, respuestas)
```

### 3.2 Componentes Clave del Backend

#### 3.2.1 main.go - Servidor HTTP
```go
func main() {
    // Configurar endpoint principal
    http.HandleFunc("/command", utils.CORSMiddleware(routes.HandlePostUser))
    
    // Iniciar servidor en puerto 8080
    fmt.Println("Servidor iniciado en http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**Funciones:**
- **Servidor HTTP**: Escucha en puerto 8080
- **Middleware CORS**: Permite comunicación cross-origin
- **Punto de entrada único**: Todos los comandos pasan por `/command`

#### 3.2.2 handlePostUser.go - Router de Comandos
```go
func HandlePostUser(w http.ResponseWriter, r *http.Request) {
    // 1. Validar método HTTP
    if r.Method != http.MethodPost {
        http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
        return
    }

    // 2. Leer y parsear el cuerpo de la petición
    body, err := io.ReadAll(r.Body)
    var command structs.Command
    err = json.Unmarshal(body, &command)

    // 3. Determinar tipo de comando y rutear
    commandLower := strings.ToLower(command.Command)
    commandBase := strings.Split(commandLower, " ")[0]

    switch commandBase {
    case "mkdisk":
        handleMkdisk(w, command.Command)
    case "fdisk":
        handleFdisk(w, command.Command)
    case "mount":
        handleMount(w, command.Command)
    case "mkfs":
        handleMkfs(w, command.Command)
    case "login":
        handleLogin(w, command.Command)
    // ... más comandos
    }
}
```

**Funciones principales:**
- **Routing centralizado**: Distribuye comandos a handlers específicos
- **Validación de entrada**: Verifica formato JSON y método HTTP
- **Parseo de comandos**: Identifica el tipo de comando a ejecutar

#### 3.2.3 Handlers Específicos

**Ejemplo: handleMkdisk.go**
```go
func handleMkdisk(w http.ResponseWriter, command string) {
    // 1. Parsear parámetros del comando
    params, err := Analyzer.ParseMkdisk(command)
    if err != nil {
        response := structs.Response{
            Message: fmt.Sprintf("Error en el comando: %s", err.Error()),
            Type:    "system-error",
        }
        utils.SendJSONResponse(w, response, http.StatusBadRequest)
        return
    }

    // 2. Ejecutar lógica de negocio
    err = utils.CreateDisk(params)
    if err != nil {
        // Manejo de errores
        return
    }

    // 3. Enviar respuesta exitosa
    response := structs.Response{
        Message: fmt.Sprintf("Disco creado exitosamente: %s", params.Path),
        Data:    params,
        Type:    "system",
    }
    utils.SendJSONResponse(w, response, http.StatusCreated)
}
```

### 3.3 Arquitectura por Capas del Backend

#### 3.3.1 Capa de Rutas (Routes)
- **Responsabilidad**: Manejo de peticiones HTTP
- **Funciones**: Validación de entrada, routing, respuestas HTTP
- **Ubicación**: `/routes/`

#### 3.3.2 Capa de Análisis (Analyzer)
- **Responsabilidad**: Parseo y validación de comandos
- **Funciones**: Extraer parámetros, validar sintaxis
- **Ubicación**: `/Analyzer/`

#### 3.3.3 Capa de Proveedores (Provider)
- **Responsabilidad**: Lógica de negocio del sistema de archivos
- **Módulos principales**:
  - `mounted/`: Gestión de particiones montadas
  - `files/`: Operaciones con archivos
  - `users/`: Gestión de usuarios y grupos
  - `reports/`: Generación de reportes

#### 3.3.4 Capa de Utilidades (Utils)
- **Responsabilidad**: Funciones auxiliares
- **Funciones**: CORS, respuestas JSON, operaciones de disco

## 4. Protocolo de Comunicación

### 4.1 Formato de Petición
```json
{
  "command": "mkdisk -size=10 -path=/home/user/disk1.mia",
  "timestamp": "2025-09-12T10:30:00Z",
  "adjunt_file": null
}
```

### 4.2 Formato de Respuesta
```json
{
  "message": "Disco creado exitosamente: /home/user/disk1.mia",
  "data": {
    "Size": 10,
    "Path": "/home/user/disk1.mia",
    "Unit": "M",
    "Fit": "FF"
  },
  "type": "system"
}
```

### 4.3 Tipos de Respuesta

- **`system`**: Operaciones exitosas del sistema
- **`system-error`**: Errores en comandos o parámetros
- **`auth-success`**: Autenticación exitosa
- **`auth-error`**: Errores de autenticación
- **`mounted-partitions`**: Información de particiones montadas

## 5. Gestión de Estados y Sesiones

### 5.1 Estado del Frontend
- **Mensajes**: Historial de comandos y respuestas
- **UI State**: Estado de componentes y animaciones
- **File State**: Contenido de archivos .smia cargados

### 5.2 Estado del Backend
- **Particiones montadas**: Mapa en memoria de particiones activas
- **Sesión de usuario**: Estado de autenticación (singleton)
- **Archivos binarios**: Gestión de archivos .mia en disco

## 6. Flujo de Operaciones EXT2

### 6.1 Creación y Formateo de Disco
```
Frontend: "mkdisk -size=10 -path=/disk.mia"
    ↓
Backend: Parseo → Validación → Creación archivo binario → Inicialización MBR
    ↓
Frontend: Confirmación de creación

Frontend: "fdisk -size=5 -path=/disk.mia -name=part1"
    ↓
Backend: Lectura MBR → Cálculo espacio → Creación partición → Actualización MBR
    ↓
Frontend: Confirmación de partición

Frontend: "mount -path=/disk.mia -name=part1"
    ↓
Backend: Validación partición → Registro en memoria → Generación ID
    ↓
Frontend: ID de montaje

Frontend: "mkfs -id=382part1"
    ↓
Backend: Inicialización superbloque → Creación bitmaps → Tabla inodos → Archivo users.txt
    ↓
Frontend: Confirmación de formateo EXT2
```

### 6.2 Operaciones de Archivos
```
Frontend: "login -user=root -pass=123 -id=382part1"
    ↓
Backend: Validación credenciales → Actualización sesión → Permisos
    ↓
Frontend: Sesión iniciada

Frontend: "mkdir -path=/home -id=382part1"
    ↓
Backend: Verificación permisos → Búsqueda inodo padre → Creación inodo directorio → Actualización estructura
    ↓
Frontend: Directorio creado

Frontend: "mkfile -path=/home/test.txt -size=100"
    ↓
Backend: Creación inodo archivo → Asignación bloques → Escritura contenido
    ↓
Frontend: Archivo creado
```

## 7. Manejo de Errores y Validaciones

### 7.1 Validaciones del Frontend
- **Formato de entrada**: Verificación de comandos válidos
- **Estado de conexión**: Manejo de errores de red
- **Tipos de archivo**: Validación de archivos .smia

### 7.2 Validaciones del Backend
- **Sintaxis de comandos**: Parseo y validación de parámetros
- **Estado del sistema**: Verificación de particiones montadas y sesiones
- **Integridad de datos**: Validación de estructuras EXT2
- **Permisos**: Control de acceso a archivos y directorios

## 8. Configuración CORS

### 8.1 Configuración del Backend
```go
func EnableCORS(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
    w.Header().Set("Access-Control-Allow-Credentials", "true")
}
```

### 8.2 Middleware CORS
```go
func CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        EnableCORS(w, r)
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        next(w, r)
    }
}
```

## 9. Escalabilidad y Extensibilidad

### 9.1 Patrón de Diseño
- **Router centralizado**: Fácil adición de nuevos comandos
- **Separación de responsabilidades**: Cada capa tiene una función específica
- **Interfaces bien definidas**: Comunicación clara entre componentes

### 9.2 Extensiones Posibles
- **Nuevos comandos**: Agregar handler y parser correspondiente
- **Múltiples tipos de filesystem**: Extensión de la capa de providers
- **Autenticación avanzada**: Mejoras en el sistema de sesiones
- **Interface gráfica**: Extensión del frontend con componentes visuales

## 10. Conclusión

La arquitectura del sistema frontend-backend está diseñada para simular de manera efectiva un sistema de archivos EXT2, proporcionando:

- **Interfaz intuitiva**: Terminal moderna con Vue.js
- **Comunicación robusta**: API REST con manejo de errores
- **Separación clara**: Frontend para UI, Backend para lógica de negocio
- **Extensibilidad**: Arquitectura modular para futuras mejoras
- **Simulación completa**: Implementación fiel de estructuras EXT2

El sistema mantiene la integridad de los datos mediante validaciones en ambas capas y proporciona una experiencia de usuario fluida para la gestión de sistemas de archivos simulados.
