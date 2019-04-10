package pretty

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/transaction"
	"github.com/pachyderm/pachyderm/src/server/pkg/pretty"
)

const (
	//TransactionHeader is the header for transactions.
	TransactionHeader = "TRANSACTION\tSTARTED\tLENGTH\t\n"
)

type PrintableTransactionInfo struct {
	*transaction.TransactionInfo
	FullTimestamps bool
}

func PrintTransactionInfo(w io.Writer, info *transaction.TransactionInfo, fullTimestamps bool) {
	fmt.Fprintf(w, "%s\t", info.Transaction.ID)
	if fullTimestamps {
		fmt.Fprintf(w, "%s\t", info.Started.String())
	} else {
		fmt.Fprintf(w, "%s\t", pretty.Ago(info.Started))
	}
	fmt.Fprintf(w, "%d\n", len(info.Requests))
}

func PrintDetailedTransactionInfo(info *PrintableTransactionInfo) error {
	template, err := template.New("TransactionInfo").Funcs(funcMap).Parse(
		`ID: {{.Transaction.ID}}{{if .FullTimestamps}}
Started: {{.Started}}{{else}}
Started: {{prettyAgo .Started}}{{end}}
Requests:
{{transactionRequests .Requests}}
`)
	if err != nil {
		return err
	}
	return template.Execute(os.Stdout, info)
}

func transactionRequests(requests []*transaction.TransactionRequest) string {
	if len(requests) == 0 {
		return "  -"
	}

	lines := []string{}
	for i, _ := range requests {
		lines = append(lines, fmt.Sprintf("  request %d", i))
	}

	return strings.Join(lines, "\n")
}

var funcMap = template.FuncMap{
	"prettyAgo":           pretty.Ago,
	"prettySize":          pretty.Size,
	"transactionRequests": transactionRequests,
}
