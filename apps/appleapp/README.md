# Appleapp Application
The appleapp application is used to generate users and transactions for a supply chain use-case.

### Configuration
The `config.json` contains the configuration for the application. We can configure the channel and chaincode name that we want our application to use, the users we want to enroll and the transactions we want to initialize. Transactions have 4 fields:
1. `user`: This field is required. We have to specify which user to use when making the transaction.
2. `txFunction`: This field is required. We have to specify here the chaincode function that should be called.
3. `key`: This field is required. We have to specify here the key to be written to the ledger.
4. `name`: This field is optional. We can specify here the name of the facility we create with the transaction (farm, factory or shop). Can be used with `addFarm`, `addFactory` and `addShop`.
5. `state`: This field is optional. We can specify here the state of the facility we create with the transaction (farm, factory or shop). Can be used with `addFarm`, `addFactory` and `addShop`.
6. `farm`: This field is optional. We can reference a farm by its key. Can be used with `createCrate`.
7. `from`: This field is optional. We can reference a facility (farm, factory or shop) from which the transport departs. Can be used with `createTransport`.
8. `to`: This field is optional. We can reference a facility (farm, factory or shop) to which the transport arrives. Can be used with `createTransport`.
9. `asset`: This field is optional. We can reference an asset (crate, jam or juice) that is transported. Can be used with `createTransport`.
10. `factory`: This field is optional. We can reference a factory in which the product is produced. Can be used with `createJam` and `createJuice`.
11. `crate`: This field is optional. We can reference a crate of apples of which the product is made. Can be used with `createJam` and `createJuice`.
12. `shop`: This field is optional. We can reference a shop in which the product is sold. Can be used with `createSale`.
13. `product`: This field is optional. We can reference a product that is being sold. Can be used with `createSale`.

###  User enrollment and registration
To enroll admins, register and enroll users, run the following command:
```
make users
```

###  Invoke transactions
To add key/value pairs, run
```
make invoke
```

###  Query key
To make a query, run
```
make query KEY=key1
```
