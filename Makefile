PACKAGES=`go list ./... | grep -v vendor`
FILES=$(shell find . -name "*.go" | grep -v vendor)

# TODO: Remove 'should have comment'
lint:
	@echo -n ">> fmt" && \
	gofmt -s -l -w ${FILES} | awk '{print ""; print} END {if (NR > 0) {exit 1} else print "... ok"}' && \
	echo -n ">> lint" && \
	golint ${PACKAGES} | grep -v 'should have comment' | awk '{print} END {if (NR > 0) {exit 1} else print "... ok"}' && \
	echo -n ">> go vet" && \
	go vet ${PACKAGES} | awk '{print ""; print} END {if (NR > 0) {exit 1} else print "... ok"}' && \
	echo -n ">> long line" && \
	lll -g -l 100 ${FILES}  | awk '{print ""; print} END {if (NR > 0) {exit 1} else print "... ok"}' && \
	echo -n ">> errcheck" && \
	errcheck ${PACKAGES} | awk '{print ""; print} END {if (NR > 0) {exit 1} else print "... ok"}' && \
	echo -n ">> go simple" && \
	gosimple ${PACKAGES} | awk '{print ""; print} END {if (NR > 0) {exit 1} else print "... ok"}' && \
	echo -n ">> unused" && \
	unused ${PACKAGES} | awk '{print ""; print} END {if (NR > 0) {exit 1} else print "... ok"}'

