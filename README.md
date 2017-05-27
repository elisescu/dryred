## What
This is a scratchpad project used mainly to play and learn golang and a bit about the SSH protocol.
It started as an attempt to create a simple reverse proxy, that will allow the user to connect
to a service running on a machine behind NAT, by relying on a publicly accessible server, and using only the SSH identities that the user probably already had.

I had two strong requirements that I wanted to implement:

    * End-to-end secure (encrypted and authenticated). The proxy server doesn't have to be trusted by the clients, so it should act only as a publicly accessible forwarding only.
    * Very simple to use and allow select which backend clients to connect to.

In the end it proved to be impossible to piggyback on the SSH protocol to do what I wanted, and on top of that, I also found a project that was doing something similar to what I wanted, so I gave up (at least for now).