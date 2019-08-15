/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { FileSystemWallet, Gateway } = require('fabric-network');
const fs = require('fs');
const path = require('path');

const configPath = path.resolve(__dirname, 'config.json');
const configJSON = fs.readFileSync(configPath, 'utf8');
const config = JSON.parse(configJSON);

const ccpPath = path.resolve(__dirname, config.connection_profile);
const ccpJSON = fs.readFileSync(ccpPath, 'utf8');
const ccp = JSON.parse(ccpJSON);

async function main() {
    try {

        // Create a new file system based wallet for managing identities.
        const walletPath = path.join(process.cwd(), 'wallet');
        const wallet = new FileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        // Check to see if we've already enrolled all the users.
        for (let i = 0; i < config.users.length; i++) {
            var userExists = await wallet.exists(config.users[i].name);
            if (!userExists) {
                console.log('An identity for the user does not exist in the wallet: ', config.users[i].name);
                console.log('Run the registerUser.js application before retrying');
                return;
            }
        }
        
        for (let i = 0; i < config.transactions.length; i++) {

            let tx = config.transactions[i];
            // Create a new gateway for connecting to our peer node.
            const gateway = new Gateway();
            await gateway.connect(ccp, { wallet, identity: tx.user, discovery: { enabled: false } });

            // Get the network (channel) our contract is deployed to.
            const network = await gateway.getNetwork(config.channel.channelName);

            // Get the contract from the network.
            const contract = network.getContract(config.channel.contract);

            // Submit the transaction.
            if (tx.key){
                switch(tx.txFunction) {
                    case "addFarm":
                        await contract.submitTransaction(tx.txFunction, tx.key, tx.name, tx.state);
                        console.log(`Transaction has been submitted: ${tx.user}\t${tx.txFunction}\t${tx.key}\t${tx.name}\t${tx.state}`);
                        break;
                    case "addFactory":
                        await contract.submitTransaction(tx.txFunction, tx.key, tx.name, tx.state);
                        console.log(`Transaction has been submitted: ${tx.user}\t${tx.txFunction}\t${tx.key}\t${tx.name}\t${tx.state}`);
                        break;
                    case "addShop":
                        await contract.submitTransaction(tx.txFunction, tx.key, tx.name, tx.state);
                        console.log(`Transaction has been submitted: ${tx.user}\t${tx.txFunction}\t${tx.key}\t${tx.name}\t${tx.state}`);
                        break;
                    case "createCrate":
                        await contract.submitTransaction(tx.txFunction, tx.key, tx.farm);
                        console.log(`Transaction has been submitted: ${tx.user}\t${tx.txFunction}\t${tx.key}\t${tx.farm}`);
                        break;
                    case "createJam":
                        await contract.submitTransaction(tx.txFunction, tx.key, tx.factory, tx.crate);
                        console.log(`Transaction has been submitted: ${tx.user}\t${tx.txFunction}\t${tx.key}\t${tx.factory},\t${tx.crate}`);
                        break;
                    case "createJuice":
                        await contract.submitTransaction(tx.txFunction, tx.key, tx.factory, tx.crate);
                        console.log(`Transaction has been submitted: ${tx.user}\t${tx.txFunction}\t${tx.key}\t${tx.factory},\t${tx.crate}`);
                        break;
                    case "createSale":
                        await contract.submitTransaction(tx.txFunction, tx.key, tx.shop, tx.product);
                        console.log(`Transaction has been submitted: ${tx.user}\t${tx.txFunction}\t${tx.key}\t${tx.shop},\t${tx.product}`);
                        break;
                    case "createTransport":
                        await contract.submitTransaction(tx.txFunction, tx.key, tx.from, tx.to, tx.asset);
                        console.log(`Transaction has been submitted: ${tx.user}\t${tx.txFunction}\t${tx.key}\t${tx.from},\t${tx.to},\t${tx.asset}`);
                        break;
                    default:
                        console.log(`${tx.txFunction}: Unknown transaction name!`)
                        break;
                }
            }
            // Disconnect from the gateway.
            await gateway.disconnect();
        }
    } catch (error) {
        console.error(`Failed to submit transaction: ${error}`);
        process.exit(1);
    }
}

main();