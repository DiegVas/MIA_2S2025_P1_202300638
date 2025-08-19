package mounted

import "sync"

var (
	mountManager     *MountManager
	mountManagerOnce sync.Once
)

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
