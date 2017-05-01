build:
	mkdir target
	GOOS=linux go build -o target/vr
	docker build -t prollynomial/vr:latest .

run:
	docker run --privileged --name vr-dev --detach prollynomial/vr:latest

clean:
	docker stop vr-dev
	docker rm vr-dev
	rm -rf target
