package mock_infrastructure

import (
	"github.com/minio/minio-go/v7"
	"io"
)

type MockMinioObject struct {
	*minio.Object
	Reader io.Reader
}

func (m *MockMinioObject) Read(p []byte) (n int, err error) {
	return m.Reader.Read(p)
}

func (m *MockMinioObject) Close() error {
	return nil
}
