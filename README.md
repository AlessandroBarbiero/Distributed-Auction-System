# Distributed-Auction-System
Use replication to design an auction system service that is resilient to crashes in Golang.

## Run network
In order to easily run all the server nodes at ones we wrote a shell and a batch file.
The scripts read ports from config file `client/DNS_Cache.info` and runs servers on those ports.
Clients have to be run manually by typing:

`go run client/client.go`

For debug purposes a name can be added at the end of the command, but it will be only used to show it in the log file.

### Unix

```bash
./run.sh
```
Each server is spawned as new terminal emulator. Emulator name has to be saved in variable `$TERM`  
To set required terminal emulator run `TERM=${your terminal emulator}` before running `./run.sh`

### Windows

```bash
run.bat
```

Each server is spawned as new cmd prompt.


## Config file

Ports for all servers have to be specified in `client/DNS_Cache.info` file.
It represents a sort of DNS cache for the clients, because the client will search there the ports they need to reach
in order to find the servers.
The same file is also used to spawn the different servers.

Each line of the file represents a single node and the only thing on that line should be the port of that node.

## Use the network

After running the script you will see `n` (n being the number of nodes in your client/DNS_Cache.info file) terminals.
In each terminal you will see the message showing witch port the node is running on.
Once started the clients you will be able to use them via two commands:
- s
- b [number]

The first one will show the current auction status, while the second can be used to place a bid.

## Script tested on

- `5.18.0-1parrot1-amd64` (basically debian)
- `Windows 10`