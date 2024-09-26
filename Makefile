.PHONY: build deploy clean

build:
	GOOS=linux GOARCH=amd64 go build -o ./auth-service/bootstrap ./auth-service/cmd/main.go

deploy:
	cd deploy-scripts && cdk deploy

deploy-swap:
	cd deploy-scripts && cdk deploy --hotswap

clean:
	rm -rf ./auth-service/build

