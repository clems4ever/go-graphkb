# go-graphkb

go-graphkb is a Graph-oriented Knowledge Base written in Go.

The Knowledge Base can be queried using [openCypher](https://github.com/opencypher/openCypher)
and results can be visualized in the UI as shown below.

![go-graphkb ui](./docs/images/go-graphkb.png)


## Getting started

Run the following commands

    # Spin up GraphKB in few seconds with (wait 15 seconds for mariadb to start).
    source bootstrap.sh && docker-compose up -d

    # Insert the example data available in examples/ directory
    # with the following command:
    go run cmd/importer-csv/main.go --config cmd/importer-csv/config.yml

Then visit the web UI accessible at http://127.0.0.1:3000.


## LICENSE

**go-graphkb** is licensed under Apache 2.0.
