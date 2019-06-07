VERSION := $(shell git describe --tags --always --dirty="-dev")

release: clean github-release dist
	github-release release \
		--user fiveai \
		--repo terraform-provider-k8s \
		--tag $(VERSION) \
		--name $(VERSION)
		--security-token $$GITHUB_TOKEN

	#========================================================================
	# GNU/Linux
	#========================================================================
	# X86
	#------------------------------------------------------------------------
	github-release upload \
		--user fiveai \
		--repo terraform-provider-k8s \
		--tag $(VERSION) \
		--name terraform-provider-k8s_$(VERSION)-linux-amd64 \
		--file terraform-provider-k8s_$(VERSION)-linux-amd64 \
		--security-token $$GITHUB_TOKEN

	#------------------------------------------------------------------------
	# arm
	#------------------------------------------------------------------------
	github-release upload \
		--user fiveai \
		--repo terraform-provider-k8s \
		--tag $(VERSION) \
		--name terraform-provider-k8s_$(VERSION)-linux-arm \
		--file terraform-provider-k8s_$(VERSION)-linux-arm \
		--security-token $$GITHUB_TOKEN

	github-release upload \
		--user fiveai \
		--repo terraform-provider-k8s \
		--tag $(VERSION) \
		--name terraform-provider-k8s_$(VERSION)-linux-arm64 \
		--file terraform-provider-k8s_$(VERSION)-linux-arm64 \
		--security-token $$GITHUB_TOKEN

	#========================================================================
	# macOS
	#========================================================================
	github-release upload \
		--user fiveai \
		--repo terraform-provider-k8s \
		--tag $(VERSION) \
		--name terraform-provider-k8s_$(VERSION)-darwin-amd64 \
		--file terraform-provider-k8s_$(VERSION)-darwin-amd64 \
		--security-token $$GITHUB_TOKEN


dist: goget
	#========================================================================
	# GNU/Linux
	#========================================================================
	# X86
	#------------------------------------------------------------------------
	GOOS=linux GOARCH=amd64 go build -o terraform-provider-k8s_$(VERSION)-linux-amd64

	#------------------------------------------------------------------------
	# arm
	#------------------------------------------------------------------------
	GOOS=linux GOARCH=arm go build -o terraform-provider-k8s_$(VERSION)-linux-arm
	GOOS=linux GOARCH=arm64 go build -o terraform-provider-k8s_$(VERSION)-linux-arm64

	#========================================================================
	# macOS
	#========================================================================
	GOOS=darwin GOARCH=amd64 go build -o terraform-provider-k8s_$(VERSION)-darwin-amd64

goget:
	go get

clean:
	rm -rf terraform-provider-k8s*

github-release:
	go get -u github.com/aktau/github-release

.PHONY: clean github-release
