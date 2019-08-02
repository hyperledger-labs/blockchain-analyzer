## Prerequisites
* Go (v1.12.7+)  
* Docker  (v19.03.0+)
* docker-compose (v1.24.1+)  
* Node.js (v8.10+)  
* python (2.7)
* pip (v9.0.1+)  
* virtualenv (16.7.0+)  

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
npm -v
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
