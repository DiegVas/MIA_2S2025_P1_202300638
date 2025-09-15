package reports

import (
	"fmt"
	"main/provider/mounted"
	"os"
	"strings"
	"time"
)

func generateLSReport(partitionID, dirPath string) (string, error) {
	// Validar parámetros
	if dirPath == "" {
		dirPath = "/"
	}

	// Obtener información de la partición montada
	mountManager := mounted.GetMountManager()
	mountedPartition, found := mountManager.GetMountedPartition(partitionID)
	if !found {
		return "", fmt.Errorf("partición con ID '%s' no está montada", partitionID)
	}

	// Leer el superbloque
	superblock, err := readSuperblock(mountedPartition.Path, partitionID)
	if err != nil {
		return "", fmt.Errorf("error leyendo superbloque: %v", err)
	}

	// Abrir el archivo del disco
	file, err := os.Open(mountedPartition.Path)
	if err != nil {
		return "", fmt.Errorf("error abriendo disco: %v", err)
	}
	defer file.Close()

	// Buscar el inodo del directorio especificado
	dirInodeNum, err := findInodeByPath(file, dirPath, superblock)
	if err != nil {
		return "", fmt.Errorf("error encontrando directorio '%s': %v", dirPath, err)
	}

	// Leer el inodo del directorio
	dirInode, err := readInodeAt(file, dirInodeNum, superblock.S_inode_start)
	if err != nil {
		return "", fmt.Errorf("error leyendo inodo del directorio: %v", err)
	}

	// Verificar que sea un directorio
	if dirInode.I_type != 1 {
		return "", fmt.Errorf("'%s' no es un directorio", dirPath)
	}

	// Obtener entradas del directorio
	entries, err := getDirectoryEntries(file, dirInode, superblock)
	if err != nil {
		return "", fmt.Errorf("error leyendo entradas del directorio: %v", err)
	}

	// Generar contenido DOT
	var dot strings.Builder
	dot.WriteString("digraph LS_Report {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=plaintext];\n")
	dot.WriteString("    edge [style=invisible];\n")

	// Leer archivo users.txt para obtener información de usuarios
	usersInfo, err := readUsersFile(file, superblock)
	if err != nil {
		// Si no existe users.txt, usar UIDs/GIDs numéricos
		usersInfo = make(map[int32]UserInfo) // Crear mapa vacío si no se puede leer
	}

	// Título del reporte
	dot.WriteString("    title [label=<<TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" BGCOLOR=\"lightblue\">\n")
	dot.WriteString(fmt.Sprintf("        <TR><TD COLSPAN=\"7\"><B>Reporte LS - Directorio: %s</B></TD></TR>\n", dirPath))
	dot.WriteString("        <TR>\n")
	dot.WriteString("            <TD><B>Permisos</B></TD>\n")
	dot.WriteString("            <TD><B>Owner</B></TD>\n")
	dot.WriteString("            <TD><B>Group</B></TD>\n")
	dot.WriteString("            <TD><B>Size</B></TD>\n")
	dot.WriteString("            <TD><B>Date</B></TD>\n")
	dot.WriteString("            <TD><B>Type</B></TD>\n")
	dot.WriteString("            <TD><B>Name</B></TD>\n")
	dot.WriteString("        </TR>\n")

	// Procesar cada entrada
	for _, entry := range entries {
		if entry.Name == "" {
			continue
		}

		// Leer el inodo de la entrada
		entryInode, err := readInodeAt(file, entry.InodeNum, superblock.S_inode_start)
		if err != nil {
			continue // Saltar entradas con error
		}

		// Obtener información de la entrada
		perms := formatPermissions(entryInode.I_perm)

		// Obtener nombres de usuario y grupo desde users.txt
		ownerName := fmt.Sprintf("%d", entryInode.I_uid)
		groupName := fmt.Sprintf("%d", entryInode.I_gid)
		if userInfo, exists := usersInfo[entryInode.I_uid]; exists {
			ownerName = userInfo.Username
			groupName = userInfo.Group
		}

		size := fmt.Sprintf("%d", entryInode.I_size)

		// Formatear fecha de creación
		creationTime := strings.TrimRight(string(entryInode.I_ctime[:]), "\x00")
		date := "N/A"

		if creationTime != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", creationTime); err == nil {
				date = t.Format("2006-01-02")
			}
		}

		// Determinar tipo
		fileType := "file"
		if entryInode.I_type == 1 {
			fileType = "dir"
		}

		// Agregar fila a la tabla (sin columna de Time)
		dot.WriteString("        <TR>\n")
		dot.WriteString(fmt.Sprintf("            <TD>%s</TD>\n", perms))
		dot.WriteString(fmt.Sprintf("            <TD>%s</TD>\n", ownerName))
		dot.WriteString(fmt.Sprintf("            <TD>%s</TD>\n", groupName))
		dot.WriteString(fmt.Sprintf("            <TD>%s</TD>\n", size))
		dot.WriteString(fmt.Sprintf("            <TD>%s</TD>\n", date))
		dot.WriteString(fmt.Sprintf("            <TD>%s</TD>\n", fileType))
		dot.WriteString(fmt.Sprintf("            <TD>%s</TD>\n", entry.Name))
		dot.WriteString("        </TR>\n")
	}

	dot.WriteString("    </TABLE>>];\n")
	dot.WriteString("}\n")

	return dot.String(), nil
}
