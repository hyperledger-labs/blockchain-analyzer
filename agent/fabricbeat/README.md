# Fabricbeat

Welcome to Fabricbeat.

Ensure that this folder is at the following location:
`${GOPATH}/src/github.com/hyperledger-labs/blockchain-analyzer/fabricbeat`
Or run:
```
export BEAT_PATH=$PWD
export GOPATH=$GOPATH:$BEAT_PATH
```
The above script sets the BEAT_PATH variable to the current directory, and adds it to the GOPATH.

## Getting Started with Fabricbeat

### Requirements

* [Golang](https://golang.org/dl/) 1.7

### Init Project
To get running with Fabricbeat and also install the
dependencies, run the following command:

```
make setup
```

It will create a clean git history for each major step. Note that you can always rewrite the history if you wish before pushing your changes.

To push Fabricbeat in the git repository, run the following commands:

```
git remote set-url origin https://github.com/hyperledger-labs/blockchain-analyzer/fabricbeat
git push origin master
```

For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).

### Build

To build the binary for Fabricbeat run the command below. This will generate a binary
in the same directory with the name fabricbeat.

```
make
```


### Run

To run Fabricbeat with debugging output enabled, run:

```
./fabricbeat -c fabricbeat.yml -e -d "*"
```


### Test

To test Fabricbeat, run the following command:

```
make testsuite
```

alternatively:
```
make unit-tests
make system-tests
make integration-tests
make coverage-report
```

The test coverage is reported in the folder `./build/coverage/`

### Update

Each beat has a template for the mapping in elasticsearch and a documentation for the fields
which is automatically generated based on `fields.yml` by running the following command.

```
make update
```


### Cleanup

To clean  Fabricbeat source code, run the following command:

```
make fmt
```

To clean up the build directory and generated artifacts, run:

```
make clean
```


### Clone

To clone Fabricbeat from the git repository, run the following commands:

```
mkdir -p ${GOPATH}/src/github.com/hyperledger-labs/blockchain-analyzer/fabricbeat
git clone https://github.com/hyperledger-labs/blockchain-analyzer/fabricbeat ${GOPATH}/src/github.com/hyperledger-labs/blockchain-analyzer/fabricbeat
```


For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).


## Packaging

The beat frameworks provides tools to crosscompile and package your beat for different platforms. This requires [docker](https://www.docker.com/) and vendoring as described above. To build packages of your beat, run the following command:

```
make release
```

This will fetch and create all images required for the build process. The whole process to finish can take several minutes.

## Comparison and Testing with Hyperledger Explorer

The ledger can be inspected from both Kibana and Hyperledger Explorer, which makes it easy to compare the two, and make sure that the visualized data in Kibana is correct.

### Hyperledger Explorer Configuration

Configuring HL Explorer: https://github.com/hyperledger/blockchain-explorer  
To connect HL Explorer to `basic` network, you should make the following changes:

1. Change `blockchain-explorer/app/explorerconfig.json` to this:

```
{
	"persistence": "postgreSQL",
	"platforms": ["fabric"],
	"postgreSQL": {
		"host": "127.0.0.1",
		"port": "5432",
		"database": "fabricexplorer",
		"username": "hppoc",
		"passwd": "password"
	},
	"sync": {
		"type": "local",
		"platform": "fabric",
		"blocksSyncTime": "1"
	},
	"jwt": {
		"secret": "a secret phrase!!",
		"expiresIn": "2h"
	}
}
```

2. Change `blockchain-explorer/app/platform/fabric/config.json` to look like this:

```
{
	"network-configs": {
		"elastic-network": {
			"name": "elasticnetwork",
			"profile": "project-location/blockchain-analyzer/network/basic/connectionProfile.json",
			"enableAuthentication": false
		}
	},
	"license": "Apache-2.0"
}
```

...where `project-location` is the absolute path to the `blockchain-analyzer` project location on disk.

3. Also make sure that you have set up the database correctly! (Follow [these](https://github.com/hyperledger/blockchain-explorer#50-database-setup----) instructions.)
