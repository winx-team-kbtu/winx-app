package photo

import "io"

type StoreDTO struct {
	UserID      int64
	Email       string
	File        io.ReadSeeker
	FileName    string
	ContentType string
	SizeBytes   int64
}
