# Kratos Project Template

## Install Kratos
```
go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
```
## Create a service
```
# Create a template project
kratos new ragdesk

cd ragdesk
# Add a proto template
kratos proto add api/ragdesk/ragdesk.proto
# Generate the proto code
kratos proto client api/ragdesk/ragdesk.proto
# Generate the source code of service by proto file
kratos proto server api/ragdesk/ragdesk.proto -t internal/service

go generate ./...
go build -o ./bin/ ./...
./bin/ragdesk -conf ./configs
```
## Generate other auxiliary files by Makefile
```
# Download and update dependencies
make init
# Generate API files (include: pb.go, http, grpc, validate, swagger) by proto file
make api
# Generate all files
make all
```
## Automated Initialization (wire)
```
# install wire
go get github.com/google/wire/cmd/wire

# generate wire
cd cmd/ragdesk
wire
```

## Docker
```bash
# build
docker build -t <your-docker-image-name> .

# run
docker run --rm -p 8000:8000 -p 9000:9000 -v </path/to/your/configs>:/data/conf <your-docker-image-name>
```

