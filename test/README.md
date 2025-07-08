# End-to-End Testing

### Requirements
[canined](https://github.com/JackalLabs/canine-chain/releases)  
[venom](https://github.com/ovh/venom/tree/master?tab=readme-ov-file#installing)  
docker  
docker compose  

### Running Tests

> Add `--var dev-mode=true` to skip tests that alter current setup files  
> Add `-vv` flag to debug tests  

Setup canined and sequoia containers
```bash
venom run e2e/init.yaml
```
Run tests:
```bash
venom run ./test/e2e/*_test-*.yaml
```
Or all at once:
```bash
venom run e2e/init.yaml `find e2e/ -type f -name "*_test-*.yaml" | sort`
```

Stop and delete test containers
```bash
docker compose down
```

### Writing Tests

If one test depends on the other, make the file name in [sort](https://en.wikipedia.org/wiki/Sort_(Unix))ed order.
