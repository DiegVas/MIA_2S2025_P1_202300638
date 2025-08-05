package structs

type SUPER_BLOCK struct {
	FileSystemType int    `json:"file_system_type"` 
	Inode_count int    `json:"inode_count"`      
	Block_count int    `json:"block_count"`      
	Free_inodes int    `json:"free_inodes"`     
	Free_blocks int    `json:"free_blocks"`     
	Mount_timestamp string `json:"system_timestamp"` 
	Desmount_timestamp string `json:"desmount_timestamp"` 
	Mount_count int    `json:"mount_count"`   
	System_magic int    `json:"system_magic"`    
	Inode_size int    `json:"inode_size"`       
	Block_size int    `json:"block_size"`      
	First_inode int    `json:"first_inode"`     
	First_block int    `json:"first_block"`      
	BFitmap_inode int    `json:"bitmap_inode"`     
	Bitmap_block int    `json:"bitmap_block"`     
	Inode_table int    `json:"inode_table"`    
	Block_table int    `json:"block_table"`     
}

type INODE struct {
	Uuid_user string `json:"uuid_user"`      
	Gid_group string `json:"gid_group"`       
	File_size int    `json:"file_size"`      
	Last_modify string `json:"modify_time"`    
	Create_time string `json:"create_time"`     
	Last_read string `json:"last_read"`       
	Inode_blocks []int `json:"inode_blocks"`    
	Inode_type byte   `json:"inode_type"`      
	UGO_permissions byte `json:"UGO_permissions"` 
}

type FOLDER_BLOCK struct {
	B_content []FOLDER_CONTENT `json:"b_content"` 
}

type FOLDER_CONTENT struct {
	Folder_name string `json:"folder_name"` 
	Folder_inode *int    `json:"folder_inode"` 
}

type FILE_BLOCK struct {
	B_content []byte `json:"b_content"`
}

type POINTER_BLOCK struct {
	B_pointers []int `json:"b_pointers"`
}
