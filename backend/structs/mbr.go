package structs

type MBR struct {
	Mbr_size       int32         
	Mbr_timestamp  [10]byte      
	Mbr_uuid       [16]byte        
	Mbr_partitions int32         
	Partitions     [4]PARTITION    
	Dsk_fit        [1]byte       
}

type PARTITION struct {
	Part_status     [1]byte   
	Part_type       [1]byte   
	Part_fit        [1]byte   
	Part_start      int32     
	Part_size       int32     
	Part_name       [16]byte  
	Part_uuid       [16]byte   
	Part_correlative int32     
}

type EBR struct {
	Part_mount  [1]byte    
	Part_fit    [1]byte    
	Part_start  int32      
	Part_size   int32     
	Part_next   int32      
	Part_name   [16]byte   
}