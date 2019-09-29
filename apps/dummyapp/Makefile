
#set +x

NETWORK ?= basic
CHANNEL ?= mychannel

install:
	npm install

network:
	echo ${NETWORK}

users:
	NETWORK=${NETWORK} node enrollAdmins.js
	NETWORK=${NETWORK} node registerUsers.js

invoke:
	NETWORK=${NETWORK} CHANNEL=${CHANNEL} node invoke.js

query:
	NETWORK=${NETWORK} CHANNEL=${CHANNEL} node query.js ${KEY}

query-all:
	NETWORK=${NETWORK} node queryAll.js

clean:
	rm -rf ./wallet

