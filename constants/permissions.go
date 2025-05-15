package constants

const (
	// FilePermission represents the default file permission (rw-r--r--)
	// This permission gives:
	// - read/write (rw-) permissions to the owner
	// - read-only (r--) permissions to the group
	// - read-only (r--) permissions to others
	DefaultFilePerms = 0644

	// DefaultDirPerms represents the default directory permission (rwxr-xr-x)
	// This permission gives:
	// - read/write/execute (rwx) permissions to the owner
	// - read/execute (r-x) permissions to the group
	// - read/execute (r-x) permissions to others
	//
	// Execute permission on directories is required to:
	// - List directory contents (ls)
	// - Access files within the directory (cd)
	// - Create/delete files in the directory
	// Without execute permission, users cannot traverse into or use the directory,
	// even if they have read/write permissions on files inside it
	DefaultDirPerms = 0755 // For directories
)
