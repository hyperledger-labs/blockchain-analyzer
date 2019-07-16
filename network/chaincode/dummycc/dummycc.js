/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Contract } = require('fabric-contract-api');
const crypto = require('crypto');
var counter = 1;
var lastKey;

class DummyCC extends Contract {

    async queryValue(ctx, key) {
        const dataAsBytes = await ctx.stub.getState(key);
        if (!dataAsBytes || dataAsBytes.length === 0) {
            throw new Error(`${key} does not exist`);
        }
        console.log(key.toString());
        // return dataAsBytes.toString();
        return JSON.parse(dataAsBytes)
    }

    async setValue(ctx, key, previousKey) {
        var hash = crypto.createHash('sha256').update(Buffer.from(counter.toString())).digest('hex');
        const data = {
            hash,
            previousKey
        }
        await ctx.stub.putState(key, Buffer.from(JSON.stringify(data)));
        console.info('Added <--> ', key.toString() + ': ' + JSON.stringify(data));
        counter++;
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