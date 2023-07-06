.PHONY: install, clean, dev, build, test

install: SHELL:=/bin/bash
install:
	@echo "Installing dependencies..."
	@pnpm install

dev: install
dev:
	@SK_ADAPTER=node pnpm run dev

build: install
build:
	@SK_ADAPTER=vercel pnpm run build

test: install
test:
	@pnpm run test

clean: SHELL:=/bin/bash
clean:
	@rm -rf ./node_modules ./dist ./build
