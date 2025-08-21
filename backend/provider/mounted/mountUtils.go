package mounted

import "fmt"

// GetMountedPartition obtiene una partición montada por su ID
func GetMountedPartition(id string) (*MountedPartition, error) {
	manager := GetMountManager()
	partition, found := manager.GetMountedPartition(id)
	if !found {
		return nil, fmt.Errorf("partición con ID %s no encontrada", id)
	}
	return &partition, nil
}
