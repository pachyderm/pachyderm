run:
	kubectl $(KUBECTLFLAGS) run migrate --image="golang:1.7" --restart=Never -- sh -c 'wget https://raw.githubusercontent.com/pachyderm/pachyderm/master/migration/1.2.2-1.2.3/migrate.go && go get github.com/dancannon/gorethink && go run migrate.go'
