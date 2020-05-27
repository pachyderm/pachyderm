package server

import (
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
)

// NewObjClient creates an obj.Client by selecting a construcot from the obj package.
func NewObjClient(conf *serviceenv.Configuration) (obj.Client, error) {
	dir := conf.StorageRoot
	switch conf.StorageBackend {
	case MinioBackendEnvVar:
		// S3 compatible doesn't like leading slashes
		if len(dir) > 0 && dir[0] == '/' {
			dir = dir[1:]
		}
		return obj.NewMinioClientFromSecret(dir)

	case AmazonBackendEnvVar:
		// amazon doesn't like leading slashes
		if len(dir) > 0 && dir[0] == '/' {
			dir = dir[1:]
		}
		return obj.NewAmazonClientFromSecret(dir)

	case GoogleBackendEnvVar:
		// TODO figure out if google likes leading slashses
		return obj.NewGoogleClientFromSecret(dir)

	case MicrosoftBackendEnvVar:
		return obj.NewMicrosoftClientFromSecret(dir)

	case LocalBackendEnvVar:
		fallthrough

	default:
		return obj.NewLocalClient(dir)
	}
}
