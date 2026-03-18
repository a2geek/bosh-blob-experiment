# BOSH blobstore experiment

The thought experiment: What happens if your blobstore gets whacked, or somehow hosed, or you don't have access, etc?
Can it be rebuilt?

```bash
$ bbx --help
NAME:
   bbx - bosh blob experiment

USAGE:
   bbx [global options] [command [command options]]

COMMANDS:
   report, blobs-report           Report on blob status
   mklocal, make-blobstore-local  Make this blobstore local
   import, import-release-blobs   Import blobs from a release
   help, h                        Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --project string, -p string  project directory (default: .) [$BBX_PROJECT]
   --verbose, -v                enable logging (default: false)
   --debug                      dump stacktrace for any error raised (default: false)
   --help, -h                   show help
```

## Blobs report

List every blob and optionally verify the SHA checksum.

```bash
$ bbx report --help
NAME:
   bbx report - Report on blob status

USAGE:
   bbx report [options]

DESCRIPTION:
   Generate a report of expected blobs for each final release version

OPTIONS:
   --version string, -v string  version regex (default: .*)
   --verify, -V                 Verify the SHA against the file. WARNING: This may take a bit when blobstore is remote.
   --all, -a                    Include all blobs found in the blobstore.
   --help, -h                   show help

GLOBAL OPTIONS:
   --project string, -p string  project directory (default: .) [$BBX_PROJECT]
```

## Make blobstore local

Make the blobstore local (aka switching to Git LFS) and configures for using `<project>/final_blobs` directory. Renames existing `final.yml` and `private.yml` files. Optionally copies files out of old blobstore into new blobstore.

```bash
$ bbx mklocal --help
NAME:
   bbx mklocal - Make this blobstore local

USAGE:
   bbx mklocal [options]

DESCRIPTION:
   Convert this blobstore to be a local blobstore and download all blobs

OPTIONS:
   --directory string, -d string  subdirectory for blob storage (default: "final_blobs")
   --copy, -c                     copy blobs from existing blobstore (default: false)
   --help, -h                     show help

GLOBAL OPTIONS:
   --project string, -p string  project directory (default: .) [$BBX_PROJECT]
```

## Import blobs

Imports blobs from an existing release. Useful if blobs are missing.

```bash
$ bbx import --help
NAME:
   bbx import - Import blobs from a release

USAGE:
   bbx import [options]

DESCRIPTION:
   Fetch a release (file or URL allowed) and import any blobs that aren't present

OPTIONS:
   --help, -h  show help

GLOBAL OPTIONS:
   --project string, -p string  project directory (default: .) [$BBX_PROJECT]
```

# References

* [BOSH releases with Git LFS](https://web.archive.org/web/20230926152942/https://www.starkandwayne.com/blog/bosh-releases-with-git-lfs/)
