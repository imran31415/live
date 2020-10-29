# Backend for zoom integrated live streaming

## Quick Start: Building with Docker-compose

Requirements:
- Docker: https://docs.docker.com/docker-for-mac/install/

```
git clone git@github.com:nemo-ai/admin.git
cd admin
docker-compose up --build
```

You can now access the following URLS:
1. Admin Website: http://localhost:8183/admin
2. HTTP Json API: http://localhost:50052/get_user?id=2
3. GRPC API http://localhost:50051 
 
---

## Deploying to kubernetes:

Install gcloud and configure project:

```bash
gcloud init
export PROJECT_ID=livehub-277906
gcloud auth configure-docker
```

Build a docker image and push to gcloud container registry:
Note: ensure to bump version for each revision

```bash

export VERSION="390.0"
export PROJECT_ID=livehub-277906
export APP=backend

docker build -t gcr.io/${PROJECT_ID}/${APP}:${VERSION} .
docker push gcr.io/${PROJECT_ID}/${APP}:${VERSION}
kubectl apply -f ${APP}.yaml

```

- Build a staging.  (Note use with caution), this is just an example
````
export VERSION="287.0"
export PROJECT_ID=livehub-277906
export APP=backend-staging

docker build -t gcr.io/${PROJECT_ID}/${APP}:${VERSION} .
docker push gcr.io/${PROJECT_ID}/${APP}:${VERSION}
kubectl apply -f ${APP}.yaml

```

Ensure the docker image created matches the image in `backend.yaml` and apply the deployment to kubernetes:

```kubectl apply -f backend.yaml```


## Datadog instructions:

https://app.datadoghq.com/account/settings#agent/kubernetes
```
helm install ddog -f datadog-values.yaml --set datadog.apiKey=e07cd4173533639592e78cdf29e48d45 stable/datadog 
helm upgrade -f datadog-values.yaml ddog stable/datadog
```

## Building without Docker-compose
1. Install go
`brew install go`
`brew upgrade go`


2. Install protoc
 - *NOTE: this step is not required for build since these files should be generated and checked into git

    `go get github.com/golang/protobuf/protoc-gen-go`
    
    generate go bindings:
    
    ```
    protoc -I protos/ protos/nemo.proto --go_out=plugins=grpc:protos
    ```

    
    generate js bindings:
    
    `protoc -I protos/ protos/nemo.proto --js_out=library:protos `

3. Install mysql
`brew install mysql`

4. Configure mysql
`mv config.sample.json config.json` and update creds

7. Run GRPC/http server
`go run main.go`

Expexted output:
```bash
imran@MacBook-Pro admin % go run main.go 
2020/05/27 18:08:47 Running HTTP Server...
2020/05/27 18:08:47 Running GRPC Server....
```

You can now access the following URLS:
1. Admin Website: http://localhost:9033/admin
2. HTTP Json API: http://localhost:50052/get_user?id=2
3. GRPC API http://localhost:50051 
 
## Info

Nemo backend consists of the following key infra pieces:

#### Go GRPC/Protobuf server:

   - We define our proto interfaces in `protos/*.proto` files
   - Next, `protoc` is run on these files which auto generates language specific bindings:
        - see more info at: https://developers.google.com/protocol-buffers/docs/proto3
   - Our core server logic for the backend will be stored in `server/server.go`, which will use the grpc generated `proto/nemo.pb.go` bindings
   - Our client can also generate javascript bindings to do the same thing (I have generated some example javascript)
        - We dont have to do this though and we can generate an HTTP API if better 
        
### MYSQL database

   - NOTE: Our mysql connection params acre currently hardcoded in some places (TODO) 
        - `main.go` and `server/server.go` (TODO) move these to env vars
   - If developing locally without docker ensure the creds match your local mysql 

### Database ORM and Admin

   - php myadmin - vanilla php myadmin to control db
   - gorm as an ORM for the sql data that we can use in our server code
        - we want to be able to wrap the tables created via go-admin in an ORM so we can easily marshal/unmarshal from the database to our server layer:
            - see models/*.go for the ORM database wrapper code
            - Here we also marshal/unmarshal our db layer to our protobuf layer. 
                 - This way our db/server/ client layers are all clearly delineated

## Example Dev workflow for a new feature
   - Add RPCs/interfaces in `protos/nemo.proto`
   - Generate server/client bindings running protoc: 
        - `protoc -I protos/ protos/nemo.proto --proto_path=protos/api-common-protos/ --go_out=plugins=grpc:protos`
   - Update DB layer'
        - add code in server/repo.go and test in server/repo_test.go
   - Connect DB layer to proto layer:
        - add appropriate db marshallers/unmarshallers in `model/*.go` 
   - Now in our backend `server/server.go`, we can call the generated RPC methods to get/post/delete data. 
   - Rinse and repeat
   - See `GetUser` in server/server.go as an example of the above flow in action.

