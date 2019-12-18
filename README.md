# get-priv-data
Retrieve NvFBC Private Data (UUID) from a Steam installation

## Method for finding the UUID constant

The C code might look something like this
```
NvFBCCreateParams createParams;
memset(&createParams, 0, sizeof(createParams));

createParams.pPrivateData = (void*)enableKey;
createParams.dwPrivateDataSize = 16;
```

So we want to find an asignment (MOV dword ptr) to an offset (ptr) in .rdata

followed by an assignment (MOV dword ptr) of value 16 (10h)

## Usage

There are release precompiled for you! 
You can find them in the [Releases tab](https://github.com/trevor403/get-priv-data/releases)

You can get the executable via `go get` as well
```
go get github.com/trevor403/get-priv-data/cmd/...
```

## Disclaimer
Executing this program may put you in violation of Steam's EULA

I do not provide any legal guarantees around this software or it's usage. However it is my opinion that the Reverse Engineering effort that went into developing it is covered by the DMCA as it promotes interoperability.
