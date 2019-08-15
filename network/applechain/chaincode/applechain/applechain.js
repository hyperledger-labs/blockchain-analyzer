/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Contract } = require('fabric-contract-api');
const crypto = require('crypto');

class Farm {
    constructor(name, state) {
        this.name = name;
        this.state = state;
    }
}

class Crate {
    constructor(farm) {
        this.farm = farm;
    }
}

class Factory {
    constructor(name, state) {
        this.name = name;
        this.state = state;
    }
}

class Jam {
    constructor(factory, crate) {
        this.crate = crate;
        this.factory = factory;
    }
}

class Juice {
    constructor(factory, crate) {
        this.factory = factory;
        this.crate = crate;
    }
}

class Shop {
    constructor(name, state) {
        this.name = name;
        this.state = state;
    }
}

class Sale {
    constructor(shop, product) {
        this.shop = shop;
        this.product = product;
    }
}

class Transport {
    constructor(from, to, asset) {
        this.from = from;
        this.to = to;
        this.asset = asset;
    }
}

class AppleChain extends Contract {

    async addFarm(ctx, key, name, state) {
        const newFarm = new Farm(name, state);
        await ctx.stub.putState(key, Buffer.from(JSON.stringify(newFarm)));
        console.info('Added <--> ', key.toString() + ': ' + JSON.stringify(newFarm));
    }

    async addFactory(ctx, key, name, state) {
        const newFactory = new Factory(name, state);
        await ctx.stub.putState(key, Buffer.from(JSON.stringify(newFactory)));
        console.info('Added <--> ', key.toString() + ': ' + JSON.stringify(newFactory));
    }

    async addShop(ctx, key, name, state) {
        const newShop = new Shop(name, state);
        await ctx.stub.putState(key, Buffer.from(JSON.stringify(newShop)));
        console.info('Added <--> ', key.toString() + ': ' + JSON.stringify(newShop));
    }

    async createCrate(ctx, key, farm) {
        const newCrate = new Crate(farm);
        await ctx.stub.putState(key, Buffer.from(JSON.stringify(newCrate)));
        console.info('Added <--> ', key.toString() + ': ' +JSON.stringify(newCrate));
    }

    async createJam(ctx, key, factory, crate) {
        const newJam = new Jam(factory, crate);
        await ctx.stub.putState(key, Buffer.from(JSON.stringify(newJam)));
        console.info('Added <--> ', key.toString() + ': ' +JSON.stringify(newJam));
    }

    async createJuice(ctx, key, factory, crate) {
        const newJuice = new Juice(factory, crate);
        await ctx.stub.putState(key, Buffer.from(JSON.stringify(newJuice)));
        console.info('Added <--> ', key.toString() + ': ' +JSON.stringify(newJuice));
    }

    async createSale(ctx, key, shop, product) {
        const newSale = new Sale(shop, product);
        await ctx.stub.putState(key, Buffer.from(JSON.stringify(newSale)));
        console.info('Added <--> ', key.toString() + ': ' +JSON.stringify(newSale));
    }

    async createTransport(ctx, key, from, to, asset) {
        const newTransport = new Transport(from, to, asset);
        await ctx.stub.putState(key, Buffer.from(JSON.stringify(newTransport)));
        console.info('Added <--> ', key.toString() + ': ' +JSON.stringify(newTransport));
    }
}

module.exports = AppleChain;