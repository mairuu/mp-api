package storage

import (
	"fmt"
	"strings"
)

type BaseURLResolver struct {
	baseURL string
}

func NewBaseURLResolver(baseURL string) *BaseURLResolver {
	return &BaseURLResolver{baseURL: baseURL}
}

func (r *BaseURLResolver) ToPublicURL(objectName string) string {
	return fmt.Sprintf("%s/%s", r.baseURL, objectName)
}

func (r *BaseURLResolver) ToObjectName(url string) string {
	if len(url) > len(r.baseURL) && url[:len(r.baseURL)] == r.baseURL {
		url = url[len(r.baseURL):]
	}
	return strings.TrimPrefix(url, "/")
}
