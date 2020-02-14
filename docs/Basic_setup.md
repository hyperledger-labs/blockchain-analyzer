# Getting Started with Basic Network

This is an example to setup the project with `basic` network on a new Ubuntu 18.04 / 16.04 virtual machine from scratch. The instructions should also work on Mac OS.

1. [Install Prequisites](#install-prerequisites)
2. [Clone the repository](#clone-the-repository)
3. [Start / stop a Hyperledger Fabric network using `basic` configuration](#start--stop-the-basic-network). See `multichannel` page for multi-channel configuration.
4. [Create users and transactions using dummy application](#create-users-and-transactions-using-dummy-application)
5. [Start Elastic stack](#start-elastic-stack)
6. [Fabricbeat agent](#fabricbeat-agent)
7. [Configuring Indices for the first time in Kibana](#configuring-indices-for-the-first-time-in-Kibana)
8. [Viewing dashboards](#viewing-dashboards) that store data

## Install Prerequisites

Please make sure that you have set up the environment for the project. Follow the steps listed in [Prerequisites](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/Prerequisites.md).   

## Clone the repository
To get started with the project, clone the git repository. If you want to build Fabricbeat yourself, it is important that you place the project under `$GOPATH/src/github.com`. Otherwise, you can clone the repository anywhere you want (you do not need to install Go to use the pre-compiled executable or the Docker image). 
```
$ mkdir $GOPATH/src/github.com -p
$ cd $GOPATH/src/github.com  
$ git clone https://github.com/hyperledger-labs/blockchain-analyzer.git
```

## Start / stop the `basic` network
We will use the `basic` Hyperpedger Fabric network. It is a simple test 
network with four organizations, one peer per organization, a solo orderer 
communicating over TLS and a sample chaincode called `dummycc`. It writes 
deterministically generated hashes and (optionally) previous keys as value 
to the ledger.  


Issue the following command in the `network/basic` directory
```
make start
```

To enter the Fabric CLI Docker container, issue the command:
```
docker exec -it cli bash  
```

Inside CLI, the `/scripts` folder contains scripts that can be used to install, instantiate and invoke chaincode (though the `make start` command takes care of installation and instantiation).

### Stop the network

To stop the network and delete all the generated data (crypto material, channel artifacts 
and dummyapp wallet), run:

```
make destroy
```

### Remove chaincode images

If you make changes to `dummycc` chaincode, you will need to remove the old images:

```
make rmchaincode
```

## Create users and transactions using dummy application

The dummyapp application is used to create users and generate transactions
for different scenarios so that the resulting transactions can be 
analyzed with the Elastic stack. Examples of scenarios include which 
channels to use, and which Fabric ca users to invoke transactions.

This application can connect to both the basic and the multichannel networks.  

The commands in this section should be issued from the `blockchain-analyzer/apps/dummyapp` directory.

### Install dummy application dependencies
Before the first run, we have to install the necessary node modules:
```
npm install
```

Take a look at the `config.json` file. This file contains the transactions that will be invoked into the `basic` network.


###  User enrollment and registration
To enroll admins, register and enroll users, run the following command:
```
make users
```

###  Invoke transactions
To add key-value pairs, run:
```
make invoke
```

### Query
To query a specific key, run
```
make query KEY=Key1
```
To query all key-value pairs, run
```
make query-all
```



## Start Elastic stack

This project includes an Elasticsearch and Kibana setup to index and 
visualize blockchain data.  
The commands in this section should be issued from the `blockchain-analyzer/stack` folder.

If you are working in a machine with low memory, the Elasticsearch container may not start. In this case, issue the following command:
```
sudo sysctl -w vm.max_map_count=262144
```
to set the vm.max_map_count kernel setting to 262144, then destroy and bring up the Elastic Stack again.

### Credit
This setup is borrowed from
https://github.com/maxyermayank/docker-compose-elasticsearch-kibana

### Description
Kibana container and Elasticsearch cluster with nginx and 3 Elasticsearch containers. To view Kibana in browser, navigate to http://localhost:5601

### Start
To start the containers, navigate to `blockchain-analyzer/stack` directory and issue:
```
make start
```

### Stop and destroy
To stop the containers, issue
```
make destroy
```
To stop the containers and remove old data, run
```
make erase
```


## Fabricbeat Agent

The fabricbeat Beats agent is responsible for connecting to a specified peer, periodically querying its ledger, processing the data and shipping it to Elasticsearch. Multiple instances can be run at the same time, each querying a different peer and sending its data to the Elasticsearch cluster.  

### Docker image

Fabricbeat agent is also available as a Docker image ([balazsprehoda/fabricbeat](https://hub.docker.com/r/balazsprehoda/fabricbeat)). You can use this image, or build it using the command  
```
$ docker build -t <IMAGE NAME> .
```  
from the project root directory.

To start the agent, you have to mount two configuration files and the necessary crypto materials:

- `fabricbeat.yml`: configuration file for the agent (see `blockchain-analyzer/agent/fabricbeat/fabricbeat.yml` for reference)
- connection profile yaml file referenced from `fabricbeat.yml`
- crypto materials referenced from the connection profile and `fabricbeat.yml`

If you use environment variables in the configuration file, do not forget to set these variables in the container!

For a sample Docker setup, see [`/blockchain-analyzer/docker-agent/`](https://github.com/hyperledger-labs/blockchain-analyzer/tree/master/docker-agent).

### Running the agent locally

The commands in this section should be issued from the `blockchain-analyzer/agent/fabricbeat` directory.

You can build the agent yourself, or you can use a pre-built one from the `blockchain-analyzer/agent/fabricbeat/prebuilt` directory. To use an executable from the `prebuilt` dir, choose the appropriate for your system and copy it into the `blockchain-analyzer/agent/fabricbeat` folder.

#### Environment setup

Before configuring and building the fabricbeat agent, please make sure that the `GOPATH` variable is set correctly. Then, add `$GOPATH/bin` to the `PATH`:
```
export PATH=$PATH:$GOPATH/bin
```

Ensure that Python version is 2.7.*. 

Get module dependencies:
```
make go-get
```

Build the agent:
```
make update
make
```


#### Start fabricbeat locally

To start the agent, issue the following command from the `fabricbeat` directory:
```
./fabricbeat -e -d "*"
```
To use the agent with the `basic` network from the `blockchain-analyzer/network` folder, you can start the agent using:
```
ORG_NUMBER=1 PEER_NUMBER=0 NETWORK=basic ./fabricbeat -e -d "*"
```
The variables passed are used in the configuration (`fabricbeat.yml`). To connect to another network or peer, change the configuration (and/or the passed variables) accordingly.

#### Stop fabricbeat

To stop the agent, simply type `Ctrl+C`


### Configuring Indices for the first time in Kibana

Next, we navigate to http://localhost:5601.

Click the dashboards icon on the left:
![alt text](https://github.com/hyperledger-labs/blockchain-analyzer/blob/master/docs/images/Starting_page.png "Kibana starting page")


Kibana takes us to select a default index pattern. Click `fabricbeat-*`, then the star in the top right corner:
![alt text](https://github.com/hyperledger-labs/blockchain-analyzer/blob/master/docs/images/Index_pattern_selection_basic.png "Setting default index pattern")


### About indices

Three different Elasticsearch indices per Fabric organization are setup. One for blocks, one for transactions and one for single writes.  If multiple agents are run for peers in the same organization, they are going to send their data to the same indices. You can then select the peer on the dashboards to view its data only.  
If multiple instances are run for peers in different organizations, you will see the data of different organizations on different dashboards.  

The name of the indices can be customized in the fabricbeat configuration file (\_meta/beat.yml and `make update` or directly in fabricbeat.yml).


## Viewing dashboards

After that, we can click the dashboards and see the overview of our data on the Overview Dashboard (org1):
![alt text](https://github.com/hyperledger-labs/blockchain-analyzer/blob/master/docs/images/Org1_overview_basic.png "Org1 overview")
**If the dashboards are empty, set the time range wider!**

We can go on discovering the dashboards by scrolling and clicking the link fields, or by selecting another dashboards from the Dashboards menu.
