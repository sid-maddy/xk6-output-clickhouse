# xk6-output-clickhouse

k6 output [extension](https://k6.io/docs/extensions/guides/what-are-k6-extensions/) for ClickHouse.

> :warning: the API of k6 outputs [will likely change in the future](https://github.com/grafana/k6/issues/2430), so repos using it (like this repo) are not guaranteed to be working with any future version of k6.

## Build

To build a `k6` binary with this extension, first ensure you have the prerequisites:

- [Go toolchain](https://go101.org/article/go-toolchain.html)
- Git
- [xk6](https://github.com/grafana/xk6)

1. Build a `k6` binary using `xk6`:

```bash
xk6 build --with github.com/sid-maddy/xk6-output-clickhouse
```

This will result in a `k6` binary in the current directory.

2. Run a test using the built `k6` binary:

```bash
./k6 run -o clickhouse <script.js>
```
