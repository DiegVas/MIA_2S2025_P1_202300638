package structs

type SUPER_BLOCK struct {
	FileSystemType     int32     
	Inode_count        int32     
	Block_count        int32     
	Mount_timestamp    [17]byte  
	Desmount_timestamp [17]byte  
	Mount_count        int32     
	System_magic       int32     
	Inode_size         int32     
	Block_size         int32     
	First_inode        int32     
	First_block        int32     
	BFitmap_inode      int32     
	Bitmap_block       int32     
	Inode_table        int32     
	Block_table        int32     
}

type INODE struct {
	Uuid_user       int32       
	Gid_group       int32       
	File_size       int32       
	Last_modify     [17]byte    
	Create_time     [17]byte    
	Last_read       [17]byte    
	Inode_blocks    [15]int32   
	Inode_type      [1]byte     
	UGO_permissions [3]byte     
}

type FOLDER_BLOCK struct {
	B_content [4]FOLDER_CONTENT 
}
	
type FOLDER_CONTENT struct {
	Folder_name  [16]byte 
	Folder_inode int32    
}

type FILE_BLOCK struct {
	B_content [64]byte 
}

type POINTER_BLOCK struct {
	B_pointers [16]int32 
}
