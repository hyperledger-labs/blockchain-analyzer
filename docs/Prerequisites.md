## Prerequisites
* Go (v1.12.7+)  
* Docker  (v19.03.0+)
* docker-compose (v1.24.1+)  
* Node.js (v8.10+)  
* python (2.7) 
* virtualenv (16.7.0+)  

The instructions below have been tested on a Ubuntu 16.04 machine in AWS:

### Golang

Install [`goenv`](https://github.com/syndbg/goenv/blob/master/INSTALL.md) tool for installing the appropriate version of Go:

To install `goenv`:
```
$ git clone https://github.com/syndbg/goenv.git ~/.goenv
$ echo 'export GOENV_ROOT="$HOME/.goenv"' >> ~/.bash_profile
$ echo 'export PATH="$GOENV_ROOT/bin:$PATH"' >> ~/.bash_profile
$ echo 'eval "$(goenv init -)"' >> ~/.bash_profile
$ echo 'export PATH="$GOROOT/bin:$PATH"' >> ~/.bash_profile
$ echo 'export PATH="$GOPATH/bin:$PATH"' >> ~/.bash_profile
```

Install Go and set this version to be used globally:
```
goenv install 1.12.7
goenv global 1.12.7
```

To verify installation, run
```
go version
```

Restart the shell, create GOPATH
```
echo $GOPATH
mkdir $GOPATH -p
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

### Install [`nodenv`](https://github.com/nodenv/nodenv) and Node.js

`nodenv` is used to install an appropriate Node.js version.

```
git clone https://github.com/nodenv/nodenv.git ~/.nodenv
cd ~/.nodenv && src/configure && make -C src
# For bash only
echo 'export PATH="$HOME/.nodenv/bin:$PATH"' >> ~/.bash_profile
```

Install [`node-build`](https://github.com/nodenv/node-build) plugin:

```
mkdir -p "$(nodenv root)"/plugins
git clone https://github.com/nodenv/node-build.git "$(nodenv root)"/plugins/node-build
```

Update nodenv:
```
nodenv update
```

Install Nodejs and upgrade `npm`:
```
nodenv install 8.16.1
nodenv global 8.16.1
sudo npm install npm@latest
```

We can check the installed version by running
```
nodejs -v
npm -v
```

### Python

Install python 2.7:

```
sudo apt-get install python
python --version
```

### Virtualenv

```
sudo apt install virtualenv
```
