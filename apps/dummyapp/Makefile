
#set +x

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

