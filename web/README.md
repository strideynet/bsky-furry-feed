# web

This directory contains the source for various web front-ends.

## Building

To build the web front-ends, you'll need to have `pnpm` installed. Then, run:

```sh
make install
```

This takes care of installing deps and setting up any necessary files.

Running `make dev` will start a livereload development server, and `make build` will build a production bundle using the given Adapter (defaults to `@sveltejs/adapter-node` if not specified using the `SK_ADAPTER` environment variable).

## Testing

Unit test are handled via Vitest and can be run using `make test`.
