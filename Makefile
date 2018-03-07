###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################

include Configfile

SHELL := /bin/bash

.PHONY: my-version
my-version:
	$(eval IMAGE_VERSION := $(shell git rev-parse --short HEAD))

app-version: my-version

push: check-env app-version

include Makefile.docker
