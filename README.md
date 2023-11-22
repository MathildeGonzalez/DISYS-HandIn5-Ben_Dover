# DISYS-HandIn5-Ben_Dover

Hand-in 5, Distributed Systems

## How To start the servers

To start the servers, navigate the console to the server-folder:

```console
cd server
```

You can then start servers by writing the following in seperate consoles:

```console
go run . 0
```

```console
go run . 1
```

```console
go run . 2
```

To start 3 servers on port `5000`, `5001`, `5002`.

The program is hardcoded to start 3 servers on these ports. (In client.go, line 45)

## How To start the client(s)

To start the client, navigate the console to the client-folder:

```console
cd client
```

You can then start the client by writing:

```console
go run . name
```

For example you can write:

```console
go run . Casper
```

If you want to start more clients, you can open a new console and navigate to the client-folder again, and start a new client from there (with a unique name).

The frontend inside the client-file will automatically connect to the servers. 

## How To use the client

When the client is started, you can write commands in the console. 

The command for bidding is as follows:
```console
bid <amount>
```
where amount must be an integer.
For example you can write:
```console
bid 100
```

The first bid from a client will officially start the auction. 
The auction runs for 60 seconds. 

To query the result of the auction, you can write
```console
result
```
which will either return the current highest bid or the winner of the auction, if the auction has ended. 

## How To test the crash-handling
If you want to see how the program proceeds when a server crashes, you can try to kill one of the servers by fx closing its terminal. The program will then continue to run, and the auction will also continue using the remaining servers.
If there are multiple clients, you can also kill one of the clients, and the auction will also still continue. 