package media

import "github.com/warmans/coldlink"

//ImageMirror wraps other library to ensure consistent image creation (same targets, max size etc.)
type ImageMirror struct {
	coldlink *coldlink.Coldlink
}

func NewImageMirror(storageDir string) *ImageMirror {
	return &ImageMirror{
		coldlink: &coldlink.Coldlink{StorageDir: storageDir, MaxOrigImageSizeInBytes: 1024 * 1024 * 1024 * 5},
	}
}

func (i *ImageMirror) Mirror(remoteURL string, localName string) (map[string]string, error) {
	return i.coldlink.Get(
		remoteURL,
		localName,
		[]*coldlink.TargetSpec{
			&coldlink.TargetSpec{Name: "orig", Op: coldlink.OpOriginal},
			&coldlink.TargetSpec{Name: "sm", Op: coldlink.OpThumb, Width: 150, Height: 150},
			&coldlink.TargetSpec{Name: "xs", Op: coldlink.OpThumb, Width: 60, Height: 60},
		},
	)
}
