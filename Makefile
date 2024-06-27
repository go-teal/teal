make:
	if [ -f scaffold/go.mod ]; then	rm scaffold/go.mod; fi;
	if [ -f scaffold/go.sum ]; then	rm scaffold/go.sum; fi;	

	rm -rf scaffold/cmd
	rm -rf scaffold/internal

	if [ -f scaffold/docs/graph.wsd ]; then rm scaffold/docs/graph.wsd; fi;
	if [ -f internal/application/templates/scaffold.tar.gz ]; then rm internal/application/templates/scaffold.tar.gz; fi;

	tar --exclude='._*' -zcvf scaffold.tar.gz -C scaffold .
	mv scaffold.tar.gz ./internal/application/templates/scaffold.tar.gz
	go build -o bin/teal ./cmd/teal

install:
	go install ./cmd/teal

test: 
	go build -o teal.out ./cmd/teal
	./teal.out gen --project-path=scaffold  

test_clean:
	go build -o teal.out ./cmd/teal
	./teal.out clean --project-path=scaffold --clean-main