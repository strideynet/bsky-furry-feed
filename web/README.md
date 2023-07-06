# web

This directory contains the source for various web front-ends.

## Building

To build the web front-ends, you'll need to have `pnpm` installed. Then, run:

```sh
make install
```

This takes care of installing deps and setting up any necessary files.

For any given front-end, its source will be in a subdir (e.g. `./admin`).

Running `make dev-<name>` will start a livereload development server for that front-end, and `make build-<name>` will build a production bundle.

## Testing

Unit tests are handled via Vitest and can be run using `make test-<name>`.
