
#set +x

HLNETWORK?=el-network
# default path to be used in connectionprofile.json
ABSPATH?=/home/prehi/internship/testNetwork
# ABSPATH="/Users/default/code"

start:
	make generate ABSPATH=${ABSPATH}
	./start-dummy.sh
	
destroy:
	./destroy-network.sh
	rm -rf *-e
	rm -rf ../../apps/dummyapp/wallet

generate:
	cp docker-compose-TEMPLATE.yaml docker-compose.yaml
	sed -i -e "s/HLNETWORK/${HLNETWORK}/g" docker-compose.yaml
	cp connectionProfile-TEMPLATE.json connectionProfile.json
	sed -i -e "s/HLNETWORK/${HLNETWORK}/g" connectionProfile.json
	sed -i -e "s~ABSPATH~${ABSPATH}~g" connectionProfile.json
