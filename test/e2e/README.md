# End-to-End Testing

### Requirements
[venom](https://github.com/ovh/venom/tree/master?tab=readme-ov-file#installing)  

### Running Tests
```bash
venom run ./test/e2e/*.yml
```

`dev-mode`: skip tests that might alter existing files and configs
```bash
venom run --var dev-mode=true ./test/e2e/*.yml
```
