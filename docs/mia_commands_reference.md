# MIA Project Commands Reference Guide

## Tabla de Contenidos

1. [Administración de Discos](#administración-de-discos)
2. [Administración del Sistema de Archivos](#administración-del-sistema-de-archivos)
3. [Administración de Usuarios y Grupos](#administración-de-usuarios-y-grupos)
4. [Administración de Carpetas, Archivos y Permisos](#administración-de-carpetas-archivos-y-permisos)
5. [Reportes](#reportes)

---

## Administración de Discos

### 1. MKDISK
**Propósito**: Crea un archivo binario que simula un disco duro virtual con extensión `.mia`.

**Parámetros**:
- `-size` (Obligatorio): Tamaño del disco a crear (debe ser positivo y mayor que cero)
- `-fit` (Opcional): Tipo de ajuste para particiones
  - `BF`: Best Fit (Mejor Ajuste)
  - `FF`: First Fit (Primer Ajuste) - *valor por defecto*
  - `WF`: Worst Fit (Peor Ajuste)
- `-unit` (Opcional): Unidades del parámetro size
  - `K`: Kilobytes (1024 bytes)
  - `M`: Megabytes (1024 * 1024 bytes) - *valor por defecto*
- `-path` (Obligatorio): Ruta donde se creará el archivo del disco

**Efectos en las Estructuras**:
- Crea un archivo binario inicializado con ceros
- Escribe la estructura MBR en el primer sector
- Inicializa las 4 particiones con valores por defecto
- Establece el tamaño total del disco

**Ejemplos**:
```bash
# Crea un disco de 3000 KB
mkdisk -Size=3000 -unit=K -path=/home/user/Disco1.mia

# Crea un disco de 5 MB con Best Fit
mkdisk -size=5 -unit=M -fit=BF -path="/home/mis discos/Disco3.mia"

# Crea un disco de 10 MB (unit por defecto)
mkdisk -size=10 -path="/home/mis discos/Disco5.mia"
```

### 2. RMDISK
**Propósito**: Elimina un archivo que representa un disco duro virtual.

**Parámetros**:
- `-path` (Obligatorio): Ruta del archivo de disco a eliminar

**Efectos en las Estructuras**:
- Elimina físicamente el archivo `.mia` del sistema
- Libera el espacio ocupado en el disco físico

**Ejemplo**:
```bash
# Elimina Disco4.mia
rmdisk -path="/home/mis discos/Disco4.mia"
```

### 3. FDISK
**Propósito**: Administra particiones en el disco virtual (crear, eliminar, modificar).

**Parámetros**:
- `-size` (Obligatorio para crear): Tamaño de la partición (positivo y mayor que cero)
- `-unit` (Opcional): Unidades del parámetro size
  - `B`: Bytes
  - `K`: Kilobytes - *valor por defecto*
  - `M`: Megabytes
- `-path` (Obligatorio): Ruta del disco donde se creará la partición
- `-type` (Opcional): Tipo de partición
  - `P`: Partición Primaria - *valor por defecto*
  - `E`: Partición Extendida
  - `L`: Partición Lógica
- `-fit` (Opcional): Tipo de ajuste
  - `BF`: Best Fit
  - `FF`: First Fit
  - `WF`: Worst Fit - *valor por defecto*
- `-name` (Obligatorio): Nombre de la partición

**Restricciones**:
- Máximo 4 particiones primarias + extendidas
- Solo una partición extendida por disco
- Particiones lógicas solo dentro de extendida
- Nombres únicos por disco

**Efectos en las Estructuras**:
- **Particiones Primarias**: Actualiza el MBR, reserva espacio en el disco
- **Particiones Extendidas**: Actualiza el MBR, crea el primer EBR
- **Particiones Lógicas**: Crea nuevos EBRs enlazados, actualiza la cadena de EBRs

**Ejemplos**:
```bash
# Crea partición primaria de 300KB
fdisk -Size=300 -path=/home/Disco1.mia -name=Particion1

# Crea partición extendida
fdisk -type=E -path=/home/Disco2.mia -Unit=K -name=Particion2 -size=300

# Crea partición lógica con Best Fit
fdisk -size=1 -type=L -unit=M -fit=BF -path="/mis discos/Disco3.mia" -name="Particion3"
```

### 4. MOUNT
**Propósito**: Monta una partición del disco en el sistema, asignándole un ID único.

**Parámetros**:
- `-path` (Obligatorio): Ruta del disco que contiene la partición
- `-name` (Obligatorio): Nombre de la partición a montar

**Sistema de IDs**:
Formato: `[Últimos 2 dígitos del carnet][Número de partición][Letra del disco]`
- Ejemplo con carnet 202401234: 341A, 341B, 341C, 342A, 343A

**Efectos en las Estructuras**:
- Actualiza el campo `part_status` a montado
- Asigna el correlativo de partición
- Genera y asigna el ID único
- Mantiene registro en memoria RAM (no en disco)

**Ejemplos**:
```bash
mount -path=/home/Disco2.mia -name=Part2 # ID: 341A
mount -path=/home/Disco2.mia -name=Part1 # ID: 342A
mount -path=/home/Disco1.mia -name=Part3 # ID: 341B
```

### 5. MOUNTED
**Propósito**: Muestra todas las particiones montadas en memoria.

**Parámetros**: Ninguno

**Efectos en las Estructuras**:
- Solo lectura, no modifica estructuras
- Consulta el registro en memoria RAM

**Ejemplo**:
```bash
mounted
# Salida esperada: 341A, 342A, 341B, 341C
```

---

## Administración del Sistema de Archivos

### 6. MKFS
**Propósito**: Formatea una partición montada con el sistema de archivos EXT2.

**Parámetros**:
- `-id` (Obligatorio): ID de la partición montada generado con MOUNT
- `-type` (Opcional): Tipo de formateo
  - `full`: Formateo completo - *valor por defecto*

**Efectos en las Estructuras**:
- Crea el **Superbloque** con metadatos del sistema EXT2
- Inicializa **Bitmap de Inodos** (todos en 0, usables)
- Inicializa **Bitmap de Bloques** (todos en 0, usables)
- Crea la **Tabla de Inodos** vacía
- Crea la **Tabla de Bloques** vacía
- Crea el **archivo users.txt** en la raíz con usuario root por defecto
- Establece los punteros de inicio para todas las estructuras

**Cálculo de Estructuras**:
```
tamaño_particion = sizeof(superblock) + n + 3*n + n*sizeof(inodo) + 3*n*sizeof(block)
numero_estructuras = floor(n)
```

**Ejemplos**:
```bash
# Formateo completo
mkfs -type=full -id=341A

# Formateo por defecto
mkfs -id=342A
```

### 7. CAT
**Propósito**: Muestra el contenido de uno o varios archivos concatenados.

**Parámetros**:
- `-filen` (Obligatorio): Lista de archivos a mostrar (file1, file2, file3, etc.)

**Validaciones**:
- Verifica permisos de lectura del usuario actual
- Valida existencia de archivos
- Requiere sesión activa

**Efectos en las Estructuras**:
- **Lectura**: Accede a los inodos de los archivos
- **Lectura**: Sigue los punteros directos e indirectos
- **Lectura**: Lee el contenido de los bloques de archivo
- Actualiza `i_atime` (último acceso) en los inodos

**Ejemplos**:
```bash
# Lee un archivo
cat -file1=/home/user/docs/a.txt

# Concatena múltiples archivos
cat -file1="/home/a.txt" -file2="/home/b.txt" -file3="/home/c.txt"
```

---

## Administración de Usuarios y Grupos

### 8. LOGIN
**Propósito**: Inicia sesión en el sistema con un usuario específico.

**Parámetros**:
- `-user` (Obligatorio): Nombre de usuario (case-sensitive)
- `-pass` (Obligatorio): Contraseña del usuario (case-sensitive)
- `-id` (Obligatorio): ID de la partición montada donde iniciar sesión

**Efectos en las Estructuras**:
- **Lectura**: Accede al archivo `users.txt` para validar credenciales
- Establece la sesión activa en memoria
- Define el contexto de trabajo (partición y permisos)

**Restricciones**:
- Solo una sesión activa a la vez
- Debe usar LOGOUT antes de otro LOGIN

**Ejemplo**:
```bash
login -user=root -pass=123 -id=062A
```

### 9. LOGOUT
**Propósito**: Cierra la sesión activa del usuario.

**Parámetros**: Ninguno

**Efectos en las Estructuras**:
- Limpia la sesión de memoria
- No modifica estructuras del disco

**Ejemplo**:
```bash
logout
```

### 10. MKGRP
**Propósito**: Crea un nuevo grupo de usuarios (solo usuario root).

**Parámetros**:
- `-name` (Obligatorio): Nombre del grupo (case-sensitive)

**Efectos en las Estructuras**:
- **Escritura**: Modifica el archivo `users.txt`
- Agrega nueva línea: `GID, G, nombre_grupo`
- Incrementa el contador de GID
- Actualiza los bitmaps e inodos del archivo users.txt

**Formato en users.txt**:
```
2, G, usuarios
```

**Ejemplo**:
```bash
mkgrp -name=usuarios
```

### 11. RMGRP
**Propósito**: Elimina un grupo existente (solo usuario root).

**Parámetros**:
- `-name` (Obligatorio): Nombre del grupo a eliminar

**Efectos en las Estructuras**:
- **Escritura**: Modifica el archivo `users.txt`
- Cambia el GID a 0 (marca como eliminado)
- Actualiza los bloques del archivo users.txt

**Formato en users.txt después de eliminar**:
```
0, G, usuarios
```

**Ejemplo**:
```bash
rmgrp -name=usuarios
```

### 12. MKUSR
**Propósito**: Crea un nuevo usuario (solo usuario root).

**Parámetros**:
- `-user` (Obligatorio): Nombre del usuario (máximo 10 caracteres)
- `-pass` (Obligatorio): Contraseña (máximo 10 caracteres)
- `-grp` (Obligatorio): Grupo al que pertenece (debe existir)

**Efectos en las Estructuras**:
- **Escritura**: Modifica el archivo `users.txt`
- Agrega nueva línea: `UID, U, grupo, usuario, contraseña`
- Incrementa el contador de UID
- Valida existencia del grupo

**Formato en users.txt**:
```
2, U, usuarios, user1, usuario
```

**Ejemplo**:
```bash
mkusr -user=user1 -pass=usuario -grp=usuarios
```

### 13. RMUSR
**Propósito**: Elimina un usuario existente (solo usuario root).

**Parámetros**:
- `-user` (Obligatorio): Nombre del usuario a eliminar

**Efectos en las Estructuras**:
- **Escritura**: Modifica el archivo `users.txt`
- Cambia el UID a 0 (marca como eliminado)

**Ejemplo**:
```bash
rmusr -user=user1
```

### 14. CHGRP
**Propósito**: Cambia el grupo de un usuario existente (solo usuario root).

**Parámetros**:
- `-user` (Obligatorio): Usuario al que cambiar el grupo
- `-grp` (Obligatorio): Nuevo grupo (debe existir)

**Efectos en las Estructuras**:
- **Escritura**: Modifica el archivo `users.txt`
- Actualiza la línea del usuario con el nuevo grupo

**Ejemplo**:
```bash
chgrp -user=user2 -grp=grupo1
```

---

## Administración de Carpetas, Archivos y Permisos

### 15. MKFILE
**Propósito**: Crea un archivo nuevo en el sistema de archivos.

**Parámetros**:
- `-path` (Obligatorio): Ruta completa del archivo a crear
- `-r` (Opcional): Crea carpetas padres si no existen
- `-size` (Opcional): Tamaño en bytes (contenido: números 0-9 repetidos)
- `-cont` (Opcional): Ruta a archivo local para copiar contenido

**Permisos por Defecto**: 664 (rw-rw-r--)

**Efectos en las Estructuras**:
- **Creación de Inodo**: Nuevo inodo con `I_type=0` (archivo)
- **Asignación de Bloques**: Según el tamaño del archivo
- **Actualización de Bitmap Inodos**: Marca el inodo como ocupado (1)
- **Actualización de Bitmap Bloques**: Marca los bloques usados (1)
- **Modificación de Carpeta Padre**: Agrega entrada en bloque de carpeta
- **Punteros Directos/Indirectos**: Según el tamaño del archivo

**Manejo de Bloques por Tamaño**:
- 0-764 bytes: Solo punteros directos (12 bloques)
- 765-16148 bytes: Incluye puntero indirecto simple
- Más bytes: Punteros dobles y triples indirectos

**Ejemplos**:
```bash
# Archivo con tamaño específico
mkfile -size=15 -path=/home/user/docs/a.txt -r

# Archivo con contenido de archivo local
mkfile -path=/home/user/docs/b.txt -r -cont=/home/Documents/b.txt

# Archivo vacío
mkfile -path="/home/mis documentos/archivo 1.txt"
```

### 16. MKDIR
**Propósito**: Crea una carpeta nueva en el sistema de archivos.

**Parámetros**:
- `-path` (Obligatorio): Ruta completa de la carpeta a crear
- `-p` (Opcional): Crea carpetas padres si no existen

**Permisos por Defecto**: 664 (rw-rw-r--)

**Efectos en las Estructuras**:
- **Creación de Inodo**: Nuevo inodo con `I_type=1` (carpeta)
- **Creación de Bloque Carpeta**: Con entradas "." y ".."
- **Actualización de Bitmaps**: Marca inodo y bloque como ocupados
- **Modificación de Carpeta Padre**: Agrega entrada para la nueva carpeta

**Estructura del Bloque Carpeta**:
```
Entrada 1: "." -> inodo actual
Entrada 2: ".." -> inodo padre
Entrada 3: (vacía)
Entrada 4: (vacía)
```

**Ejemplos**:
```bash
# Crear carpeta con padres
mkdir -p -path=/home/user/docs/usac

# Crear carpeta simple
mkdir -path="/home/mis documentos/archivos clases"
```

---

## Reportes

### 17. REP
**Propósito**: Genera reportes visuales del sistema de archivos usando Graphviz.

**Parámetros**:
- `-name` (Obligatorio): Tipo de reporte a generar
- `-path` (Obligatorio): Ruta y nombre del archivo de salida
- `-id` (Obligatorio): ID de la partición a reportar
- `-path_file_ls` (Opcional): Para reportes `file` y `ls`

**Tipos de Reportes Disponibles**:

#### MBR
Muestra información detallada del Master Boot Record y EBRs.
```bash
rep -id=A118 -path=/home/user/reports/reporte1.jpg -name=mbr
```

#### DISK
Visualiza la estructura de particiones y porcentajes de uso.
```bash
rep -id=A118 -path=/home/user/reports/report2.pdf -name=disk
```

#### INODE
Muestra todos los inodos utilizados con su información completa.
```bash
rep -id=A118 -path=/home/user/reports/report3.jpg -name=inode
```

#### BLOCK
Muestra información de todos los bloques utilizados.
```bash
rep -id=A118 -path=/home/user/reports/report4.jpg -name=block
```

#### BM_INODE
Genera bitmap de inodos en formato texto (20 registros por línea).
```bash
rep -id=A118 -path=/home/user/reports/report5.txt -name=bm_inode
```

#### BM_BLOCK
Genera bitmap de bloques en formato texto (20 registros por línea).
```bash
rep -id=A118 -path=/home/user/reports/report6.txt -name=bm_block
```

#### TREE
Genera árbol completo del sistema de archivos EXT2.
```bash
rep -id=A118 -path=/home/user/reports/report7.jpg -name=tree
```

#### SB
Muestra información del superbloque en formato tabla.
```bash
rep -id=A118 -path=/home/user/reports/report8.jpg -name=sb
```

#### FILE
Muestra nombre y contenido completo de un archivo específico.
```bash
rep -id=A118 -path=/home/user/reports/report9.txt -path_file_ls=/home/a.txt -name=file
```

#### LS
Lista archivos y carpetas con permisos, propietarios, fechas y tipos.
```bash
rep -id=A118 -path=/home/user/reports/report10.jpg -path_file_ls=/ -name=ls
```

---

## Notas Importantes

### Sistema de Permisos
- **Usuario Root**: Tiene permisos 777 sobre todos los archivos y carpetas
- **Permisos UGO**: User, Group, Other en formato octal
- **Permisos por Defecto**: 664 para archivos y carpetas nuevos

### Limitaciones y Restricciones
- Solo particiones primarias pueden ser montadas para el sistema de archivos
- Tamaño de bloque fijo: 64 bytes para contenido útil
- Máximo 4 entradas por bloque de carpeta
- Sistema case-sensitive para usuarios y grupos

### Estructura del Archivo users.txt
```
1, G, root
1, U, root, root, 123
2, G, nombre_grupo
2, U, nombre_grupo, nombre_usuario, contraseña
```

### Comandos que Requieren Sesión Activa
Todos los comandos excepto MKFS y LOGIN requieren una sesión activa iniciada con LOGIN.