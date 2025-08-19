package files

import (
	"errors"
	"main/structs"
	"os"
)

func FindFreeBlocks(file *os.File, bmBlockPos, blockCount, blocksNeeded int32) ([]int32, error) {
	bitmap := make([]byte, blockCount)
	_, err := file.ReadAt(bitmap, int64(bmBlockPos))
	if err != nil {
		return nil, err
	}

	var freeBlocks []int32
	for i := int32(2); i < blockCount && int32(len(freeBlocks)) < blocksNeeded; i++ {
		if bitmap[i] == 0 {
			freeBlocks = append(freeBlocks, i)
		}
	}

	if int32(len(freeBlocks)) < blocksNeeded {
		return nil, errors.New("no hay suficientes bloques libres disponibles")
	}

	return freeBlocks, nil
}

func WriteContentToBlocks(file *os.File, content string, blocks []int32, blockPos int32) error {
	contentBytes := []byte(content)
	for i, blockNum := range blocks {
		blockData := make([]byte, structs.BLOCK_SIZE)

		start := i * int(structs.BLOCK_SIZE)
		end := start + int(structs.BLOCK_SIZE)
		if end > len(contentBytes) {
			end = len(contentBytes)
		}

		if start < len(contentBytes) {
			copy(blockData, contentBytes[start:end])
		}

		_, err := file.WriteAt(blockData, int64(blockPos+(blockNum*structs.BLOCK_SIZE)))
		if err != nil {
			return err
		}
	}
	return nil
}

func MarkBlockAsUsed(file *os.File, bmBlockPos, blockNum int32) error {
	_, err := file.WriteAt([]byte{1}, int64(bmBlockPos+blockNum))
	return err
}
