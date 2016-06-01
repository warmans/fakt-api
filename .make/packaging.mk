PACKAGE_VERSION ?=
PACKAGE_CONTENT_DIR ?= .packaging
PACKAGE_OUTPUT_DIR ?= .
PACKAGE_TYPE ?= deb
PACKAGE_NAME ?=

.PHONY: package
package:

	@if [ -z "$(shell which fpm 2>/dev/null)" ]; then \
		echo "error:\nPackaging requires effing package manager (fpm) to run.\nsee https://github.com/jordansissel/fpm\n"; \
		exit 1; \
	fi

	#run make install against the packaging dir
	mkdir -p ${PACKAGE_CONTENT_DIR} && $(MAKE) _configure_package

	mkdir -p ${PACKAGE_OUTPUT_DIR}

	#build package
	fpm \
		--rpm-os linux \
		--force \
		-s dir \
		-p ${PACKAGE_OUTPUT_DIR} \
		-t ${PACKAGE_TYPE} \
		-n ${PACKAGE_NAME} \
		-v ${PACKAGE_VERSION} \
		-C ${PACKAGE_CONTENT_DIR} .