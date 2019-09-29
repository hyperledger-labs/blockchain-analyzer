/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const FabricCAServices = require('fabric-ca-client');
const { FileSystemWallet, X509WalletMixin } = require('fabric-network');
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

        ccpPath = path.resolve(__dirname, config.connection_profile);
        ccpJSON = fs.readFileSync(ccpPath, 'utf8');
        ccp = JSON.parse(ccpJSON);

        // Create a new file system based wallet for managing identities.
        const walletPath = path.join(process.cwd(), 'wallet');
        const wallet = new FileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        // For each organization in the config file, get the CA from the connection profile.
        const orgs = Object.keys(config.organizations);
        for (let i = 0; i < orgs.length; i++) {
            // Check if organization is present in connection profile and if it contains at least one CA.
            if (ccp.organizations[orgs[i]].certificateAuthorities && ccp.organizations[orgs[i]].certificateAuthorities.length != 0) {
                // Create a new CA client for interacting with the CA.
                const caURL = ccp.certificateAuthorities[ccp.organizations[orgs[i]].certificateAuthorities[0]].url;
                const ca = new FabricCAServices(caURL);

                // Check to see if we've already enrolled the admin user for this CA.
                const adminExists = await wallet.exists('admin' + orgs[i]);
                if (adminExists) {
                    console.log('An identity for the admin user already exists in the wallet: admin', orgs[i]);
                    return;
                }
                // Enroll the admin user, and import the new identity into the wallet.
                const enrollment = await ca.enroll({ enrollmentID: 'admin', enrollmentSecret: 'adminpw' });
                const identity = X509WalletMixin.createIdentity(config.organizations[orgs[i]].MSP, enrollment.certificate, enrollment.key.toBytes());
                wallet.import('admin' + orgs[i], identity);
                console.log('Successfully enrolled admin user and imported it into the wallet: admin' + orgs[i]);
            }
        }
    } catch (error) {
        console.error(`Failed to enroll admin user "admin": ${error}`);
        process.exit(1);
    }
}

main();