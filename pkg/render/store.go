package render

//go:generate mockgen -destination ./mocks/file_store_mock.go github.com/solo-io/valet/pkg/render FileStore

type FileStore interface {
	Load(path string) (string, error)
}
