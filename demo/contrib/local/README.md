# Dev scripts
For manual testing. Works on my box(*) ...


*) OSX

```
make install
cd contrib/local
rm -rf /tmp/trash
HOME=/tmp/trash bash setup_meshd.sh
HOME=/tmp/trash bash start_node.sh
```

Next shell:

```
cd contrib/local
./01-gov.sh
```

## Shell script development

[Use `shellcheck`](https://www.shellcheck.net/) to avoid common mistakes in shell scripts.
[Use `shfmt`](https://github.com/mvdan/sh) to ensure a consistent code formatting.
