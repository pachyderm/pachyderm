include $(GOROOT)/src/Make.inc

TARG=pcre

CGOFILES=\
	pcre.go

include $(GOROOT)/src/Make.pkg

.PHONY: install-debian
install-debian:
	install -D _obj/$(TARG).a $(DESTDIR)/$(pkgdir)/$(TARG).a 
