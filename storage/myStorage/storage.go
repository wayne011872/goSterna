package myStorage

import "io"

type Storage interface {
	Save(filePath string, file []byte) (string, error)
	SaveByReader(fp string, reader io.Reader) (string, error)
	Delete(filePath string) error
	Get(filePath string) ([]byte, error)
	FileExist(fp string) bool
	List(dir string) []string
}
