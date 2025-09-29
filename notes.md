# Implementation notes

## General

* Switched to more Go idiomatic style where we accept interfaces and return concrete types.
* No IPLD prime - CBOR gen is adequete and significantly less complicated. The `cid.Cid` type is actually useful, despite it being a bit heavy. We _have_ to use it anyways, since it's the only thing that implements `datamodel.Link`. Also `Link` is just such a nothing interface.
* Fewer generics. Generic types in Go is not super powerful and can easily get in the way. We use generics more sparingly in this version.

## Specifics

* `DID` is now in string representation (not their binary representation as a string). You must call `Encode` and `Decode` to move to/from binary. Note, it does not have a `Bytes()` method since encoding to bytes may raise an error - you must use `Encode` instead.
* Receipt is not defined properly in the specs...
* Signatures
  * Varsig does not implement anything other than ed25519 signature and dag-cbor payload right now.
  * Signatures are now just raw bytes - no multibase prefix since signature info is all communicated in varsig header.
* Principal
    * No RSA principal implementation.
    * Signer moved from `principal/<type>/signer` to `principal/<type>` for ease of use.
    * Renamed `Encode()` method on `Signer` and `Verifier` to `Bytes()`, since it just returns the (multibase prefixed) bytes.
    * Ed25519 signer byte representation is now just the multiformats tagged private key bytes. Go internally uses 64 bytes for the private key which redundantly includes the public key.
