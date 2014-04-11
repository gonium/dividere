BUILDDIR = build

all: $(BUILDDIR)
	@cd $(BUILDDIR) && go build ../.

$(BUILDDIR):
	@-mkdir $(BUILDDIR)

clean:
	@rm -rf $(BUILDDIR)
