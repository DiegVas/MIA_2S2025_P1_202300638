package structs

// MkfsParams estructura para los parámetros del comando MKFS
type MkfsParams struct {
	ID   string // ID de la partición montada
	Type string // Tipo de formateo (full por defecto)
}

func (p *MkfsParams) String() string {
	return "ID: " + p.ID + ", Type: " + p.Type
}
