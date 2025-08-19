package files

import (
	"errors"
	"os"
)

func FindFreeInode(file *os.File, bmInodePos, inodeCount int32) (int32, error) {
	bitmap := make([]byte, inodeCount)
	_, err := file.ReadAt(bitmap, int64(bmInodePos))
	if err != nil {
		return -1, err
	}

	for i := int32(2); i < inodeCount; i++ { // Comenzar desde 2 (después de raíz y users.txt)
		if bitmap[i] == 0 {
			return i, nil
		}
	}
	return -1, errors.New("no hay inodos libres disponibles")
}

func MarkInodeAsUsed(file *os.File, bmInodePos, inodeNum int32) error {
	_, err := file.WriteAt([]byte{1}, int64(bmInodePos+inodeNum))
	return err
}
