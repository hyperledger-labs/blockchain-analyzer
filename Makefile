WHICHOS := $(shell uname)
GOPATH?=$(shell dirname $(shell dirname $(shell dirname ${PWD})))

install:
	go get github.com/hyperledger/fabric-sdk-go
	cd apps/dummyapp && make install
	cd apps/appleapp && make install

basic:
	go get github.com/hyperledger/fabric-sdk-go
	cd stack && make start
	cd network/basic && make start
	cd apps/dummyapp &&	sed -i -e "s/\"channelName\": \"fourchannel\"/\"channelName\": \"mychannel\"/g" config.json &&	sed -i -e "s/\"connection_profile\": \"..\/..\/network\/multichannel\/connectionProfile.json\"/\"connection_profile\": \"..\/..\/network\/basic\/connectionProfile.json\"/g" config.json
	make remove-intermediate
	cd apps/dummyapp && make install && make users && make invoke
	cd agent/fabricbeat/_meta && rm -rf beat.yml && cp templates/basic-config.yml beat.yml
	export PATH=${PATH}:${GOPATH}/bin && cd agent/fabricbeat && make update && make
	#Waiting for Kibana
	sleep 15
	cd agent/fabricbeat && ORG_NUMBER=1 PEER_NUMBER=0 ./fabricbeat -e -d "*"

apple:
	go get github.com/hyperledger/fabric-sdk-go
	cd stack && make start
	cd network/applechain && make start
	cd apps/appleapp && make install && make users && make invoke
	cd agent/fabricbeat/_meta && rm -rf beat.yml && cp templates/apple-config.yml beat.yml
	export PATH=${PATH}:${GOPATH}/bin && cd agent/fabricbeat && make update && make
	#Waiting for Kibana
	sleep 15
	cd agent/fabricbeat && ORG_NUMBER=1 PEER_NUMBER=0 ./fabricbeat -e -d "*"

multichannel:
	go get github.com/hyperledger/fabric-sdk-go
	cd stack && make start
	cd network/multichannel && make start
	cd apps/dummyapp &&	sed -i -e "s/\"channelName\": \"mychannel\"/\"channelName\": \"fourchannel\"/g" config.json &&	sed -i -e "s/\"connection_profile\": \"..\/..\/network\/basic\/connectionProfile.json\"/\"connection_profile\": \"..\/..\/network\/multichannel\/connectionProfile.json\"/g" config.json
	make remove-intermediate
	cd apps/dummyapp && make install && make users && make invoke
	cd agent/fabricbeat/_meta && rm -rf beat.yml && cp templates/multichannel-config.yml beat.yml
	export PATH=${PATH}:${GOPATH}/bin && cd agent/fabricbeat && make update && make
	#Waiting for Kibana
	sleep 15
	cd agent/fabricbeat && ORG_NUMBER=1 PEER_NUMBER=0 ./fabricbeat -e -d "*" &
	cd agent/fabricbeat && ORG_NUMBER=1 PEER_NUMBER=1 ./fabricbeat -e -d "*" &
	cd agent/fabricbeat && ORG_NUMBER=2 PEER_NUMBER=0 ./fabricbeat -e -d "*"

destroy-basic:
	cd stack && make erase
	cd network/basic && make destroy
	cd agent/fabricbeat/_meta && rm -rf beat.yml && cp templates/default-config.yml beat.yml

destroy-apple:
	cd stack && make erase
	cd network/applechain && make destroy
	cd agent/fabricbeat/_meta && rm -rf beat.yml && cp templates/default-config.yml beat.yml

destroy-multichannel:
	cd stack && make erase
	cd network/multichannel && make destroy
	cd agent/fabricbeat/_meta && rm -rf beat.yml && cp templates/default-config.yml beat.yml

remove-intermediate:
ifeq ($(WHICHOS),Darwin)
	cd apps/dummyapp && rm *-e
endif