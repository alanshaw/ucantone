# Implementation notes

* `DID` is now in string representation (not their binary representation as a string). You must call `Encode` and `Decode` to move to/from binary. Note, they do not have a `Bytes()` method since encoding to bytes may raise an error - you must use `Encode` instead.
* No IPLD prime - CBOR gen is adequete and significantly less complicated. The `cid.Cid` type is actually useful, despite it being a bit heavy. We _have_ to use it anyways, since it's the only thing that implements `datamodel.Link`. Also `Link` is just such a nothing interface.
* Switched to more Go idiomatic style where we accept interfaces and return concrete types.
* Varsig does not implement anything other than ed25519 and dag-cbor right now.
