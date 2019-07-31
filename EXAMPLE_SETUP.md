# Example Setup

This is an example to setup the project on a new Ubuntu 18.04 virtual machine from scratch.

## Prerequisites

### Golang

To install the latest version of Go, run the following commands:
```
sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt-get update
sudo apt-get install golang-go
mkdir $HOME/go
export GOPATH=$HOME/go
```

To verify installation, run
```
go version
```

### Docker
Use these commands to install the latest version of Docker:
```
sudo apt-get update
sudo apt-get install \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent \
    software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"
sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io
```

To verify installation, run
```
docker --version
sudo docker run hello-world
```

We should add our user to the docker group:
```
sudo groupadd docker
sudo usermod -aG docker $USER
newgrp docker
```

Then see if we can run the hello world container without sudo:
```
docker run hello-world
```

### Docker Compose
To install Docker Compose, run these commands:
```
sudo curl -L "https://github.com/docker/compose/releases/download/1.24.1/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```
To verify installation, run
```
docker-compose --version
```

### Node.js and npm

To install Node.js and npm, run the following commands:
```
sudo apt install nodejs
sudo apt install npm
```

We can check the installed version by running
```
nodejs -v
```

### Python

Ubuntu comes with python 2.7 already installed. We can check the version by running
```
python --version
```

### Pip

```
sudo apt install python-pip
```

### Virtualenv

```
sudo apt install virtualenv
```

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
make generate ABSPATH=$GOPATH/src/github.com
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
export BEAT_PATH=$GOPATH/src/github.com/hyperledger-elastic/agent/fabricbeat
```

Next, we have to update the `_meta/beat.yml` file to contain the correct paths:
```
connectionProfile: ${GOPATH}/src/github.com/hyperledger-elastic/network/basic/connection-profile-${ORG_NUMBER}.yaml
adminCertPath: "${GOPATH}/src/github.com/hyperledger-elastic/network/basic/crypto-config/peerOrganizations/org${ORG_NUMBER}.el-network.com/users/Admin@org${ORG_NUMBER}.el-network.com/msp/signcerts/Admin@org${ORG_NUMBER}.el-network.com-cert.pem"
adminKeyPath: "${GOPATH}/src/github.com/hyperledger-elastic/network/basic/crypto-config/peerOrganizations/org${ORG_NUMBER}.el-network.com/users/Admin@org${ORG_NUMBER}.el-network.com/msp/keystore/adminKey${ORG_NUMBER}"
```
The above settings define the peer we want to connect to, and the user credentials used for initializing transactions.  
Then, we can generate the config file and docs:
```
make update
```

Afte updating, build the agent:

```
make
```

If the build is successful, we can start the agent and connect to peer0.org1.el-network.com by issuing the command
```
ORG_NUMBER=1 ORG_NAME=org1 ./fabricbeat -e -d "*"
```

Next, we can navigate to http://localhost:5601. Click the dashboards icon on the left. Kibana is taking us to select a default index pattern. Click `fabricbeat-*`, then the star in the upper right corner.
After that, we can click the dashboards and see the overview of our data on the Overview Dashboard peer0.org1.el-network.com (org1).
If the dashboards are empty, set the time range wider!
