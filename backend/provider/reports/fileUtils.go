package reports

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"main/structs"
	"os"
	"strconv"
	"strings"
)

func readFileContent(file *os.File, fileInode *structs.Inode, superblock *structs.Superblock) (string, error) {
	var content strings.Builder

	// Leer bloques directos
	for i := 0; i < 12; i++ {
		if fileInode.I_block[i] != 0 {
			blockBytes := make([]byte, structs.BLOCK_SIZE)

			var blockPos int32
			if superblock.S_block_start == 0 {
				blockPos = superblock.S_inode_start + (superblock.S_inodes_count * structs.INODE_SIZE)
				blockPos += (fileInode.I_block[i] * structs.BLOCK_SIZE)
			} else {
				blockPos = superblock.S_block_start + (fileInode.I_block[i] * structs.BLOCK_SIZE)
			}

			_, err := file.ReadAt(blockBytes, int64(blockPos))
			if err != nil {
				continue
			}

			// Leer como FileBlock
			var fileBlock structs.FileBlock
			buffer := bytes.NewReader(blockBytes)
			err = binary.Read(buffer, binary.LittleEndian, &fileBlock)
			if err != nil {
				continue
			}

			blockContent := strings.TrimRight(string(fileBlock.B_content[:]), "\x00")
			content.WriteString(blockContent)
		}
	}

	return content.String(), nil
}

// readUsersFile lee el archivo users.txt del sistema de archivos y extrae información de usuarios
func readUsersFile(file *os.File, superblock *structs.Superblock) (map[int32]UserInfo, error) {
	usersMap := make(map[int32]UserInfo)

	// Buscar el archivo users.txt de manera recursiva en el sistema de archivos
	usersInodeNum, err := findFileRecursively(file, "users.txt", 0, superblock)
	if err != nil {
		return usersMap, fmt.Errorf("no se encontró users.txt: %v", err)
	}

	// Leer el inodo del archivo users.txt
	usersInode, err := readInodeAt(file, usersInodeNum, superblock.S_inode_start)
	if err != nil {
		return usersMap, fmt.Errorf("error leyendo inodo de users.txt: %v", err)
	}

	// Leer el contenido del archivo users.txt
	content, err := readFileContent(file, usersInode, superblock)
	if err != nil {
		return usersMap, fmt.Errorf("error leyendo contenido de users.txt: %v", err)
	}

	// Parsear el contenido del archivo users.txt
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) >= 5 {
			// Formato: ID,tipo,grupo,nombre,password
			userType := strings.TrimSpace(parts[1])
			if userType == "U" { // Es un usuario
				if uid, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
					username := strings.TrimSpace(parts[3])
					groupName := strings.TrimSpace(parts[2])

					usersMap[int32(uid)] = UserInfo{
						Username: username,
						Group:    groupName,
						UID:      int32(uid),
						GID:      0, // Se puede calcular después si es necesario
					}
				}
			}
		}
	}

	return usersMap, nil
}
