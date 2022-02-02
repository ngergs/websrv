package filesystem

import "io/fs"

type modifiedSizeInfo struct {
	size int64
	fs.FileInfo
}

func (mod *modifiedSizeInfo) Size() int64 {
	return mod.size
}
