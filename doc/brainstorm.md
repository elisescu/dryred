#### Terms

* `back_client` (BC) - the "exit" component that will do the forwarding with the destination IP/port
* `front_client` (FC) - the user side client that needs the FW
* `server` (S) - the publicly available server, that stays between the FC and BC
* `target_address` - the address to which the `back_client` will forward the connection coming from
the `front_client`
* `fw_connection` - a "virtual" connection between the FC and the target address
* `fw_session` - something that starts and ends when the user wants, and is identifying a 
communication between the FC and BC

#### Events

* `back_client` loses connection to the `fw_server`
* user wants a new forwarding channel, to a new `target_address`
* user connects and opens a `fw_connection`

When everything is setup, then me (the user/developer) on the client device I wanna do:

Example:

`back_clients` : po.elisescu.com, pi.elisescu.com, ubuntu.elisescu.com

```
dr list  # lists all forward routes, with their status and info
dr ssh <name>   # connects with ssh via the <name> fw route
dr ssh_config <name> >> ~/.ssh/config # add ssh hosts to config
dr start <name|--all>  # starts a FW connection
dr add-fw <route|name=localport:remote_host:remote_port>  # adds new fw tunnel
dr add-rv <route|name=listen_localport:host:hostport

dr ssh po.ssh  # hostname/device=po, service:ssh 
```

```
# dr_config
[po]
auth_token=fba889c0ffee..5bcdf88d

[po.ssh]
forwarder=tcp:8000:localhost:22

[po.router]
destination_address=tcp:192.168.0.1:80
```

#### Commands
```
dr-client --server elisescu.com:9000
dr-server
```

`ssh elisescu:pi@dryred.io`

/bin/ssh --SSH--> [go:ssh_server] dryred.io [go:ssh_client]  <--SSH--  [go:ssh_server] back_client
