## Fabricbeat Architecture

The following picture shows the basic data flow: 

![alt text](https://github.com/hyperledger-labs/blockchain-analyzer/blob/master/docs/images/Fabric%20Network%20And%20Fabricbeat.jpg "Fabric Network And Fabricbeat Basic Data Flow")

## About Elasticsearch indices

Three different Elasticsearch indices per Fabric organization are setup. One for blocks, one for transactions and one for single writes.  If multiple agents are run for peers in the same organization, they are going to send their data to the same indices. You can then select the peer on the dashboards to view its data only.  
If multiple instances are run for peers in different organizations, you will see the data of different organizations on different dashboards.  

The name of the indices can be customized in the fabricbeat configuration file (\_meta/beat.yml and `make update` or directly in fabricbeat.yml).