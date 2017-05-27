* DRServer listens on port 8080

* *DRBackClient* `raspberrypi_1` connects to *DRServer*. Both sides are authenticated using ssh keys
* *DRBackClient* `raspberrypi_2` connects to *DRServer*. Both sides are authenticated using ssh keys
* *DRClient* `client_1` connects to *DRServer*. Both sides authenticated using ssh keys
* *DRClient* asks the server to open a connection to `raspberrypi_1` via a protocol command. The
server opens the connection and `client_1` and `raspberrypi_1` have now a virtual end-to-end
connection to be used.

All of this is via TCP+SSH connection:  `DRClient <--SSH--> DRServer <--SSH--> DRBackClient`

A new end-to-end SSH session is established between the thw clients `DRClient<==SSH==>DRBackClient`.

The server will use the underlying TCP connection 

Example:

```
                           +----------+                 +-------------+
                           |          | <----SSH------+ |raspberrypi_1|
                           |          |                 +-------------+
+-------------+            | DRServer |
|  client_1   | ---SSH---> |          |
+-------------+            |          |                 +-------------+
                           |          | <----SSH------+ |raspberrypi_2|
                           +----------+                 +-------------+

```