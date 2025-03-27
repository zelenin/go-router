package router

import "io/fs"

func NewFileFallbackFs(subFs fs.FS, fallbackFile string) *FileFallbackFs {
	return &FileFallbackFs{
		fs:           subFs,
		fallbackFile: fallbackFile,
	}
}

type FileFallbackFs struct {
	fs           fs.FS
	fallbackFile string
}

func (f *FileFallbackFs) Open(name string) (fs.File, error) {
	file, err := f.fs.Open(name)
	if err == nil {
		return file, nil
	}
	return f.fs.Open(f.fallbackFile)
}
