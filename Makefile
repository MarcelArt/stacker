build:
	@go build -o stacker .

install: build
	@mkdir -p $(HOME)/.local/bin
	@cp stacker $(HOME)/.local/bin/stacker
