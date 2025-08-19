package users

import (
	"bufio"
	"errors"
	"fmt"
	"main/provider/mounted"
	"main/structs"
	"main/utils"
	"os"
	"strings"
)

func LoginUser(params *structs.LoginParams) error {
	session := structs.GetSession()

	if session.Active {
		return errors.New("ya hay una sesión activa, debe hacer logout primero")
	}

	mountManager := mounted.GetMountManager()
	partition, found := mountManager.GetMountedPartition(params.ID)

	if !found {
		return fmt.Errorf("no se encontró ninguna partición montada con el ID '%s'", params.ID)
	}

	mbr, err := utils.ReadMBR(partition.Path)

	if err != nil {
		return fmt.Errorf("error al leer el MBR del disco: %v", err)
	}

	var targetPartition structs.PARTITION
	found = false

	for _, p := range mbr.Partitions {
		partName := strings.TrimRight(string(p.Part_name[:]), "\x00")
		if partName == partition.Name {
			targetPartition = p
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("la partición '%s' no se encontró en el disco", partition.Name)
	}

	start := targetPartition.Part_start
	inodeCount := int32(float64(targetPartition.Part_size) * 0.05 / float64(structs.INODE_SIZE))
	blockCount := (targetPartition.Part_size - inodeCount*structs.INODE_SIZE) / structs.BLOCK_SIZE
	inodePos := start + structs.SUPERBLOCK_SIZE + inodeCount + blockCount
	blockPos := inodePos + (inodeCount * structs.INODE_SIZE)
	usersBlockPos := blockPos + structs.BLOCK_SIZE
	file, err := os.Open(partition.Path)

	if err != nil {
		return fmt.Errorf("error al abrir el disco: %v", err)
	}

	defer file.Close()

	file.Seek(int64(usersBlockPos), 0)
	block := make([]byte, structs.BLOCK_SIZE)
	_, err = file.Read(block)

	if err != nil {
		return fmt.Errorf("error al leer users.txt: %v", err)
	}

	content := string(block)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}
		if len(fields) == 5 && fields[1] == "U" {
			if fields[3] == params.User {
				if fields[4] == params.Pass {
					session.Login(fields[3], params.ID, fields[2])
					return nil
				} else {
					return errors.New("autenticación fallida: contraseña incorrecta")
				}
			}
		}
	}
	return errors.New("usuario no existe")
}
