# Boffin

##

- [x] boffin init - initialize new repo
- [x] boffin update - update repository with local changes
- [x] boffin update --force - check actual content of the files even if size/date are unchanged
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
- [x] boffin import - import new and changed files from the remote repo
- [ ] boffin import --delete - delete local files that have been deleted remotely
- [ ] allow separate staging directories for each source client

### diff algo

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

### update algo

# get list of files that should be checked
- for each file on the file system
  - if the path matches exactly
    - if forced check or if size/date changed
      - put file in to-be-checked list
    - else
      - mark checked
  - else
    - put file in to-be-checked list

# bulk calculate checksums
- for each file in the to-be-checked list
  - calculate checksum
  - put checksum in the 'updates' map[hash][]result

# match everything you can using hashes
- for each update
  - get current files with matching hashes

# prioritize matching of files with same meta data
  - for each current file in matching hashes
    - for each result in update
      - if whole path for existing file and result matches
        - mark as unchanged
        - mark existing file as checked
        - remove existing file from matching hashes
        - remove result from update

# prioritize matching of files with same name
  - for each current file in matching hashes
    - for each result in update
      - if filename for existing file and result matches
        - mark as moved
        - mark existing file as checked
        - remove existing file from matching hashes
        - remove result from update

# simple case of file being renamed
  - if only one file in matching hashes and one result in updates
    - mark as moved
    - mark existing file as checked
    - remove existing file from matching hashes
    - remove result from update

  - if only updates remain
    - for every result in updates
      - mark as new
      - remove result from update
  <!-- - else if only matches remain -->
  <!--   - for every current file in matching hashes -->
  <!--     - mark as deleted -->
  - else
    - for every result in updates
      - mark as conflict
      - remove result from update
    - for every current file in matching hashes
      - mark as conflict

# match everything you can using paths
- for each update
  - if existing file exists with the same path
    - mark as changed
    - mark as checked
  - else
    - mark as new

# any file not checked means it was deleted
- for each existing file
  - if not checked
    - mark as deleted
    - mark as checked





































