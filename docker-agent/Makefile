
#set +x

HLNETWORK?=el-network
ORG_NUMBER?=1
NETWORK?=basic
PEER_NUMBER?=0

WHICHOS := $(shell uname)

start:
	make generate
	docker-compose up -d
	
destroy:
	docker-compose down
	rm -rf *-e
	rm -rf ./docker-compose*.yaml ./connection-profile*.yaml

generate:
	cp templates/docker-compose-TEMPLATE.yaml docker-compose.yaml
	sed -i -e "s/CONFIG_ORG_NUM/${ORG_NUMBER}/g" docker-compose.yaml
	sed -i -e "s/CONFIG_NETWORK_NAME/${NETWORK}/g" docker-compose.yaml
	sed -i -e "s/CONFIG_PEER_NUM/${PEER_NUMBER}/g" docker-compose.yaml
	for NUMBER in 1 2 3 4 ; \
	do \
		cp templates/network/${NETWORK}/connection-profile-TEMPLATE.yaml connection-profile-$${NUMBER}.yaml; \
		sed -i -e  "s/HLNETWORK/${HLNETWORK}/g" connection-profile-$${NUMBER}.yaml; \
		sed -i -e  "s/ORG/Org$${NUMBER}/g" connection-profile-$${NUMBER}.yaml; \
	done
ifeq ($(WHICHOS),Darwin)
	rm *-e
endif
