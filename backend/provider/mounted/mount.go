package mounted

import (
	"fmt"
	"sync"
)

type MountedPartition struct {
	Path        string // Ruta del disco
	DiskPath    string // Alias para Path (compatibilidad)
	Name        string
	ID          string
	Correlative int
	Letter      byte
}

type MountManager struct {
	MountedPartitions []MountedPartition
	DiskLetters       map[string]byte
	DiskCounters      map[string]int
	Mutex             sync.Mutex
}

func (mm *MountManager) AddMountedPartition(path, name string, carnet string) (string, int, error) {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()

	suffix := carnet[len(carnet)-2:]

	var letter byte
	var counter int

	diskLetter, diskExists := mm.DiskLetters[path]

	if diskExists {

		letter = diskLetter
		mm.DiskCounters[path]++
		counter = mm.DiskCounters[path]
	} else {

		lastLetter := byte('A' - 1)

		for _, l := range mm.DiskLetters {
			if l > lastLetter {
				lastLetter = l
			}
		}

		letter = lastLetter + 1
		if letter < 'A' {
			letter = 'A'
		}

		mm.DiskLetters[path] = letter

		mm.DiskCounters[path] = 1
		counter = 1
	}

	id := suffix + fmt.Sprintf("%d", counter) + string(letter)

	mountedPartition := MountedPartition{
		Path:        path,
		DiskPath:    path, // Mantener compatibilidad
		Name:        name,
		ID:          id,
		Correlative: counter,
		Letter:      letter,
	}

	mm.MountedPartitions = append(mm.MountedPartitions, mountedPartition)

	return id, counter, nil
}

func (mm *MountManager) GetMountedPartitions() []MountedPartition {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()

	partitions := make([]MountedPartition, len(mm.MountedPartitions))
	copy(partitions, mm.MountedPartitions)

	return partitions
}

func (mm *MountManager) IsMounted(path, name string) bool {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()

	for _, partition := range mm.MountedPartitions {
		if partition.Path == path && partition.Name == name {
			return true
		}
	}

	return false
}

func (mm *MountManager) GetMountedPartition(id string) (MountedPartition, bool) {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()

	for _, partition := range mm.MountedPartitions {
		if partition.ID == id {
			return partition, true
		}
	}

	return MountedPartition{}, false
}

func (mm *MountManager) UnmountPartition(id string) bool {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()

	for i, partition := range mm.MountedPartitions {
		if partition.ID == id {
			mm.MountedPartitions = append(mm.MountedPartitions[:i], mm.MountedPartitions[i+1:]...)
			return true
		}
	}

	return false
}
