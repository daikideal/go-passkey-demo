migrate:
	@echo "==> Migrating database..."

	@(cd ./migration && go run .)

	@echo "finish."