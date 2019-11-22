BEAT_NAME=fabricbeat
BEAT_PATH?=$(shell cd .. && echo ${PWD})
BEAT_GOPATH=$(firstword $(subst :, ,${GOPATH}))
SYSTEM_TESTS=false
TEST_ENVIRONMENT=false
ES_BEATS?=./vendor/github.com/elastic/beats
LIBBEAT_MAKEFILE=$(ES_BEATS)/libbeat/scripts/Makefile
GOPACKAGES=$(shell govendor list -no-status +local)
GOBUILD_FLAGS=-i -ldflags "-X $(BEAT_PATH)/vendor/github.com/elastic/beats/libbeat/version.buildTime=$(NOW) -X $(BEAT_PATH)/vendor/github.com/elastic/beats/libbeat/version.commit=$(COMMIT_ID)"
MAGE_IMPORT_PATH=${BEAT_PATH}/vendor/github.com/magefile/mage
NO_COLLECT=true

# Path to the libbeat Makefile
-include $(LIBBEAT_MAKEFILE)

# Initial beat setup
.PHONY: setup
setup: pre-setup git-add

pre-setup: copy-vendor git-init
	$(MAKE) -f $(LIBBEAT_MAKEFILE) mage -debug ES_BEATS=$(ES_BEATS)
	$(MAKE) -f $(LIBBEAT_MAKEFILE) update BEAT_NAME=$(BEAT_NAME) ES_BEATS=$(ES_BEATS) NO_COLLECT=$(NO_COLLECT)

# Copy beats into vendor directory
.PHONY: copy-vendor
copy-vendor:
	mkdir -p vendor/github.com/elastic
	cp -R ${BEAT_GOPATH}/src/github.com/elastic/beats vendor/github.com/elastic/
	rm -rf vendor/github.com/elastic/beats/.git vendor/github.com/elastic/beats/x-pack
	mkdir -p vendor/github.com/magefile
	cp -R ${BEAT_GOPATH}/src/github.com/elastic/beats/vendor/github.com/magefile/mage vendor/github.com/magefile

.PHONY: git-init
git-init:
	git init

.PHONY: git-add
git-add:
	git add -A
	git commit -m "Add generated fabricbeat files"

.PHONY: clean
go-get:
	go get -v github.com/hyperledger/fabric/core/ledger
	go get -v github.com/hyperledger/fabric/common/util
	go get -v github.com/hyperledger/fabric/protoutil
	# these packages are included twice
	# also present in github.com/golang.org/x/net/trace
	rm -rf ../../../hyperledger/fabric/vendor/golang.org/x/net/trace
	# also present outside vendor, need to delete vendored package to avoid type def collision
	rm -rf ../../../hyperledger/fabric/vendor/github.com/hyperledger/fabric-protos-go
	rm -rf ../../../hyperledger/fabric/vendor/github.com/golang/protobuf
	go get github.com/go-kit/kit/sd
	cd ../../../go-kit/kit/; git checkout tags/v0.8.0 -b v0.8.0
	go get github.com/cloudflare/cfssl/csr
	cd ../../../cloudflare/cfssl; git checkout tags/1.3.3 -b 1.3.3
	go get -v
	# this package is include twice. also present in github.com/golang.org/x/net/trace
	rm -rf ../../../hyperledger/fabric/vendor/golang.org/x/net/trace
go-branch-del:
	cd ../../../go-kit/kit/; git checkout master; git branch -d v0.8.0
	cd ../../../cloudflare/cfssl; git checkout master; git branch -d 1.3.3
clean::
	rm -rf build/
