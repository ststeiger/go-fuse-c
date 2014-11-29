package fuse

import (
	"time"
)

// Raw operations for Fuse's LowLevel API.
// TODO: allow implementing partial option set.
type RawFileSystem interface {
	Init(*ConnInfo)
	Destroy()
  // StatFs

	Lookup(dir int64, name string) (entry *EntryParam, err Status)
  // Forget

	GetAttr(ino int64, fi *FileInfo) (attr *InoAttr, err Status)
  // SetAttr

  // Directory handling
	ReadDir(ino int64, fi *FileInfo, off int64, size int, w DirEntryWriter) Status
  // OpenDir
  // ReleaseDir
  // FsyncDir

  // File handling
	Open(ino int64, fi *FileInfo) Status
  Read(p []byte, ino int64, off int64, fi *FileInfo) (n int, err Status)
  // Create

  // TODO: extended attribute handling
}

type DirEntryWriter interface {
	Add(name string, ino int64, mode int, next int64) bool
}

type FileInfo struct {
	Flags     int
	Writepage bool
	// Bitfields not supported by CGO.
	// TODO: create separate wrapper?
	//DirectIo     bool
	//KeepCache    bool
	//Flush        bool
	//NonSeekable  bool
	//FlockRelease bool
	Handle    uint64
	LockOwner uint64
}

func (f *FileInfo) AccessMode() AccessMode {
	return AccessMode(f.Flags & 3)
}

type ConnInfo struct {
	// TODO
}

type EntryParam struct {
	/** Unique inode number
	 *
	 * In lookup, zero means negative entry (from version 2.5)
	 * Returning ENOENT also means negative entry, but by setting zero
	 * ino the kernel may cache negative entries for entry_timeout
	 * seconds.
	 */
	Ino int64

	/** Generation number for this entry.
	 *
	 * If the file system will be exported over NFS, the
	 * ino/generation pairs need to be unique over the file
	 * system's lifetime (rather than just the mount time). So if
	 * the file system reuses an inode after it has been deleted,
	 * it must assign a new, previously unused generation number
	 * to the inode at the same time.
	 *
	 * The generation must be non-zero, otherwise FUSE will treat
	 * it as an error.
	 *
	 */
	Generation int64

	/**
	 * Inode attributes.
	 */
	Attr *InoAttr

	/** Validity timeout (in seconds) for the attributes */
	AttrTimeout float64

	/** Validity timeout (in seconds) for the name */
	EntryTimeout float64
}

/** Inode attributes.
 *
 * Even if Timeout == 0, attr must be correct. For example,
 * for open(), FUSE uses attr.Size from lookup() to determine
 * how many bytes to request. If this value is not correct,
 * incorrect data will be returned.
 */
type InoAttr struct {
	Ino   int64
	Size  int64
	Mode  int
	Nlink int

	Atim time.Time
	Ctim time.Time
	Mtim time.Time

	/** Validity timeout (in seconds) for the attributes */
	Timeout float64
}