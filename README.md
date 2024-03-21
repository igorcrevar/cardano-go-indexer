# Cardano Tx Indexer
The Cardano Indexer is a Go library built using the gouroboros library, available at [blinklabs-io/gouroboros](https://github.com/blinklabs-io/gouroboros).

## Key Features
- **Address Specification**: Users can specify addresses of interest that may appear in both inputs or outputs of transactions. This allows for targeted monitoring of specific addresses.  
- **Configurable Block Confirmation**: The indexer supports a configurable number of children blocks after a block is considered final. This flexibility enables users to adjust confirmation criteria based on their requirements.
- **Restart Resilience**: In the event of a restart, the indexer resumes from the latest confirmed point (block), ensuring continuity and consistency in data indexing.
