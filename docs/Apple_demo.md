## Apple Demo

This project provides a simple way of installing and trying out the main features, just to see its purpose. First, make sure that `GOPATH` is set correctly. Then, in the top directory of the project (`hyperledger-elastic`), run `make apple` to build and run components with a simple configuration imitating a supply chain use-case (for the topology of the network, see [Apple Network](https://github.com/balazsprehoda/hyperledger-elastic/tree/master#apple-network)!).

Next, we can navigate to http://localhost:5601.
Click the dashboards icon on the left:
![alt text](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/images/Starting_page.png "Kibana starting page")
Kibana is taking us to select a default index pattern. Click `fabricbeat-*`, then the star in the upper right corner:
![alt text](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/images/Index_pattern_selection_basic.png "Setting default index pattern")
After that, we can click the dashboards again to see the dashboards:
![alt text](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/images/Dashboards_basic.png "Dashboards")
By clicking the dropdown lists, we can see that Org1 is member of `applechannel`, and there is one fabricbeat instance shipping data from one peer (`peer0.org1.el-network.com`) in Org1.
**If the dashboards are empty, set the time range wider!**

We can see all the writes that occurred to the ledger on the Key Dashboard. To see all the transports that departed from `Factory0`, filter for the term `value.from: "Factory0"`.

We can use graph visualizations to see the related documents. For further instructions, please refer to [Apple setup example graphs section](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/Apple_setup_example.md#graphs)

You can stop the fabricbeat agent with `Ctrl+C`, and bring down the whole network and remove generated data by issuing `make destroy-apple`.