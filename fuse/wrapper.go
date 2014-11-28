package fuse

import (
	"syscall"
	"unsafe"
)

// #include "wrapper.h"
import "C"

func Version() int {
	return int(C.fuse_version())
}

//export LL_Init
func LL_Init(t unsafe.Pointer, cinfo *C.struct_fuse_conn_info) {
	ops := (*Operations)(t)
	info := &ConnInfo{}
	(*ops).Init(info)
}

//export LL_Destroy
func LL_Destroy(t unsafe.Pointer) {
	ops := (*Operations)(t)
	(*ops).Destroy()
}

//export LL_Lookup
func LL_Lookup(t unsafe.Pointer, dir C.fuse_ino_t, name *C.char,
	cent *C.struct_fuse_entry_param) C.int {

	ops := (*Operations)(t)
	err, ent := (*ops).Lookup(int64(dir), C.GoString(name))
	if err == OK {
		ent.ToC(cent)
	}
	return C.int(err)
}

//export LL_GetAttr
func LL_GetAttr(t unsafe.Pointer, ino C.fuse_ino_t, fi *C.struct_fuse_file_info,
	cattr *C.struct_stat, ctimeout *C.double) C.int {

	ops := (*Operations)(t)
	err, attr := (*ops).GetAttr(int64(ino), NewFileInfo(fi))
	if err == OK {
		ToCStat(attr.Attr, cattr)
		(*ctimeout) = C.double(attr.AttrTimeout)
	}
	return C.int(err)
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

func NewFileInfo(fi *C.struct_fuse_file_info) *FileInfo {
	return &FileInfo{
		Flags:     int(fi.flags),
		Writepage: fi.writepage != 0,
		Handle:    uint64(fi.fh),
		LockOwner: uint64(fi.lock_owner),
	}
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

	/** Inode attributes.
	 *
	 * Even if attr_timeout == 0, attr must be correct. For example,
	 * for open(), FUSE uses attr.st_size from lookup() to determine
	 * how many bytes to request. If this value is not correct,
	 * incorrect data will be returned.
	 */
	Attr *syscall.Stat_t

	/** Validity timeout (in seconds) for the attributes */
	AttrTimeout float64

	/** Validity timeout (in seconds) for the name */
	EntryTimeout float64
}

type Attr struct {
	/** Inode attributes.
	 *
	 * Even if attr_timeout == 0, attr must be correct. For example,
	 * for open(), FUSE uses attr.st_size from lookup() to determine
	 * how many bytes to request. If this value is not correct,
	 * incorrect data will be returned.
	 */
	Attr *syscall.Stat_t

	/** Validity timeout (in seconds) for the attributes */
	AttrTimeout float64
}

func ToCStat(s *syscall.Stat_t, o *C.struct_stat) {
	o.st_ino = C.__ino_t(s.Ino)
	o.st_mode = C.__mode_t(s.Mode)
	o.st_nlink = C.__nlink_t(s.Nlink)
	o.st_size = C.__off_t(s.Size)
}

func (e *EntryParam) ToC(o *C.struct_fuse_entry_param) {
	o.ino = C.fuse_ino_t(e.Ino)
	o.generation = C.ulong(e.Generation)
	ToCStat(e.Attr, &o.attr)
	o.attr_timeout = C.double(e.AttrTimeout)
	o.entry_timeout = C.double(e.EntryTimeout)
}

// Operations for Fuse's LowLevel API.
// TODO: allow implementing partial option set.
type Operations interface {
	Init(*ConnInfo)
	Destroy()
	Lookup(dir int64, name string) (err Status, entry *EntryParam)
	GetAttr(ino int64, fi *FileInfo) (err Status, attr *Attr)
}