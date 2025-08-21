package structs

import "fmt"

type MkdiskParams struct {
	Size int
	Fit  string
	Unit byte
	Path string
}

func (p *MkdiskParams) GetSizeInBytes() int32 {
	switch p.Unit {
	case 'K':
		return int32(p.Size) * 1024
	case 'M':
		return int32(p.Size) * 1024 * 1024
	default:
		return int32(p.Size) * 1024 * 1024 // Por defecto Megabytes
	}
}

func (p *MkdiskParams) String() string {
	return fmt.Sprintf("Size: %d, Fit: %s, Unit: %c, Path: %s",
		p.Size, p.Fit, p.Unit, p.Path)
}

type FdiskParams struct {
	Size int
	Unit byte
	Path string
	Type byte
	Fit  string
	Name string
}

func (p *FdiskParams) GetSizeInBytes() int32 {
	switch p.Unit {
	case 'B':
		return int32(p.Size)
	case 'K':
		return int32(p.Size) * 1024
	case 'M':
		return int32(p.Size) * 1024 * 1024
	default:
		return int32(p.Size) * 1024 // Por defecto Kilobytes
	}
}

type RepParams struct {
	Name       string // mbr, disk, inode, block, bm_inode, bm_block, tree, sb, file, ls
	Path       string // Ruta donde se guardará el reporte
	ID         string // ID de la partición montada
	PathFileLS string // Para reportes file y ls (opcional)
}

func (p *RepParams) String() string {
	return fmt.Sprintf("Name: %s, Path: %s, ID: %s, PathFileLS: %s",
		p.Name, p.Path, p.ID, p.PathFileLS)
}

func (p *FdiskParams) String() string {
	return fmt.Sprintf("Size: %d, Unit: %c, Path: %s, Type: %c, Fit: %s, Name: %s",
		p.Size, p.Unit, p.Path, p.Type, p.Fit, p.Name)
}

type MountParams struct {
	Path   string
	Name   string
	ID     string
	Carnet string // Carnet del estudiante, se usarán los últimos 2 dígitos
}

func (p *MountParams) String() string {
	return fmt.Sprintf("Path: %s, Name: %s, ID: %s, Carnet: %s",
		p.Path, p.Name, p.ID, p.Carnet)
}

type MkdirParams struct {
	Path string // Ruta de la carpeta a crear
	P    bool   // Si debe crear directorios padre
}

func (p *MkdirParams) String() string {
	return fmt.Sprintf("Path: %s, P: %t", p.Path, p.P)
}

type ViewUsersParams struct {
	ID string // ID de la partición montada
}

func (p *ViewUsersParams) String() string {
	return fmt.Sprintf("ID: %s", p.ID)
}
