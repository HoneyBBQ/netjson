SCHEMA_DIR := protos
GEN_DIR := gen

.PHONY: gen clean

gen:
	buf dep update
	buf generate
clean:
	rm -rf $(GEN_DIR)

