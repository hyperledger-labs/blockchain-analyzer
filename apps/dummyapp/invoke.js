/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { FileSystemWallet, Gateway } = require('fabric-network');
const fs = require('fs');
const path = require('path');
const dotenv = require('dotenv');

const configPath = path.resolve(__dirname, 'config.json');
const configJSON = fs.readFileSync(configPath, 'utf8');
const config = JSON.parse(configJSON);

let ccpPath;
let ccpJSON;
let ccp;

async function main() {
    try {

        dotenv.config();
        if ( process.env.NETWORK != undefined) {
            config.connection_profile = config.connection_profile.replace("basic", process.env.NETWORK);
        }
        if ( process.env.CHANNEL != undefined) {
            config.channel.channelName = config.channel.channelName.replace("mychannel", process.env.CHANNEL);
        }

        ccpPath = path.resolve(__dirname, config.connection_profile);
        ccpJSON = fs.readFileSync(ccpPath, 'utf8');
        ccp = JSON.parse(ccpJSON);

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
                if (tx.previousKey) {
                    await contract.submitTransaction(tx.txFunction, tx.key, tx.previousKey);
                    console.log(`Transaction has been submitted: ${tx.user}\t${tx.txFunction}\t${tx.key}\t${tx.previousKey}`);
                }
                else {
                    await contract.submitTransaction(tx.txFunction, tx.key);
                    console.log(`Transaction has been submitted: ${tx.user}\t${tx.txFunction}\t${tx.key}`);
                }
            }
            else {
                await contract.submitTransaction(tx.txFunction);
                console.log(`Transaction has been submitted: ${tx.user}\t${tx.txFunction}`);
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