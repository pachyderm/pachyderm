When working on this, I downloaded and extracted the Trino server to this directory, so that I could hack on it and run it. It's huge, so I couldn't check it in to git, but here's how to recreate my local state (if you want to continue development):

```
# Install trino client to _trino_stuff/bin
mkdir ./bin
curl -Lo bin/trino https://repo1.maven.org/maven2/io/trino/trino-cli/406/trino-cli-406-executable.jar
chmod +x bin/trino


# Install trino server to _trino_stuff/trino-server-406
curl -LO https://repo.maven.apache.org/maven2/io/trino/trino-server/406/trino-server-406.tar.gz
tar -xvzf ./trino-server-406.tar.gz
```
