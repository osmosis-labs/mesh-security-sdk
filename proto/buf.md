# Protobufs

This is the public protocol buffers API for [Mesh Security SDK](https://github.com/osmosis-labs/mesh-security-sdk).

## Download

The `buf` CLI comes with an export command. Use `buf export -h` for details

#### Examples:

Download mesh-security protos for a commit:
```bash
## todo: not published, yet
buf export buf.build/osmosis-labs/mesh-security-sdk:${commit} --output ./tmp
```

Download all project protos:
```bash
buf export . --output ./tmp
```