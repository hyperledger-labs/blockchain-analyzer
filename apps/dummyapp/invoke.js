/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { FileSystemWallet, Gateway } = require('fabric-network');
const fs = require('fs');
const path = require('path');

const ccpPath = path.resolve(__dirname, '..', '..', 'network', 'connectionProfile.json');
const ccpJSON = fs.readFileSync(ccpPath, 'utf8');
const ccp = JSON.parse(ccpJSON);

const configPath = path.resolve(__dirname, 'config.json');
const configJSON = fs.readFileSync(configPath, 'utf8');
const config = JSON.parse(configJSON);

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
            // Create a new gateway for connecting to our peer node.
            const gateway = new Gateway();
            await gateway.connect(ccp, { wallet, identity: config.transactions[i].user, discovery: { enabled: false } });

            // Get the network (channel) our contract is deployed to.
            const network = await gateway.getNetwork('mychannel');

            // Get the contract from the network.
            const contract = network.getContract('dummycc');

            // Submit the transaction.
            await contract.submitTransaction('setValue', config.transactions[i].key);
            console.log('Transaction has been submitted');

            // Disconnect from the gateway.
            await gateway.disconnect();
        }
    } catch (error) {
        console.error(`Failed to submit transaction: ${error}`);
        process.exit(1);
    }
}

main();