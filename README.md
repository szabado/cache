# cache
A tool to cache command line queries.

## Speed

The tool ads a very minimal amount of latency to the first call of a command (I haven't noticed it), on subsequent calls has roughly the same performance as calling cat on a file containing the data of the request.

## Usage

```
cache: A Cache for slow shell commands.

Querying log clusters or curling API endpoints can have a latency that can
make it annoying to build up a pipe pipeline iteratively. This tool caches
those results for you so you iterate quickly.

cache runs the command for you and stores the result, and then returns the
output to you. Any data stored has a TTL of 1 hour, and subsequent calls of
the same command will return the stored result. cache will only store the
results of successful commands: if your bash command has a non-zero exit
code, then it will be uncached.

Usage:
  cache [flags] [command]

Flags:
  -c, --clear, --clean   Clear the cache.
  -o, --overwrite        Overwrite any cache entry for this command.
  -v, --verbose          Verbose logging.

Examples

  cache curl -X GET example.com
```
**Note:** `cache` only caches the first command in a sequence of pipes. If you're piping the data through slow commands, it will still be slow.

