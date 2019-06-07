VERSION := $(shell git describe --tags --always --dirty="-dev")

release: github-release govendor clean dist
	github-release release \
		--security-token $$GITHUB_TOKEN \
		--user fiveai \
		--repo terraform-provider-k8s \
		--tag $(VERSION) \
		--name $(VERSION)

	#========================================================================
	# GNU/Linux
	#========================================================================
	# X86
	#------------------------------------------------------------------------
	github-release upload \
		--security-token $$GITHUB_TOKEN \
		--user fiveai \
		--repo terraform-provider-k8s \
		--tag $(VERSION) \
		--name aws-okta-$(VERSION)-linux-amd64 \
		--file dist/aws-okta-$(VERSION)-linux-amd64

	#------------------------------------------------------------------------
	# arm
	#------------------------------------------------------------------------
	github-release upload \
		--security-token $$GITHUB_TOKEN \
		--user fiveai \
		--repo terraform-provider-k8s \
		--tag $(VERSION) \
		--name aws-okta-$(VERSION)-linux-arm \
		--file dist/aws-okta-$(VERSION)-linux-arm

	github-release upload \
		--security-token $$GITHUB_TOKEN \
		--user fiveai \
		--repo terraform-provider-k8s \
		--tag $(VERSION) \
		--name aws-okta-$(VERSION)-linux-arm64 \
		--file dist/aws-okta-$(VERSION)-linux-arm64

	#========================================================================
	# macOS
	#========================================================================
	github-release upload \
		--security-token $$GITHUB_TOKEN \
		--user fiveai \
		--repo aws-okta \
		--tag $(VERSION) \
		--name aws-okta-$(VERSION)-darwin-amd64 \
		--file dist/aws-okta-$(VERSION)-darwin-amd64


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
