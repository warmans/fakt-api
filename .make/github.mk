# Github API token for your account. This token just needs repository permissions. (REQUIRED)
GH_API_TOKEN ?=

# Github repo owner name. This will be your github username if you own the repo. (REQUIRED)
GH_REPO_OWNER ?=

# Github repo to create the release for. (REQUIRED)
GH_REPO_NAME ?=

# Version number given to release. The release tag will match this value. The release name and description also
# contain the version number. (REQUIRED)
RELEASE_VERSION ?= 0.0.0

# Release target. The tag is created from this commit-ish. Defaults to master. (OPTIONAL)
RELEASE_TARGET_COMMITISH ?= master

# Release is a draft true/false. Defaults to false. (OPTIONAL)
RELEASE_DRAFT ?= false

# Release is a pre-release true/false. Defaults to false. (OPTIONAL)
RELEASE_PRE ?= false

# Release artifact dir. If set all applicable FILES in this dir will be uploaded as release
# artifacts (see RELEASE_ARTIFACT_REGEX). If blank nothing is uploaded. (OPTIONAL)
RELEASE_ARTIFACT_DIR ?=

# Only files matching this regex in the RELEASE_ARTIFACT_DIR will be uploaded to the
# release. Defaults to evertying. (OPTIONAL)
RELEASE_ARTIFACT_REGEX ?= .*

# private constants
_RELEASE_INFO = .release-info.json
_RELEASE_PAYLOAD := {"tag_name":"$(RELEASE_VERSION)", "target_commitish":"$(RELEASE_TARGET_COMMITISH)", "name":"Release $(RELEASE_VERSION)", "body":"Release of version $(RELEASE_VERSION)", "draft":$(RELEASE_DRAFT), "prerelease":$(RELEASE_PRE)}

# Validate the required variables have been set.
.PHONY: _validate
_validate:
	@if [ -z $(GH_API_TOKEN) ]; then echo "GH_API_TOKEN must contain a valid github API access token"; exit 1; fi;
	@if [ -z $(GH_REPO_OWNER) ]; then echo "GH_REPO_OWNER must contain the github repo owner's username"; exit 1; fi;
	@if [ -z $(GH_REPO_NAME) ]; then echo "GH_REPO_NAME must contain the github repo name"; exit 1; fi;
	@if [ -z $(RELEASE_VERSION) ]; then echo "RELEASE_VERSION must be set"; exit 1; fi;

# Execute the release (Github API call).
.PHONY: _release
_release:
	@echo "releasing version $(RELEASE_VERSION)..."
	curl --data '$(_RELEASE_PAYLOAD)' https://api.github.com/repos/$(GH_REPO_OWNER)/$(GH_REPO_NAME)/releases?access_token=$(GH_API_TOKEN) | tee $(_RELEASE_INFO)

# Execute the artifact upload for the release (if required).
.PHONY: _artifacts
_artifacts:
	@if [ ! -z $(RELEASE_ARTIFACT_DIR) ]; then \
		if [ ! -f $(_RELEASE_INFO) ]; then echo "missing release info. Unable to publish artifacts without a release ID"; exit 1; fi; \
		echo "uploading artifacts..."; \
		for ARTIFACT in `find  $(RELEASE_ARTIFACT_DIR) -maxdepth 1 -regex $(RELEASE_ARTIFACT_REGEX) -type f -printf '%p\n'`; \
			do echo `basename $${ARTIFACT}`; \
			curl -H "Authorization: token $(GH_API_TOKEN)" -H "Content-Type: `file --mime-type $${ARTIFACT}`" \
			 --data-binary @$${ARTIFACT} \
			 "https://uploads.github.com/repos/$(GH_REPO_OWNER)/$(GH_REPO_NAME)/releases/`grep '^  "id":' $(_RELEASE_INFO) | tr -cd '[[:digit:]]'`/assets?name=`basename $${ARTIFACT}`"; \
		done; \
	fi;

# Cleanup any files created by release process.
.PHONY: _cleanup
_cleanup:
	rm -f $(_RELEASE_INFO)

# Meta-target to execute all targets in correct order.
.PHONY: release
release: _validate _release _artifacts _cleanup
