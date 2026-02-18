package storage

type URLResolver interface {
	ToPublicURL(objectName string) string
	ToObjectName(url string) string
}
