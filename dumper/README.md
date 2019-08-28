# Dumper

This is a simple go program that queries a specified peer for ledger data, and dumps that data into `json` files. It runs similarly to `fabricbeat`, the main difference is that it does not use Elasticsearch. The main purpose of this program is to experiment with analysis based on various databases.

This program uses the modules `fabricbeatsetup`, `fabricutils` and `ledgerutils` of the fabricbeat agent.

## Custom persistence
The program uses `Persistent` interface for persistence, which means we can define our custom persistence methods for any databases. All we have to do is to implement the `Persistent` interface, create an instance of our implementation and replace `DefaultConfig` with our own instance.
