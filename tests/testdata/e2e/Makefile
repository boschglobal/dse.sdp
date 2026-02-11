# Copyright 2026 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

TMPDIR := .tmp
OUTDIR := $(CURDIR)


default: build

.PHONY: build
build:
	@echo "Generating e2e test data"
	rm -rf $(TMPDIR)
	mkdir -p $(TMPDIR)

	$(MAKE) generate-files TMPDIR=$(TMPDIR)
	$(MAKE) generate-lua   TMPDIR=$(TMPDIR)
	$(MAKE) generate-fmu   TMPDIR=$(TMPDIR)

	cp $(TMPDIR)/remote_foo.txt $(OUTDIR)/
	cp $(TMPDIR)/hello.lua      $(OUTDIR)/
	cp $(TMPDIR)/hello.fmu      $(OUTDIR)/
	cp $(TMPDIR)/files.zip  $(OUTDIR)/
	cp $(TMPDIR)/models.zip $(OUTDIR)/
	cp $(TMPDIR)/fmus.zip   $(OUTDIR)/

	rm -rf $(TMPDIR)
	@echo "Generated e2e test data"



.PHONY: generate-files
generate-files:
	mkdir -p $(TMPDIR)/testdata
	printf "remote foo\n" > $(TMPDIR)/remote_foo.txt
	cp $(TMPDIR)/remote_foo.txt $(TMPDIR)/testdata/remote_foo.txt
	cd $(TMPDIR) && zip -q files.zip testdata/remote_foo.txt


.PHONY: generate-lua
generate-lua:
	mkdir -p $(TMPDIR)/testdata
	printf 'function model_create()\nprint("Hello from Lua model")\nend\n\nfunction model_step()\nend\n\nfunction model_destroy()\nend\n' \
		> $(TMPDIR)/hello.lua
	cp $(TMPDIR)/hello.lua $(TMPDIR)/testdata/hello.lua
	cd $(TMPDIR) && zip -q models.zip testdata/hello.lua


.PHONY: generate-fmu
generate-fmu:
	mkdir -p $(TMPDIR)/build

	printf '%s\n' \
'/* Minimal self-contained FMI 2.0 stub */' \
'#include <stdint.h>' \
'typedef void* fmi2Component; typedef const char* fmi2String;' \
'typedef double fmi2Real; typedef int fmi2Boolean; typedef int fmi2Status;' \
'enum { fmi2OK = 0 }; typedef enum { fmi2ModelExchange = 0 } fmi2Type;' \
'typedef struct { void* a; } fmi2CallbackFunctions;' \
'__attribute__((visibility("default"))) fmi2Component fmi2Instantiate(fmi2String a,fmi2Type b,fmi2String c,fmi2String d,const fmi2CallbackFunctions* e,fmi2Boolean f,fmi2Boolean g){return (fmi2Component)1;}' \
'__attribute__((visibility("default"))) fmi2Status fmi2SetupExperiment(fmi2Component a,fmi2Boolean b,fmi2Real c,fmi2Real d,fmi2Boolean e,fmi2Real f){return fmi2OK;}' \
'__attribute__((visibility("default"))) fmi2Status fmi2EnterInitializationMode(fmi2Component a){return fmi2OK;}' \
'__attribute__((visibility("default"))) fmi2Status fmi2ExitInitializationMode(fmi2Component a){return fmi2OK;}' \
'__attribute__((visibility("default"))) fmi2Status fmi2Terminate(fmi2Component a){return fmi2OK;}' \
'__attribute__((visibility("default"))) void fmi2FreeInstance(fmi2Component a){}' \
	> $(TMPDIR)/build/hello.c

	gcc -fPIC -shared -o $(TMPDIR)/build/hello.so $(TMPDIR)/build/hello.c

	mkdir -p $(TMPDIR)/build/fmu/binaries/linux64
	cp $(TMPDIR)/build/hello.so $(TMPDIR)/build/fmu/binaries/linux64/

	printf '%s\n' \
'<?xml version="1.0" encoding="UTF-8"?>' \
'<fmiModelDescription fmiVersion="2.0" modelName="hello" guid="{12345678-1234-1234-1234-123456789abc}">' \
'<ModelExchange modelIdentifier="hello"/>' \
'<ModelVariables/>' \
'<ModelStructure/>' \
'</fmiModelDescription>' \
	> $(TMPDIR)/build/fmu/modelDescription.xml

	cd $(TMPDIR)/build/fmu && zip -qr ../hello.fmu .

	cp $(TMPDIR)/build/hello.fmu $(TMPDIR)/hello.fmu

	mkdir -p $(TMPDIR)/fmus/simple
	cp $(TMPDIR)/hello.fmu $(TMPDIR)/fmus/simple/hello.fmu
	cd $(TMPDIR) && zip -q fmus.zip fmus/simple/hello.fmu


.PHONY: cleanall
cleanall:
	rm -rf \
		remote_foo.txt \
		hello.lua \
		hello.fmu \
		files.zip \
		models.zip \
		fmus.zip