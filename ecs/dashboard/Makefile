all: docker

ui/dist: ui/src ui/public ui
	cd ui && npm run-script build

build: *.go ui/dist
	CGO_ENABLED=0 go build -installsuffix cgo -o ecs-dashboard -ldflags="-s -w"

bin: build
	gpg --armor --detach-sig ecs-dashboard
	cp ecs-dashboard ecs-dashboard.asc ../../bin

docker: bin
	docker build -t hamstah/ecs-dashboard .
