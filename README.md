#  Vitruvius
## Extractor for the Oasis network

Vitruvius is a data extractor for Oasis. It currently powers the backend for
[Anthem](https://anthem.chorus.one), providing historical account and staking
data for accounts on the Oasis network.

## Requirements

- Go >= 1.14
- Postgres >= 10


## Getting Started

_Better instructions coming soon_.

## Contributing / Code Layout

Contributions are welcome! For a quick overview of the code structure, check
out this tree overview. It should help you navigate the project code a little
easier:

```
├── cmd                   -- The cmd directory contains the main binary.
│  └── vitruvius          
│     ├── commands        -- Subcommand code for the virtuvius binary.
│     ├── extractor       -- Background Goroutine responsible for extracting data.
│     ├── rest            -- Background Goroutine responsible for REST API to said data.
│     ├── types           -- Shared types for the project.
│     └── main.go         -- Application entry point.
├── pkg
│  └── oasis              -- Wrapper around Oasis API
│     ├── api.go          -- API Description
│     ├── grpc.go         -- gRPC Implementation of API Description
│     ├── inlet.go        -- Database batching wrapper.
│     └── types.go        -- Shared types for the library.
└── sql
   ├── migrations         -- Database schema is updated through migrations.
   ├── queries            -- All project queries can be found here.
   └── schema.sql         -- LEGACY: Schema before migrations were added.
```
