package structs

type Superblock struct {
	S_filesystem_type   int32    // Tipo de sistema de archivos (ext2=2, ext3=3)
	S_inodes_count      int32    // Número total de inodos
	S_blocks_count      int32    // Número total de bloques
	S_free_blocks_count int32    // Número de bloques libres
	S_free_inodes_count int32    // Número de inodos libres
	S_mtime             [16]byte // Última fecha de montaje
	S_umtime            [16]byte // Última fecha de desmontaje
	S_mnt_count         int32    // Número de veces que se ha montado
	S_magic             int32    // Número mágico para validar el sistema de archivos (0xEF53)
	S_inode_size        int32    // Tamaño del inodo
	S_block_size        int32    // Tamaño del bloque
	S_first_ino         int32    // Primer inodo libre
	S_first_blo         int32    // Primer bloque libre
	S_bm_inode_start    int32    // Inicio del bitmap de inodos
	S_bm_block_start    int32    // Inicio del bitmap de bloques
	S_inode_start       int32    // Inicio de la tabla de inodos
	S_block_start       int32    // Inicio de la tabla de bloques
}

// Inodo - Estructura de inodo en el sistema de archivos ext2
type Inode struct {
	I_uid   int32     // ID del usuario propietario
	I_gid   int32     // ID del grupo
	I_size  int32     // Tamaño en bytes
	I_atime [16]byte  // Última fecha en que se accedió
	I_ctime [16]byte  // Fecha de creación
	I_mtime [16]byte  // Última fecha de modificación
	I_block [15]int32 // Punteros directos e indirectos
	I_type  int32     // Tipo de inodo (0=archivo, 1=carpeta) - cambiado a int32
	I_perm  int32     // Permisos
}

// Folder - Estructura para almacenar una carpeta
type FolderBlock struct {
	B_content [4]Content // Contenido de la carpeta (4 entradas)
}

// Content - Estructura para cada entrada en una carpeta
type Content struct {
	B_name  [12]byte // Nombre del archivo/carpeta
	B_inodo int32    // Número de inodo
}

// FileBlock - Estructura para almacenar un archivo
type FileBlock struct {
	B_content [64]byte
}

// PointerBlock - Estructura para almacenar punteros a otros bloques
type PointerBlock struct {
	B_pointers [16]int32
}

// Tamaños de las estructuras en bytes
const (
	SUPERBLOCK_SIZE   = 88   // Tamaño del superbloque
	INODE_SIZE        = 128  // Tamaño de un inodo (actualizado para coincidir con struct real)
	BLOCK_SIZE        = 1024 // Tamaño de un bloque (aumentado de 64 a 1024)
	FILE_CONTENT_SIZE = 64   // Tamaño del contenido útil para archivos
	FOLDER_SIZE       = 64   // Tamaño de una carpeta
	POINTER_SIZE      = 64   // Tamaño de un bloque de punteros
)

// Constantes para el sistema de archivos
const (
	EXT2_MAGIC  = 0xEF53      // Número mágico para ext2
	ROOT_INODE  = 0           // Número de inodo raíz
	ROOT_FOLDER = "/"         // Nombre de la carpeta raíz
	USERS_FILE  = "users.txt" // Nombre del archivo de usuarios
)

// Usuario - Estructura para almacenar información de usuario
type User struct {
	Username string
	Password string
	Group    string
	UID      int32
	GID      int32
}
