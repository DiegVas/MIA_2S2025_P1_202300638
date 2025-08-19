package diskmanager

import (
	"fmt"
	"main/provider/mounted"
	"main/structs"
)

func FormatPartition(params *structs.MkfsParams) error {
	mountManager := mounted.GetMountManager()
	_, found := mountManager.GetMountedPartition(params.ID)

	if !found {
		return fmt.Errorf("no se encontró ninguna partición montada con el ID '%s'", params.ID)
	}

	// Realizar el formateo según el tipo especificado (solo 'full' por ahora)
	if params.Type == "full" {
		// Formatear la partición con ext2
		err := mounted.FormatPartition(params)
		if err != nil {
			return fmt.Errorf("error al formatear la partición: %v", err)
		}
	} else {
		return fmt.Errorf("tipo de formateo no soportado: '%s'", params.Type)
	}

	return nil
}
