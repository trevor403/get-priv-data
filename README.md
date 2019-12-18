# get-priv-data
Retrieve NvFBC Private Data (UUID) from a Steam installation

### Method for finding the UUID constant

The C code might look something like this
```
NvFBCCreateParams createParams;
memset(&createParams, 0, sizeof(createParams));

createParams.pPrivateData = (void*)enableKey;
createParams.dwPrivateDataSize = 16;
```

So we want to find an asignment (MOV dword ptr) to an offset (ptr) in .rdata

followed by an assignment (MOV dword ptr) of value 16 (10h)