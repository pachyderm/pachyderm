package gorethink

import (
	"reflect"

	"github.com/Sirupsen/logrus"

	"gopkg.in/dancannon/gorethink.v2/encoding"
)

var (
	Log *logrus.Logger
)

const (
	SystemDatabase = "rethinkdb"

	TableConfigSystemTable   = "table_config"
	ServerConfigSystemTable  = "server_config"
	DBConfigSystemTable      = "db_config"
	ClusterConfigSystemTable = "cluster_config"
	TableStatusSystemTable   = "table_status"
	ServerStatusSystemTable  = "server_status"
	CurrentIssuesSystemTable = "current_issues"
	UsersSystemTable         = "users"
	PermissionsSystemTable   = "permissions"
	JobsSystemTable          = "jobs"
	StatsSystemTable         = "stats"
	LogsSystemTable          = "logs"
)

func init() {
	// Set encoding package
	encoding.IgnoreType(reflect.TypeOf(Term{}))

	Log = logrus.New()
}

// SetVerbose allows the driver logging level to be set. If true is passed then
// the log level is set to Debug otherwise it defaults to Info.
func SetVerbose(verbose bool) {
	if verbose {
		Log.Level = logrus.DebugLevel
		return
	}

	Log.Level = logrus.InfoLevel
}

// SetTags allows you to override the tags used when decoding or encoding
// structs. The driver will check for the tags in the same order that they were
// passed into this function. If no parameters are passed then the driver will
// default to checking for the gorethink tag.
func SetTags(tags ...string) {
	encoding.Tags = tags
}
