/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Contract } = require('fabric-contract-api');
const md5 = require('md5');
var counter = 1;

class DummyCC extends Contract {

    async initLedger(ctx) {

        console.info('============= START : Initialize Ledger ===========');

        for (var i = 1; i <= 10; i++) {
            await ctx.stub.putState('Key' + i, Buffer.from(md5(i)));
            console.info('Added <--> ', 'Key' + i + ': ' + md5(i));
        }
        counter = i;
        console.info('============= END : Initialize Ledger ===========');
    }

    async queryValue(ctx, key) {
        const dataAsBytes = await ctx.stub.getState(key);
        if (!dataAsBytes || dataAsBytes.length === 0) {
            throw new Error(`${key} does not exist`);
        }
        console.log(key.toString());
        return dataAsBytes.toString();
    }

    async setValue(ctx, key) {
        console.info('============= START : Set Value ===========');
        await ctx.stub.putState(key, Buffer.from(md5(counter++)));
        console.info('Added <--> ', key.toString() + ': ' + md5(counter-1));
        console.info('============= END : Set Value ===========');
    }

    async queryAllValues(ctx) {
        const startKey = 'Key0';
        const endKey = 'Key999';

        const iterator = await ctx.stub.getStateByRange(startKey, endKey);

        const allResults = [];
        while (true) {
            const res = await iterator.next();

            if (res.value && res.value.value.toString()) {
                console.log(res.value.value.toString('utf8'));

                const Key = res.value.key;
                let Record;
                try {
                    Record = JSON.parse(res.value.value.toString('utf8'));
                } catch (err) {
                    console.log(err);
                    Record = res.value.value.toString('utf8');
                }
                allResults.push({ Key, Record });
            }
            if (res.done) {
                console.log('end of data');
                await iterator.close();
                console.info(allResults);
                return JSON.stringify(allResults);
            }
        }
    }
}

module.exports = DummyCC;