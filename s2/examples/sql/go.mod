module github.com/pachyderm/s2/examples/sql

go 1.12

require (
	github.com/gofrs/uuid v3.3.0+incompatible // indirect
	github.com/gorilla/mux v1.7.4
	github.com/jinzhu/gorm v1.9.12
	github.com/pachyderm/s2 v0.0.0-20200528231500-590b33e3c716
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/sys v0.0.0-20200602225109-6fdc65e7d980 // indirect
)

replace github.com/pachyderm/s2 => ../../
