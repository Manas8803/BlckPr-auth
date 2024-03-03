.PHONY: build deploy clean

build:
	GOOS=linux GOARCH=amd64 go build -o ./auth-service/build/main ./auth-service/cmd/main.go
	GOOS=linux GOARCH=amd64 go build -o ./email-service/build/main ./email-service/main.go
	zip -r ./deploy-scripts/auth-service.zip ./auth-service
	zip -r ./deploy-scripts/email-service.zip ./email-service

deploy:
	cd deploy-scripts && cdk deploy

deploy-swap:
	cd deploy-scripts && cdk deploy --hotswap

clean:
	rm -rf ./auth-service/build
	rm -rf ./email-service/build
	rm -f ./deploy-scripts/auth-service.zip
	rm -f ./deploy-scripts/email-service.zip
