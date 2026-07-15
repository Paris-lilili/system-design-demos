## What this verifies
A file is split into fixed size chunks, each stored under it Sha256 hash. 
A manifest records the ordered chunk list(metadata), separately from the chunk data

This shows:
1. Reuploading the same file will not write new chunks
2. The file can be rebuilt from manifest and chunks, verified by size and sha-256
3. Memory keeps constant, both chunking and rebuilding stream one chunk at a time
4. Chunks are large and immutable, accessing by hash. But manifest store in relational database

## How to Run
```
❯ go run main.go
Successfully wrote 1048676 random bytes to random.bin file.
New chunks number: 17, reused chunks number: 0, total chunks number: 17, they should be same
two files are same, total bytes: 1048676

❯ cat random.bin.manifest.json
{
  "file_name": "random.bin",
  "size": 1048676,
  "chunks": [
    "f7058de6ac949bf17415fbf56a57edde0d5bdbf9661e4478366888d238b5f29e",
    "5fbe821cd33072a64dfe5bd09e6893d9be442879d8a04981a4f6892e383031d6",
    "14940dde5c6ee1bbb1efa3dfa8b5440882f432447e15f38f016a6a4eff6b08b1",
    "7389e2946695091f56f0820e0b2054c54bd73ba63a4d324bcb8fbe967e623b5b",
    "296a7a90bcef80f2f855a7523953fa2f11e2bb01181ed776ef3d8133e44d3817",
    "2dbfcc97454073bc58fa10369314846820aa24af2525f0a822af1cb6980caa9b",
    "8bb3540552cd2b72554272c9dcb891112329d9c9809b78764654abb10cc7c38c",
    "62e485f056da76faa0d2b1e6c4637788ea1b7fe6bca36451a3b6c1b3ccf6d34a",
    "2aa78a45be161eb2cea7c9321f65fb8c717092282d09e732615c0c273d1c10dd",
    "06d4bc79f2db80ce8edc78ac6f8456ea66cdbf2f84bdd35c7d5d43de72d62df9",
    "60653279371b782ec2464871dec2d054b85df6ea631e7ec819ee58a3008649f8",
    "427a48aed2e89f5df499536f5283038df77399f866984f5426ea32821c443be5",
    "de8261dd9cbef762128df248cae57c7f5f8c61ce8bafcd221c55601f8e188977",
    "213a95ed0babdf9b775489abcb30cf2c96cb2d439af5b5db517115e0ff089bf4",
    "dad5e99a0c82c9c000cf4fefa2ab58f5615fb907c58832ef3c8f8cccdbbe8b8e",
    "e82101b64fb9f90c7c930cacf89fae25807f1e71ca8d6d056cedb88785cb72f7",
    "056fcb7da6c2c9775b813ef6a87ec920bf4206c5921054744f4dd600f1eb57b6"
  ]
}

# run again, all chunks are reused
❯ go run main.go
New chunks number: 0, reused chunks number: 17, total chunks number: 17, they should be same
two files are same, total bytes: 1048676


❯ tree
.
├── chunks
│   ├── 056fcb7da6c2c9775b813ef6a87ec920bf4206c5921054744f4dd600f1eb57b6
│   ├── 06d4bc79f2db80ce8edc78ac6f8456ea66cdbf2f84bdd35c7d5d43de72d62df9
│   ├── 14940dde5c6ee1bbb1efa3dfa8b5440882f432447e15f38f016a6a4eff6b08b1
│   ├── 213a95ed0babdf9b775489abcb30cf2c96cb2d439af5b5db517115e0ff089bf4
│   ├── 296a7a90bcef80f2f855a7523953fa2f11e2bb01181ed776ef3d8133e44d3817
│   ├── 2aa78a45be161eb2cea7c9321f65fb8c717092282d09e732615c0c273d1c10dd
│   ├── 2dbfcc97454073bc58fa10369314846820aa24af2525f0a822af1cb6980caa9b
│   ├── 427a48aed2e89f5df499536f5283038df77399f866984f5426ea32821c443be5
│   ├── 5fbe821cd33072a64dfe5bd09e6893d9be442879d8a04981a4f6892e383031d6
│   ├── 60653279371b782ec2464871dec2d054b85df6ea631e7ec819ee58a3008649f8
│   ├── 62e485f056da76faa0d2b1e6c4637788ea1b7fe6bca36451a3b6c1b3ccf6d34a
│   ├── 7389e2946695091f56f0820e0b2054c54bd73ba63a4d324bcb8fbe967e623b5b
│   ├── 8bb3540552cd2b72554272c9dcb891112329d9c9809b78764654abb10cc7c38c
│   ├── dad5e99a0c82c9c000cf4fefa2ab58f5615fb907c58832ef3c8f8cccdbbe8b8e
│   ├── de8261dd9cbef762128df248cae57c7f5f8c61ce8bafcd221c55601f8e188977
│   ├── e82101b64fb9f90c7c930cacf89fae25807f1e71ca8d6d056cedb88785cb72f7
│   └── f7058de6ac949bf17415fbf56a57edde0d5bdbf9661e4478366888d238b5f29e
├── go.mod
├── main.go
├── random.bin
├── random.bin.manifest.json
├── README.md
└── rebuiltFile.bin
```