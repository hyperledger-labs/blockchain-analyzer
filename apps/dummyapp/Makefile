
#set +x

HLNETWORK?=el-network
# default path to be used in connectionprofile.json
ABSPATH?=${GOPATH}/src/github.com
# ABSPATH="/Users/default/code"

install:
	npm install

users:
	node enrollAdmins.js
	node registerUsers.js

invoke:
	node invoke.js

query:
	node query.js ${KEY}

query-all:
	node queryAll.js 

clean:
	rm -rf ./wallet

