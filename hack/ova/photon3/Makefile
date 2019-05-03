all: build

OVA := photon3-cloud-init.ova
OVF := photon3-cloud-init.ovf
VMDK := photon3-cloud-init-1.vmdk
OVF_EXPORT_DIR ?= $(HOME)/Downloads

$(VMDK): $(OVF_EXPORT_DIR)/$(VMDK)
	@cp -f $? $@

build: $(OVA)
ova: $(OVA)
$(OVA):	$(OVF) $(VMDK)
	@rm -f $@
	tar -cf $@ $^

upload: $(OVA)
	aws s3 cp $< s3://cnx.vmware/$< \
	  --grants read=uri=http://acs.amazonaws.com/groups/global/AllUsers && \
	  echo https://s3-us-west-2.amazonaws.com/cnx.vmware/$<
.PHONY: upload

clean:
	rm -f $(OVA) $(VMDK)
.PHONY: clean