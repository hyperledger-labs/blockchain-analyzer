# Example Setup With Basic Network

This is an example to setup the project with basic network on a new Ubuntu 18.04 virtual machine from scratch.

## Prerequisites

Please make sure that you have set up the environment for the project. Follow the steps listed in [Prerequisites](https://github.com/balazsprehoda/hyperledger-elastic/blob/test-build-path/docs/Prerequisites.md).   

## Cloning the repository
To get started with the project, we have to clone the repository first. It is important that we put it under `$GOPATH/src/github.com`. 
```
mkdir -p $GOPATH/src/github.com
cd $GOPATH/src/github.com
git clone https://github.com/balazsprehoda/hyperledger-elastic.git
```

## Starting the basic network
To start the basic Fabric network, run these commands:
```
cd hyperledger-elastic/network/basic
make start
```

## Starting the application
We use the dummyapp to generate users and transactions. It digests a config file (`config.json`) and a connection profile json (`connectionProfile.json`). To get all node dependencies, run
```
cd ../../apps/dummyapp
make install
```
Open `config.json`, and make sure that `channelName` is set to *mychannel* and the `connection_profile` is set to *"../../network/basic/connectionProfile.json"*!
After that, we can generate the users:
```
make users
```
When it is done, start invoking the transactions:
```
make invoke
```

## Starting Elastic stack
This project provides an Elasticsearch cluster with 3 containers behind nginx, and a Kibana container. We use Elasticsearch and Kibana to index and view the ledger data. To start the Elastic stack, run
```
cd ../../stack
make start
```

After 1-2 minutes, navigate to http://localhost:5601 in a browser to check whether Kibana started successfully or not.
If it displays "Kibana server is not ready yet", check the logs of elasticsearch1 container:
```
docker logs elasticsearch1
```
If there is an error saying  
*"ERROR: [1] bootstrap checks failed
[1]: max virtual memory areas vm.max_map_count [65530] is too low, increase to at least [262144]"*  
issue the command  
```
sudo sysctl -w vm.max_map_count=262144
```
This sets the vm.max_map_count kernel setting to 262144.
After this, we can stop the stack by typing
```
make destroy
```
and then restart it with
```
make start
```

If you see *"Cannot connect to the Elasticsearch cluster"*, try refreshing the page, or opening the page in another tab.

## Building and running the agent
We use the agent to periodically query ledger data and ship it to Elasticsearch. To get started with the agent, run
```
cd ../agent/fabricbeat
export PATH=$PATH:$GOPATH/bin
make update
```

Afte updating, build the agent:

```
make
```

If the build is successful, we can start the agent and connect to peer0.org1.el-network.com by issuing the command
```
ORG_NUMBER=1 PEER_NUMBER=0 NETWORK=basic ./fabricbeat -e -d "*"
```

Next, we can navigate to http://localhost:5601. Click the dashboards icon on the left. Kibana is taking us to select a default index pattern. Click `fabricbeat-*`, then the star in the upper right corner.
After that, we can click the dashboards and see the overview of our data on the Overview Dashboard (org1).
If the dashboards are empty, set the time range wider!
