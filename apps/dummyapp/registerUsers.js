/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const FabricCAServices = require('fabric-ca-client');
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

        for (let i = 0; i < config.users.length; i++) {

            // Check to see if we've already enrolled the user.
            const userExists = await wallet.exists(config.users[i].name);
            if (userExists) {
                console.log('An identity for the user already exists in the wallet: ' + config.users[i].name);
                return;
            }

            // Check to see if we've already enrolled the admin user.
            const adminExists = await wallet.exists('admin' + config.users[i].organization);
            if (!adminExists) {
                console.log('Identity for the admin user does not exist in the wallet: admin' + config.users[i].organization);
                console.log('Run the enrollAdmins.js application before retrying');
                return;
            }

            // Create a new gateway for connecting to our peer node.
            var gateway = new Gateway();
            await gateway.connect(ccp, { wallet, identity: 'admin' + config.users[i].organization, discovery: { enabled: false } });

            // Get the CA client object from the gateway for interacting with the CA.
            //var ca = gateway.getClient().getCertificateAuthority();

            var caURL = ccp.certificateAuthorities[ccp.organizations[config.users[i].organization]['certificateAuthorities'][0]].url;
            var ca = new FabricCAServices(caURL);

            var adminIdentity = gateway.getCurrentIdentity();
            let affiliationService = ca.newAffiliationService();

            let registeredAffiliations = await affiliationService.getAll(adminIdentity);

            // If the CA does not have the affiliation to which the user belongs, add it
            if (!registeredAffiliations.result.affiliations.some(
                x => x.name == config.users[i].organization.toLowerCase())) {
                let affiliation = config.users[i].organization.toLowerCase() + '.department1';
                await affiliationService.create({
                    name: affiliation,
                    force: true
                }, adminIdentity);
            }

            // Register the user, enroll the user, and import the new identity into the wallet.
            var secret = await ca.register({ affiliation: config.users[i].organization.toLowerCase() + '.department1', enrollmentID: config.users[i].name, role: 'client' }, adminIdentity);
            var enrollment = await ca.enroll({ enrollmentID: config.users[i].name, enrollmentSecret: secret });
            var userIdentity = X509WalletMixin.createIdentity(config.organizations[config.users[i].organization].MSP, enrollment.certificate, enrollment.key.toBytes());
            wallet.import(config.users[i].name, userIdentity);
            console.log('Successfully registered and enrolled user and imported it into the wallet: ', config.users[i].name);
        }

    } catch (error) {
        console.error(`Failed to register user: ${error}`);
        process.exit(1);
    }
}

main();