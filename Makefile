PLANNER := splanner

build:
	go build -o $(PLANNER)

bin:
	mkdir -p "$(DESTDIR)$(PREFIX)/bin"
	cp -f $(PLANNER) "$(DESTDIR)$(PREFIX)/bin"
	chmod 755 "$(DESTDIR)$(PREFIX)/bin/$(PLANNER)"
	rm $(PLANNER)

install: build bin
