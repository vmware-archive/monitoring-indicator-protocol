Install `go-bindata` with:

```bash
go get -u github.com/go-bindata/go-bindata/...
```

Use 

```
go-bindata -o pkg/asset/schema.go -pkg asset schemas.yml
``` 

from the repo root to regenerate `schema.go`.
