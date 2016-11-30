# Easily convert PEM files to Java Keystore

## Installation

    go get github.com/jimmidyson/pemtokeystore/cmd/pemtokeystore

## Usage
    Usage of pemtokeystore:
      -ca-cert-file path
            PEM-encoded CA certificate file path(s) - repeat for multiple files
      -cert-file alias=path
            PEM-encoded certificate file(s) in the format alias=path - repeat for multiple files
      -key-file alias=path
            PEM-encoded private key file(s) in the format alias=path - repeat for multiple files
      -keystore path
            path to keystore
      -keystore-password password
            keystore password

## Example
```bash
$ pemtokeystore -keystore my.ks -keystore-password changeit \
                -ca-cert-file ca-root.pem -ca-cert-file ca-signer.pem \
                -cert-file myserver=server.pem -key-file myserver=key.pem
```