generate-mocks:
	@mockgen -destination internal/service/mocks/auth_service_mock.go -source internal/service/auth.go
	@mockgen -destination internal/repository/mocks/token_repo_mock.go -source internal/repository/token.go
	@mockgen -destination internal/repository/mocks/user_repo_mock.go -source internal/repository/user.go
	@mockgen -destination internal/pkg/auth/mocks/access_mock.go -source internal/pkg/auth/access.go

test: generate-mocks
	go test ./...

cover-test: generate-mocks
	go test ./... -cover -coverprofile cover.out
	go tool cover -html=cover.out
	rm cover.out

deploy:
	docker-compose up --build