# Boffin

Boffin is a utility that helps collect files and file changes from multiple
sources while keeping only the most recent copy of the file, as well as keeping
destination directory structure.

You can think of it as a specialised rsync. Where rsync is THE tool to use if
you need to keep two folders in sync, Boffin tries to extend the functionality
and help you if you move, rename or change files.

## Typical Use Cases

Boffin will try to manage all files. However it works best with files that do
not change often, e.g. documents, photos, music. Boffin is not a good tool for
synchronising source code and will actually damage Git like directories, where
the whole directory is treated like a database.

The typical use case Boffin tries to solve are:

1. I want to import photos from my phone. Boffin will remember imported photos,
   so even if I delete/rename/change local copy, it will not re-import files
   that were imported before.
2. Each person in my family wants to have their own photo library, but we also
   want to share photos. I want to be able to copy photos other family members,
   organise them in my directory structure, rename and touch up in a photo
   editor. Boffin will remember which files were imported and not re-import them
   if they are locally renamed or changed. Boffin should also warn me if my
   family changes the photo on their side, so that I might want to get their
   change.
3. I edit my documents on two computers, and keep them organised differently.
   Boffin will correctly import any new documents from the other computer, it
   will update any files that were changed remotely and warn me if a file is
   changed on both computers. Boffin track files even if I rename or move them.

## Current State

Boffin has been designed to work on meta-data as much as possible. This means
that most operations are non-destructive for your files. 'import' is the only
operation that will modify your files. Even 'import' has '--dry-run' option that
will allow you to inspect all actions it would take.

In terms of it's sync algorithm, Boffin has passed my personal testing and needs
proper vetting by the community before it I can recommend it for important work.

I personally use it to manage my photo library of ~30000 photos. '--dry-run'
gives me confidence that I still have the final say before it changes my files.
From time to time I have to perform a manual operation to resolve a conflict,
but the core functionality seems stable.

## Future

The current verson of Boffin is primarily meant to prove that the algorithm
works and that doing this kind of file tracking and synchronisation is possible.
If it proves successful here are some ideas for enhancing Boffin.

- redesign 'import' to allow user to select action for for 'new', 'changed',
  'renamed' and 'deleted' files
- daemonise Boffin to allow real-time monitoring for changes
- make Boffin network aware to

## License and Warranty

Boffin is GPL v3 licensed. See LICENSE file for details. Since you will probably
use it to manage important files, you really need to be comfortable with the
license and the following section in particular:

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

## Details

Main functionality of Boffin is to maintain repository of file changes. For each
file complete history of filename, size, modification time and sha256 checksum.
This meta data is used to detect changes and to be compared to another
repository for changes and new files.


By
keeping historical checksums, we can find related files even if they have been
modified.

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





































