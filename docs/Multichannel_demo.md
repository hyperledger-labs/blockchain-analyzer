## Complex Demo

This project provides a simple way of installing and trying out the main features, just to see its purpose. First, make sure that `GOPATH` is set correctly. Then, in the top directory of the project (`hyperledger-elastic`), run `make multichannel` to build and run components with a more complex configuration (for the topology of the network, see [Multichannel Network](https://github.com/balazsprehoda/hyperledger-elastic/tree/master#multichannel-network)!).

Next, we can navigate to http://localhost:5601.
Click the dashboards icon on the left:
![alt text](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/images/Starting_page.png "Kibana starting page")
Kibana is taking us to select a default index pattern. Click `fabricbeat-*`, then the star in the upper right corner:
![alt text](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/images/Index_pattern_selection_multi.png "Setting default index pattern")
After that, we can click the dashboards again to see the dashboards:
![alt text](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/images/Dashboards_multi.png "Dashboards")
See the overview of our data on the Overview Dashboard (org1). We can select peer and channel in the two dropdown lists:
![alt text](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/images/Org1_overview_multi.png "Org1 overview")
We can see that Org1 is member of both `twochannel` and `fourchannel` channels, and there are two fabricbeat instances shipping data from two peers (`peer0.org1.el-network.com` and `peer1.org1.el-network.com`) in Org1.  
To see the data of Org2, select Overview Dashboard (org2) in the Dashboard Menu:
![alt text](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/images/Org2_overview_multi.png "Org2 overview")
**If the dashboards are empty, set the time range wider!**

You can stop the fabricbeat agent with `Ctrl+C`, and bring down the whole network and remove generated data by issuing `make destroy-multichannel`.