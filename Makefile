make:
	if [ -f scaffold/go.mod ]; then	rm scaffold/go.mod; fi;
	if [ -f scaffold/go.sum ]; then	rm scaffold/go.sum; fi;

	rm -rf scaffold/cmd
	rm -rf scaffold/internal
	rm -rf scaffold/bin
	rm -rf scaffold/docs

	if [ -f scaffold/docs/graph.mmd ]; then rm scaffold/docs/graph.mmd; fi;
	if [ -f scaffold/Makefile ]; then rm scaffold/Makefile; fi;
	if [ -f scaffold/Dockerfile ]; then rm scaffold/Dockerfile; fi;
	if [ -f scaffold/store/test.duckdb ]; then rm scaffold/store/test.duckdb; fi;
	if [ -f internal/application/templates/scaffold.tar.gz ]; then rm internal/application/templates/scaffold.tar.gz; fi;

	# Copy teal-ui dist files to internal/domain/services/dist
	if [ -d ../teal-ui/dist ]; then \
		rm -rf internal/domain/services/dist; \
		mkdir -p internal/domain/services/dist; \
		cp -r ../teal-ui/dist/* internal/domain/services/dist/; \
		echo "UI assets copied from teal-ui/dist"; \
	else \
		echo "Warning: ../teal-ui/dist not found, skipping UI assets copy"; \
	fi

	tar --exclude='._*' -zcvf scaffold.tar.gz -C scaffold .
	mv scaffold.tar.gz ./internal/application/templates/scaffold.tar.gz
	go build -o bin/teal ./cmd/teal

install:
	go install ./cmd/teal

test:
	go build -o bin/teal ./cmd/teal
	./bin/teal gen --project-path=scaffold

test_clean:
	go build -o bin/teal ./cmd/teal
	./bin/teal clean --project-path=scaffold --clean-main