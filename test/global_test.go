package helmtest

import "github.com/gruntwork-io/terratest/modules/logger"

func init() {
	logger.Default = logger.Discard
}
