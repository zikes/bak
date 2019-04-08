# bak

A small backup utility to automatically copy changed files at specified intervals.

## Usage

*Note: On Windows the file will be bak.exe*

```sh
bak [flags]

Flags:
  -h, --help                help for bak
      --input string        the path to the file or directory to watch
      --interval duration   the interval to back up changed files (default 5m0s)
      --output string       the path to the directory where files should be backed up to
```

