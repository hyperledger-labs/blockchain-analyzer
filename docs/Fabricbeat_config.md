## Fabric beats configuration

### Configure fabricbeat

The agent is configured using the `fabricbeat.yml` file. This file is generated 
based on the `fabricbeat/_meta/beat.yml` file. If you want to update the generated 
config file, please edit `fabricbeat/_meta/beat.yml`, and then run:
```
make update
```

The configurable fields are the following:
* `period`: defines how often an event is sent to the output (Elasticsearch in this case)
* `organization`: defines which organization the connected peer is part of
* `peer`: defines the peer which fabricbeat should query (must be defined in the connection profile)
* `connectionProfile`: defines the location of the connection profile of the Fabric network
* `adminCertPath`: absolute path to the admin certfile
* `adminKeyPath`: absolute path to the admin keyfile
* `elasticURL`: URL of Elasticsearch (defaults to http://localhost:9200)
* `kibanaURL`: URL of Kibana (defaults to http://localhost:5601)
* `blockIndexName`: defines the name of the index to which the block data should be sent
* `transactionIndexName`: defines the name of the index to which the transaction data should be sent
* `keyIndexName`: defines the name of the index to which the key write data should be sent
* `dashboardDirectory`: folder which should contain the generated dashboards
* `templateDirectory`: folder which contains the templates for Kibana objects (index patterns, dashboards, etc.)
* `chaincodes`: describes the chaincodes installed on the peer
  * `name`: the name of the chaincode
  * `values`: the keys of the values that get persisted with the key (e.g. fabcar: key: CAR0 values: [make, model, colour, owner])
  * `linkingKey`: the name of the key that links transactions (e.g. dummycc: previousKey)
* `setup.ilm.enabled`: setting this false makes possible to define our own indices (for blocks, transactions and keys per organization)
* `output.elasticsearch.index`: the template for runtime index creation
* `output.elasticsearch.hosts`: the list of elasticsearch hosts we want our agent to connect to
* `setup.template.name`: the name of the index template that is going to be automatically created if does not exist
* `setup.template.pattern`: the index template is loaded for indices matching this pattern
* `setup.dashboards.directory`: the directory that contains the (generated) dashboards to be imported into Kibana on start of the agent