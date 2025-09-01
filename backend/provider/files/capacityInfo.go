package files

import (
	"fmt"
)

// DirectoryCapacityInfo proporciona información detallada sobre la capacidad de directorios
type DirectoryCapacityInfo struct {
	MaxEntriesDirectBlocks      int32 // 12 bloques × 4 entradas = 48
	MaxEntriesSimpleIndirect    int32 // 16 bloques × 4 entradas = 64
	MaxEntriesDoubleIndirect    int32 // 16 × 16 × 4 = 1,024
	MaxEntriesTripleIndirect    int32 // 16 × 16 × 16 × 4 = 16,384
	TotalMaxEntries             int32 // Suma de todos los anteriores
	CurrentImplementationStatus string
}

// GetDirectoryCapacityInfo retorna información sobre las capacidades del sistema de archivos
func GetDirectoryCapacityInfo() *DirectoryCapacityInfo {
	info := &DirectoryCapacityInfo{
		MaxEntriesDirectBlocks:   12 * 4,           // 48 entradas
		MaxEntriesSimpleIndirect: 16 * 4,           // 64 entradas
		MaxEntriesDoubleIndirect: 16 * 16 * 4,      // 1,024 entradas
		MaxEntriesTripleIndirect: 16 * 16 * 16 * 4, // 16,384 entradas
	}

	info.TotalMaxEntries = info.MaxEntriesDirectBlocks +
		info.MaxEntriesSimpleIndirect +
		info.MaxEntriesDoubleIndirect +
		info.MaxEntriesTripleIndirect

	info.CurrentImplementationStatus = fmt.Sprintf(
		"✅ Sistema EXT2 Completo Implementado\n"+
			"📊 Capacidades por nivel:\n"+
			"  • Bloques directos (0-11): %d entradas\n"+
			"  • Indirecto simple (12): %d entradas\n"+
			"  • Indirecto doble (13): %d entradas\n"+
			"  • Indirecto triple (14): %d entradas\n"+
			"🎯 Capacidad total máxima: %d entradas por directorio",
		info.MaxEntriesDirectBlocks,
		info.MaxEntriesSimpleIndirect,
		info.MaxEntriesDoubleIndirect,
		info.MaxEntriesTripleIndirect,
		info.TotalMaxEntries,
	)

	return info
}

// CalculateBlockUsageLevel determina en qué nivel de indirección se encuentra un directorio
func CalculateBlockUsageLevel(entriesCount int32) string {
	info := GetDirectoryCapacityInfo()

	if entriesCount <= info.MaxEntriesDirectBlocks {
		return fmt.Sprintf("🟢 Nivel: Bloques Directos (%d/%d entradas)", entriesCount, info.MaxEntriesDirectBlocks)
	}

	if entriesCount <= info.MaxEntriesDirectBlocks+info.MaxEntriesSimpleIndirect {
		return fmt.Sprintf("🟡 Nivel: Indirecto Simple (%d/%d entradas)",
			entriesCount, info.MaxEntriesDirectBlocks+info.MaxEntriesSimpleIndirect)
	}

	if entriesCount <= info.MaxEntriesDirectBlocks+info.MaxEntriesSimpleIndirect+info.MaxEntriesDoubleIndirect {
		return fmt.Sprintf("🟠 Nivel: Indirecto Doble (%d/%d entradas)",
			entriesCount, info.MaxEntriesDirectBlocks+info.MaxEntriesSimpleIndirect+info.MaxEntriesDoubleIndirect)
	}

	return fmt.Sprintf("🔴 Nivel: Indirecto Triple (%d/%d entradas)", entriesCount, info.TotalMaxEntries)
}

// EstimatePerformanceImpact estima el impacto en rendimiento según el nivel de indirección
func EstimatePerformanceImpact(entriesCount int32) string {
	info := GetDirectoryCapacityInfo()

	if entriesCount <= info.MaxEntriesDirectBlocks {
		return "⚡ Rendimiento: ÓPTIMO (acceso directo)"
	}

	if entriesCount <= info.MaxEntriesDirectBlocks+info.MaxEntriesSimpleIndirect {
		return "🚀 Rendimiento: BUENO (1 indirección adicional)"
	}

	if entriesCount <= info.MaxEntriesDirectBlocks+info.MaxEntriesSimpleIndirect+info.MaxEntriesDoubleIndirect {
		return "⚠️  Rendimiento: MODERADO (2 indirecciones adicionales)"
	}

	return "🐌 Rendimiento: LENTO (3 indirecciones adicionales)"
}
