package structs

import "fmt"

type MkdiskParams struct {
    Size int    
    Fit  string 
    Unit byte 
    Path string 
}

func (p *MkdiskParams) GetSizeInBytes() int64 {
    switch p.Unit {
    case 'K':
        return int64(p.Size) * 1024
    case 'M':
        return int64(p.Size) * 1024 * 1024
    default:
        return int64(p.Size) * 1024 * 1024 // Por defecto Megabytes
    }
}

func (p *MkdiskParams) String() string {
    return fmt.Sprintf("Size: %d, Fit: %s, Unit: %s, Path: %s", 
        p.Size, p.Fit, p.Unit, p.Path)
}