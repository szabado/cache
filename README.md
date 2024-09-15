# `cache`

A tool to cache command line queries. It caches the result of a command for 1 hour, using the command and any arguments passed in as the key.

## Use Case

This tool is useful when building long pipelined chains of bash commands. For example, when drilling into the response from an API using `jq` it can take a few iterations to extract the nested sub-fields that you need. Querying the API every time will:
- Be a lot slower, making it more frustrating to poke around.
- Potentially result in getting rate limited by the API, which stops all progress in its tracks. 

Prepending the command with `cache` makes iteration fast, cheap, and easy. Getting the uncached value when your pipeline is built up is as easy as removing the `cache` prefix.

## Performance

`cache` is blazing fast. The performance impact of caching the first response is negligible, and it matches the speed of Unix tools like `cat` when reading the cached values. 

## Usage

```
Usage:
  cache [flags] [command]

Flags:
  -c, --clear, --clean   Clear the cache.
  -o, --overwrite        Overwrite any cache entry for this command.
  -v, --verbose          Verbose logging.

Examples

  cache curl -X GET example.com
```

## Future improvements
- Allow `cache` to be used in the middle of a chain of pipes. Currently it can only be used on the first command.

