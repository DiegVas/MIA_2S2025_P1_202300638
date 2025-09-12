# Descripción de Estructuras de Datos Fundamentales del Sistema de Archivos

Este documento describe las principales estructuras de datos utilizadas en el sistema de archivos, organizadas y gestionadas dentro de los archivos binarios `.mia`. Estas estructuras proporcionan la base para la gestión de discos, particiones, sistema de archivos EXT2 e información de usuarios y grupos.

## 1. Master Boot Record (MBR) y Particiones

### 1.1 Master Boot Record (MBR)
```go
type MBR struct {
    Mbr_size       int32         // Tamaño del disco en bytes
    Mbr_timestamp  [10]byte      // Fecha de creación del disco
    Mbr_uuid       [16]byte      // Identificador único del disco
    Mbr_partitions int32         // Número de particiones en el disco
    Partitions     [4]PARTITION  // Tabla de particiones (máximo 4)
    Dsk_fit        [1]byte       // Algoritmo de ajuste del disco (BF, FF, WF)
}
```

**Función dentro del sistema:**
- Es la primera estructura de control del disco, ubicada al inicio del archivo `.mia`
- Contiene metadatos del disco completo y la tabla de particiones primarias
- Permite la identificación única del disco mediante UUID
- Define el algoritmo de ajuste para la asignación de espacios en el disco

**Organización en el archivo binario:**
- Se encuentra en los primeros bytes del archivo `.mia`
- Tamaño fijo de la estructura que incluye espacio para 4 particiones

### 1.2 Estructura de Partición (PARTITION)
```go
type PARTITION struct {
    Part_status     [1]byte   // Estado de la partición (activa/inactiva)
    Part_type       [1]byte   // Tipo: P(rimaria), E(xtendida), L(ógica)
    Part_fit        [1]byte   // Algoritmo de ajuste (BF, FF, WF)
    Part_start      int32     // Sector donde inicia la partición
    Part_size       int32     // Tamaño de la partición en bytes
    Part_name       [16]byte  // Nombre de la partición
    Part_uuid       [16]byte  // Identificador único de la partición
    Part_correlative int32    // Número correlativo de la partición
}
```

**Función dentro del sistema:**
- Define las características y ubicación de cada partición en el disco
- Controla el espacio asignado y el tipo de partición
- Facilita la identificación única de cada partición mediante UUID

### 1.3 Extended Boot Record (EBR)
```go
type EBR struct {
    Part_mount  [1]byte    // Indica si la partición está montada
    Part_fit    [1]byte    // Algoritmo de ajuste
    Part_start  int32      // Sector de inicio de la partición lógica
    Part_size   int32      // Tamaño de la partición lógica
    Part_next   int32      // Puntero al siguiente EBR (-1 si es el último)
    Part_name   [16]byte   // Nombre de la partición lógica
}
```

**Función dentro del sistema:**
- Gestiona las particiones lógicas dentro de una partición extendida
- Forma una lista enlazada de particiones lógicas
- Permite extender el número de particiones más allá del límite de 4 particiones primarias

**Organización en el archivo binario:**
- Se ubican inmediatamente después del inicio de cada partición lógica
- Forman una cadena enlazada donde cada EBR apunta al siguiente

## 2. Sistema de Archivos EXT2

### 2.1 Superbloque (Superblock)
```go
type Superblock struct {
    S_filesystem_type   int32    // Tipo de sistema de archivos (2=ext2, 3=ext3)
    S_inodes_count      int32    // Número total de inodos
    S_blocks_count      int32    // Número total de bloques
    S_free_blocks_count int32    // Número de bloques libres
    S_free_inodes_count int32    // Número de inodos libres
    S_mtime             [16]byte // Última fecha de montaje
    S_umtime            [16]byte // Última fecha de desmontaje
    S_mnt_count         int32    // Contador de montajes
    S_magic             int32    // Número mágico (0xEF53)
    S_inode_size        int32    // Tamaño de cada inodo
    S_block_size        int32    // Tamaño de cada bloque
    S_first_ino         int32    // Primer inodo libre
    S_first_blo         int32    // Primer bloque libre
    S_bm_inode_start    int32    // Inicio del bitmap de inodos
    S_bm_block_start    int32    // Inicio del bitmap de bloques
    S_inode_start       int32    // Inicio de la tabla de inodos
    S_block_start       int32    // Inicio de la tabla de bloques
}
```

**Función dentro del sistema:**
- Es la estructura de control principal del sistema de archivos EXT2
- Contiene información crítica sobre la organización y estado del sistema de archivos
- Permite validar la integridad mediante el número mágico
- Facilita la localización de todas las demás estructuras del sistema de archivos

**Organización en el archivo binario:**
- Se ubica al inicio de cada partición formateada con EXT2
- Tamaño: 88 bytes

### 2.2 Inodos (Inode)
```go
type Inode struct {
    I_uid   int32     // ID del usuario propietario
    I_gid   int32     // ID del grupo propietario
    I_size  int32     // Tamaño del archivo/directorio en bytes
    I_atime [16]byte  // Última fecha de acceso
    I_ctime [16]byte  // Fecha de creación
    I_mtime [16]byte  // Última fecha de modificación
    I_block [15]int32 // Punteros a bloques (12 directos + 3 indirectos)
    I_type  int32     // Tipo: 0=archivo, 1=directorio
    I_perm  int32     // Permisos del archivo/directorio
}
```

