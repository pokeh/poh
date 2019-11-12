build:
	GOOS=linux go build -o poh
	zip poh.zip poh
	