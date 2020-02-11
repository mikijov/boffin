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

{
  event: "deleted|moved|changed",
  timestamp: "change timestamp",
  path: "new path",
  checksum: "new checksum"
}


## to-do

- import new files into staging directory; imported files should be in the same
  directory structure as the source
- allow separate staging directories for each source client

