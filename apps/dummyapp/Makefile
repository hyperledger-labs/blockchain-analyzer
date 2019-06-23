
#set +x

HLNETWORK?=el-network
# default path to be used in connectionprofile.json
ABSPATH?=/home/prehi/internship/testNetwork
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

clean:
	rm -rf ./wallet

