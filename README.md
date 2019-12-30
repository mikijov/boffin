# Aether

## import

- load local DB
- load remote DB
- for each file in local DB:
  - if local and remote SHA differ:
    - if remote file exists:
      - mark remote file as processed
      - if remote SHA is older version of local file and NOT local SHA older
        version of remote SHA:
        - copy remote file(SHA) to local
      - else if local SHA is older version of remote SHA and NOT remote SHA older
        version of local SHA:
        - do nothing, as this means remote is newer
        - ### here you could do two way sync
      - else:
        - handle conflict, i.e. both sides changed
    - else: # remote file does not exist
      - do nothing, as this means we have the only version
      - ### here you could do two way sync
- for each file in remote DB:
  - if not processed:
    - copy remote file(SHA) to local
