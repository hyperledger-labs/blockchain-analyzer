# Example Setup With Applechain Network

This is an example to setup the project with applechain network on a new Ubuntu 18.04 virtual machine from scratch.

## Prerequisites

Please make sure that you have set up the environment for the project. Follow the steps listed in [Prerequisites](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/Prerequisites.md).   

## Cloning the repository
To get started with the project, clone the git repository. If you want to build Fabricbeat yourself, it is important that you place the project under `$GOPATH/src/github.com`. Otherwise, you can clone the repository anywhere you want (you do not need to install Go to use the pre-compiled executable or the Docker image).
```
mkdir -p $GOPATH/src/github.com
cd $GOPATH/src/github.com
git clone https://github.com/balazsprehoda/hyperledger-elastic.git
```

## Starting the applechain network
To start the applechain Fabric network, run these commands:
```
cd hyperledger-elastic/network/applechain
make start
```

## Starting the application
We use the appleapp to generate users and transactions. It digests a config file (`config.json`) and a connection profile json (`connectionProfile.json`). To get all node dependencies, run
```
cd ../../apps/appleapp
make install
```
Open `config.json`, and make sure that `channelName` is set to *applechannel* and the `connection_profile` is set to *"../../network/applechain/connectionProfile.json"*!
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

## Running the agent
We use the agent to periodically query ledger data and ship it to Elasticsearch. To run the agent on your local machine, you can build it yourself or use a pre-compiled version the same way as described in `Basic_setup` and `Multichannel_setup`. To run it inside a Docker container, you can pull the image [balazsprehoda/fabricbeat](https://hub.docker.com/r/balazsprehoda/fabricbeat), or build the image using the Dockerfile in the project root directory. For reference on using the image, please see [`blockchain-analyzer/docker-agent`](https://github.com/hyperledger-labs/blockchain-analyzer/tree/master/docker-agent).

Building locally:  
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
ORG_NUMBER=1 PEER_NUMBER=1 NETWORK=applechain ./fabricbeat -e -d "*"
```

Next, we can navigate to http://localhost:5601. Click the dashboards icon on the left. Kibana is taking us to select a default index pattern. Click `fabricbeat-*`, then the star in the upper right corner:
![alt text](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/images/Index_pattern_selection_basic.png "Setting default index pattern")
After that, we can click the dashboards and see the overview of our data on the Overview Dashboard (org1):
![alt text](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/images/Overview_apple.png "Overview")

**If the dashboards are empty, set the time range wider!**

Select Key Dashboard (org1) from the Dashboards menu! We can see all the write events that occurred to the ledger. To see only the transports that departed from `Factory0`, filter for `value.from: "Factory0"`, then click *Refresh*!
![alt text](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/images/Key_filter_for_source.png "Filter for transports from Factory0")

We can go on discovering the dashboards by scrolling and clicking the link fields, or by selecting another dashboard from the Dashboards menu.

### Graphs

One awesome feature of Elasticsearch and Kibana is the Graph representation of related documents. To create a graph, we need to upgrade our account to start a free trial:
* Click on *Management* in the menu on the left side  
* Click *License Management*  
* Click *Start Trial*

Then, we can see a new menu item called *Graph*.
![alt text](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/images/Select_graph.png "Select Graph")

Click it, then select `fabricbeat*key*org1` index pattern from the dropdown!
![alt text](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/images/Empty_graph.png "Empty graph")

Next, click the plus icon to add fields that are going to be nodes:
![alt text](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/images/Empty_graph_with_index_pattern.png "Add nodes")

In order to display every connection between the documents, we need to set certainty to 1:
![alt text](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/images/Set_certainty.png "Set certainty")

Add `creator_org`, `value.from`, `value.asset` and `value.to`. We can set custom color and icon for each field. After that, click the creator_org while pressing LShift, so that creator_org will be included in the query, but will not appear as a node (making our graph cleaner).
To search for all the transports committed by Org3, type `Org3MSP` into the Search field, then press Enter.
![alt text](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/images/Graph_for_all_transports_by_org3.png "Graph showing all transports by Org3")

After that, click "Add links between existing items":  
![alt text](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/images/Add_links_between_existing_items.png "Add links between existing items")

To add every transaction that was done by Org2, repeat the same steps for `Org2MSP`.
![alt text](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/images/Graph_for_all_transports_by_org3_and_org2.png "Graph showing all transports by Org3 and Org2")

To start more instances of the fabricbeat agent, open another tab/terminal, make sure that the GOPATH variable is set (`export GOPATH=$HOME/go`) , and run fabricbeat passing different variables from the previous run(s) (e.g.
```
ORG_NUMBER=2 PEER_NUMBER=0 NETWORK=applechain ./fabricbeat -e -d "*"
```
will start an agent querying peer0.org2.el-network.com). If the started instance queries a peer from the same organization as the previous one, we can select the peer we want to see the data of from a dropdown on the dashboards. If the new peer is shipping data from a different organization, we can see its data on a different dashboard (click the dashboards menu on the left, and choose one).
