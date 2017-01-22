# gorssget

Simple configurable rss fetcher written in go.

## Build
Install dependencies via:
```
go get -v "github.com/mmcdole/gofeed"
go get -v "gopkg.in/yaml.v2"
```

Clone the repo and build with go:
```
go build -ldflags "-s -w"
```

## Sample configuration
Uses yaml for configuration:

```
tasks:
  a task:
    rss: http://example.com/rss.xml
    cookies: "mycookiename=value"
    download: /path/subpath
    quality: 1080p
    shows:
      - Some Name
```
