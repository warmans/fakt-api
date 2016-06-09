
# Version identifier to use for package. This should be something like 1.0.0 (REQUIRED)
PACKAGE_VERSION ?= $(RELEASE_VERSION)

# Location to stage the package contents before creating the archive. (OPTIONAL)
PACKAGE_CONTENT_DIR ?= .packaging

# Where to put the artifact (deb or whatever) (OPTIONAL)
PACKAGE_OUTPUT_DIR ?= .dist

# Type of package to build. This is passed directly to fpm.
PACKAGE_TYPE ?= deb (OPTIONAL)

# The package name (REQUIRED)
PACKAGE_NAME ?=

.PHONY: package
package: _configure_package

	@if [ -z "$(shell which fpm 2>/dev/null)" ]; then \
		echo "error:\nPackaging requires effing package manager (fpm) to run.\nsee https://github.com/jordansissel/fpm\n"; \
		exit 1; \
	fi

	mkdir -p $(PACKAGE_OUTPUT_DIR)

	#build package
	fpm --rpm-os linux \
		--force \
		-s dir \
		-p $(PACKAGE_OUTPUT_DIR) \
		-t $(PACKAGE_TYPE) \
		-n $(PACKAGE_NAME) \
		-v $(PACKAGE_VERSION) \
		-C $(PACKAGE_CONTENT_DIR) .

.PHONY: _configure_package
_configure_package:
	#you must override this with something that adds the relevant files to $(PACKAGE_CONTENT_DIR)