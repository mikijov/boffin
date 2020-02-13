# Boffin

##

- [x] boffin init - initialize new repo
- [x] boffin update - update repository with local changes
- [ ] boffin update --force - check actual content of the files even if size/date are unchanged
- [x] boffin diff - display differences between two repositories
      - result
        - left
          - path
          - moved
          - changed
          - deleted
        - right
          - path
          - moved
          - changed
          - deleted
- [ ] boffin import - import new and changed files from the remote repo
- [ ] boffin import --stage - import files into a staging directory rather in-place
- [ ] boffin import --delete - delete local files that have been deleted remotely

- for each file in the remote repo
  - if not deleted and checksum exists in local repo
    - mark local file as checked
    - if match is current file version in local repo
      - 'Equal'
    - else - if match is older version
      - if local file is deleted
        - 'LocalDeleted'
      - else
        - 'LocalChanged'
  - else
    - for each checksum in file history
      - if checksum exists in local repo
        - mark local file as checked
        - if match is current file version in local repo
          - if file is deleted
            - 'RemoteDeleted'
          - else
            - 'RemoteChanged'
        - else if both files deleted
          - 'Equal'
        - else - if match is older version and they are not both deleted
          - 'Conflict'
    - else - checksum not matched
      - if file is deleted
        - 'RemoteDeleted'
      - else
        - 'RemoteAdded'

- for each file in the local repo
  - if not checked
    - if deleted
      - 'LocalDeleted'
    - else
      - 'LocalAdded'


{
  event: "deleted|moved|changed",
  timestamp: "change timestamp",
  path: "new path",
  checksum: "new checksum"
}


## to-do

- allow separate staging directories for each source client

