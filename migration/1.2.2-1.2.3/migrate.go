package main

import (
	"fmt"
	"os"

	"github.com/dancannon/gorethink"
)

type Table string

const (
	diffTable   Table = "Diffs"
	commitTable Table = "Commits"

	commitClockIndex = "CommitClockIndex"

	databaseName      = "pachyderm_pfs"
	rethinkAddressEnv = "RETHINK_PORT_28015_TCP_ADDR"
)

func main() {
	dbAddress := fmt.Sprintf("%s:28015", os.Getenv(rethinkAddressEnv))

	session, err := gorethink.Connect(gorethink.ConnectOpts{
		Address: dbAddress,
	})
	if err != nil {
		panic(err)
	}

	getTerm := func(table Table) gorethink.Term {
		return gorethink.DB(databaseName).Table(table)
	}

	if _, err = getTerm(diffTable).ForEach(func(diff gorethink.Term) gorethink.Term {
		clock := diff.Field("Clock")
		fullClock := getTerm(commitTable).GetAllByIndex(commitClockIndex, []interface{}{
			diff.Field("Repo"),
			clock.Field("Branch"),
			clock.Field("Clock"),
		}).Field("FullClock").Nth(0)
		return getTerm(diffTable).Get(diff.Field("ID")).Update(map[string]interface{}{
			"Clock": fullClock,
		}, gorethink.UpdateOpts{
			NonAtomic: true,
		})
	}).RunWrite(session); err != nil {
		panic(err)
	}

	fmt.Println("The migration has completed successfully.")
}
