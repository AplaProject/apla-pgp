# apla-pgp

## Example of apla-pgp.conf

```
LogFile = "apla.log"
StoreFile = "./store/apla-pgp.store"
OutPath = "./backup"

[Settings]
  Timeout = 2
  Compression = 1
  NodePrivateKey = "/home/user/apla/apla-data/NodePrivateKey"

[PGP]
  Path = "/home/user/.gnupg"
  Phrase = "1234"

[TCP]
  Host = "127.0.0.1:7078"
```
