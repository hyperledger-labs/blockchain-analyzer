/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { FileSystemWallet, Gateway, X509WalletMixin } = require('fabric-network');
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

        // Check to see if we've already enrolled the user.
        const userExists = await wallet.exists('user1');
        if (userExists) {
            console.log('An identity for the user "user1" already exists in the wallet');
            return;
        }

        // Check to see if we've already enrolled the admin user.
        const adminExists = await wallet.exists('admin');
        if (!adminExists) {
            console.log('An identity for the admin user "admin" does not exist in the wallet');
            console.log('Run the enrollAdmin.js application before retrying');
            return;
        }

        // Create a new gateway for connecting to our peer node.
        const gateway = new Gateway();
        await gateway.connect(ccp, { wallet, identity: 'admin', discovery: { enabled: false } });

        // Get the CA client object from the gateway for interacting with the CA.
        const ca = gateway.getClient().getCertificateAuthority();
        const adminIdentity = gateway.getCurrentIdentity();

        let affiliationService = ca.newAffiliationService();

        let registeredAffiliations = await affiliationService.getAll(adminIdentity);

        for (let i = 0; i < config.users.length; i++) {
            // Register the user, enroll the user, and import the new identity into the wallet.

            if (!registeredAffiliations.result.affiliations.some(
                x => x.name == config.users[i].organization.toLowerCase())) {
                let affiliation = config.users[i].organization.toLowerCase() + '.department1';
                await affiliationService.create({
                    name: affiliation,
                    force: true
                }, adminIdentity);
            }

            const secret = await ca.register({ affiliation: config.users[i].organization.toLowerCase() + '.department1', enrollmentID: config.users[i].name, role: 'client' }, adminIdentity);
            const enrollment = await ca.enroll({ enrollmentID: config.users[i].name, enrollmentSecret: secret });
            const userIdentity = X509WalletMixin.createIdentity(config.organizations[config.users[i].organization].MSP, enrollment.certificate, enrollment.key.toBytes());
            wallet.import(config.users[i].name, userIdentity);
            console.log('Successfully registered and enrolled user and imported it into the wallet: ', config.users[i].name);
        }

    } catch (error) {
        console.error(`Failed to register user: ${error}`);
        process.exit(1);
    }
}

main();