**Función dentro del sistema:**
- Es la estructura fundamental que representa archivos y directorios
- Contiene metadatos completos de cada elemento del sistema de archivos
- Gestiona la asignación de bloques mediante punteros directos e indirectos
- Controla permisos y propiedad de archivos y directorios

**Organización en el archivo binario:**
- Se almacenan en la tabla de inodos, cuya ubicación es definida por el superbloque
- Tamaño: 128 bytes por inodo
- El inodo 0 corresponde al directorio raíz

### 2.3 Bloques de Datos

#### 2.3.1 Bloque de Directorio (FolderBlock)
```go
type FolderBlock struct {
    B_content [4]Content // Hasta 4 entradas por bloque de directorio
}

type Content struct {
    B_name  [12]byte // Nombre del archivo/directorio (máximo 12 caracteres)
    B_inodo int32    // Número de inodo correspondiente
}
```

**Función dentro del sistema:**
- Almacena las entradas de un directorio
- Relaciona nombres de archivos/directorios con sus respectivos inodos
- Permite la navegación y búsqueda dentro del sistema de archivos

#### 2.3.2 Bloque de Archivo (FileBlock)
```go
type FileBlock struct {
    B_content [64]byte // Contenido del archivo (64 bytes por bloque)
}
```

**Función dentro del sistema:**
- Almacena el contenido real de los archivos
- Proporciona 64 bytes de espacio útil por bloque
- Se enlaza mediante los punteros en los inodos

#### 2.3.3 Bloque de Punteros (PointerBlock)
```go
type PointerBlock struct {
    B_pointers [16]int32 // 16 punteros a otros bloques
}
```

**Función dentro del sistema:**
- Implementa el direccionamiento indirecto para archivos grandes
- Permite extender el espacio de direccionamiento más allá de los 12 punteros directos del inodo
- Soporta hasta 3 niveles de indirección (simple, doble, triple)

**Organización en el archivo binario:**
- Todos los bloques tienen un tamaño fijo de 1024 bytes
- Se ubican en la región de bloques definida por el superbloque
- Los bloques se gestionan mediante bitmaps para control de espacio libre

## 3. Gestión de Usuarios y Grupos

### 3.1 Estructura de Usuario
```go
type User struct {
    Username string // Nombre de usuario
    Password string // Contraseña (puede estar encriptada)
    Group    string // Grupo al que pertenece
    UID      int32   // ID único de usuario
    GID      int32   // ID del grupo
}
```

**Función dentro del sistema:**
- Define los usuarios del sistema de archivos
- Controla el acceso y permisos mediante UID/GID
- Se almacena en el archivo `users.txt` dentro del sistema de archivos

### 3.2 Parámetros de Usuario y Grupo
```go
type UserParams struct {
    User string // Nombre de usuario
    Pass string // Contraseña
    Grp  string // Grupo
}

type GroupParams struct {
    Name string // Nombre del grupo
}
```

**Función dentro del sistema:**
- Estructuras auxiliares para la gestión de usuarios y grupos
- Facilitan las operaciones de creación y modificación de usuarios
- Permiten la validación de parámetros antes de la persistencia

## 4. Gestión de Sesiones

### 4.1 Sesión de Usuario
```go
type Session struct {
    Active bool   // Estado de la sesión
    User   string // Usuario activo
    ID     string // ID de la partición montada
    Group  string // Grupo del usuario
}
```

**Función dentro del sistema:**
- Mantiene el estado de autenticación del usuario activo
- Controla qué partición está siendo utilizada
- Implementa patrón Singleton para garantizar una sola sesión activa
- Facilita el control de permisos en operaciones del sistema de archivos

## 5. Constantes del Sistema

```go
const (
    SUPERBLOCK_SIZE   = 88    // Tamaño del superbloque
    INODE_SIZE        = 128   // Tamaño de un inodo
    BLOCK_SIZE        = 1024  // Tamaño de un bloque
    FILE_CONTENT_SIZE = 64    // Tamaño útil para contenido de archivos
    FOLDER_SIZE       = 64    // Tamaño de una carpeta
    POINTER_SIZE      = 64    // Tamaño de un bloque de punteros
    EXT2_MAGIC        = 0xEF53 // Número mágico para validación EXT2
    ROOT_INODE        = 0      // Número del inodo raíz
)
```

## 6. Organización General en el Archivo .mia

La organización típica de un archivo `.mia` es la siguiente:

1. **Sector 0**: MBR (Master Boot Record)
2. **Sectores siguientes**: Particiones según la tabla del MBR
3. **Por cada partición EXT2**:
   - Superbloque
   - Bitmap de inodos
   - Bitmap de bloques
   - Tabla de inodos
   - Tabla de bloques de datos

Esta organización permite un acceso eficiente y estructurado a toda la información del sistema de archivos, manteniendo la integridad de los datos y facilitando las operaciones de lectura, escritura y navegación dentro del sistema.

Las estructuras están diseñadas para optimizar el espacio de almacenamiento mientras proporcionan toda la funcionalidad necesaria para un sistema de archivos completo, incluyendo gestión de permisos, usuarios, grupos y operaciones avanzadas de archivos y directorios.
