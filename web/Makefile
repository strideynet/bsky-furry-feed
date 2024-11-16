.PHONY: install, clean, dev, build, test

install: SHELL:=/bin/bash
install:
	@echo "Installing dependencies..."
	@pnpm install

install-%: SHELL:=/bin/bash
install-%:
	@echo "Installing dependencies for $*..."
	@pnpm --filter $* install

dev-%: install-%
	@pnpm --filter $* dev

build-%: install-%
	@SK_ADAPTER=cloudflare pnpm --filter $* build

test-%: install-%
	@pnpm --filter $* test

clean: SHELL:=/bin/bash
clean:
	@rm -rf ./node_modules ./admin/dist ./admin/build
