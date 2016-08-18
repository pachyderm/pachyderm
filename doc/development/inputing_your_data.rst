Getting Your Data into Pachyderm
===========================
Data lives in object storage and pachyderm needs control of its own bucket because conent addressed... add more info



Input Mechanisms
----------------

## Inputting data into PFS

To start using Pachyderm, you'll need to input some of your data into a [PFS](./pachyderm_file_system.html) repo.

There are a handful of ways to do this:

1) You can do this via a [PFS](./pachyderm_file_system.html) mount

The [Fruit Stand](https://github.com/pachyderm/pachyderm/tree/master/examples/fruit_stand) example uses this method. This is probably the easiest to use if you're poking around and make some dummy repo's to get the hang of Pachyderm.

2) You can do it via the [pachctl](./pachctl.html) CLI

This is probably the most reasonable if you're scripting the input process.

3) You can use the golang client

You can find the docs for the client [here](https://godoc.org/github.com/pachyderm/pachyderm/src/client)

Again, helpful if you're scripting something, and helpful if you're already using go.


Pachctl Mount
^^^^^^^^^^^^^

Pachctl CLI
^^^^^^^^^^^

Golang Client
^^^^^^^^^^^^^

Other Clients
^^^^^^^^^^^^^


Data Sources
------------


From Local Files
^^^^^^^^^^^^^^^^

From your Object Storage
^^^^^^^^^^^^^^^^^^^^^^^^

From a Database
^^^^^^^^^^^^^^^

One-time dump

Ongoing sync with a DB. 