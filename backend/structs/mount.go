package structs

import (
	"fmt"
	"sync"
)

// MountedPartition representa una partición montada en el sistema
type MountedPartition struct {
	Path        string // Ruta del disco
	Name        string // Nombre de la partición
	ID          string // ID asignado (ej: 341A)
	Correlative int    // Número correlativo de la partición
	Letter      byte   // Letra asignada (A, B, C, etc.)
}

// MountManager gestiona las particiones montadas en memoria
type MountManager struct {
	MountedPartitions []MountedPartition
	DiskLetters       map[string]byte     // Mapa de disco a letra
	DiskCounters      map[string]int      // Contador de particiones por disco
	Mutex             sync.Mutex          // Para operaciones concurrentes seguras
}

var (
	mountManager     *MountManager
	mountManagerOnce sync.Once
)

// GetMountManager retorna la instancia singleton del gestor de montajes
func GetMountManager() *MountManager {
	mountManagerOnce.Do(func() {
		mountManager = &MountManager{
			MountedPartitions: make([]MountedPartition, 0),
			DiskLetters:       make(map[string]byte),
			DiskCounters:      make(map[string]int),
		}
	})
	return mountManager
}

// AddMountedPartition agrega una partición a la lista de montadas
func (mm *MountManager) AddMountedPartition(path, name string, carnet string) (string, int, error) {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()

	// Extraer los últimos dos dígitos del carnet
	suffix := carnet[len(carnet)-2:]
	
	// Obtener la siguiente letra disponible para este disco
	var letter byte
	var counter int
	
	// Verificar si el disco ya tiene una letra asignada
	diskLetter, diskExists := mm.DiskLetters[path]
	
	if diskExists {
		// Si el disco ya existe, usar la misma letra
		letter = diskLetter
		// Incrementar el contador para este disco
		mm.DiskCounters[path]++
		counter = mm.DiskCounters[path]
	} else {
		// Si es un nuevo disco, asignar la siguiente letra disponible
		// Encontrar la última letra usada
		lastLetter := byte('A' - 1) // Empezar desde antes de 'A'
		
		for _, l := range mm.DiskLetters {
			if l > lastLetter {
				lastLetter = l
			}
		}
		
		// La siguiente letra es la última + 1
		letter = lastLetter + 1
		if letter < 'A' { // Si no hay letras usadas, comenzar con 'A'
			letter = 'A'
		}
		
		// Registrar la nueva letra para este disco
		mm.DiskLetters[path] = letter
		// Iniciar el contador para este disco
		mm.DiskCounters[path] = 1
		counter = 1
	}
	
	// Construir el ID: últimos dos dígitos del carnet + número de partición + letra
	// Ejemplo: carnet = 34, counter = 1, letter = A -> ID = "341A"
	id := suffix + fmt.Sprintf("%d", counter) + string(letter)
	
	// Crear la nueva partición montada
	mountedPartition := MountedPartition{
		Path:        path,
		Name:        name,
		ID:          id,
		Correlative: counter,
		Letter:      letter,
	}
	
	// Agregar a la lista de particiones montadas
	mm.MountedPartitions = append(mm.MountedPartitions, mountedPartition)
	
	return id, counter, nil
}

// GetMountedPartitions retorna todas las particiones montadas
func (mm *MountManager) GetMountedPartitions() []MountedPartition {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()
	
	// Crear una copia para evitar problemas de concurrencia
	partitions := make([]MountedPartition, len(mm.MountedPartitions))
	copy(partitions, mm.MountedPartitions)
	
	return partitions
}

// IsMounted verifica si una partición ya está montada
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

// GetMountedPartition busca una partición montada por su ID
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

// UnmountPartition desmonta una partición por su ID
func (mm *MountManager) UnmountPartition(id string) bool {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()
	
	for i, partition := range mm.MountedPartitions {
		if partition.ID == id {
			// Eliminar la partición del slice
			mm.MountedPartitions = append(mm.MountedPartitions[:i], mm.MountedPartitions[i+1:]...)
			return true
		}
	}
	
	return false
}
