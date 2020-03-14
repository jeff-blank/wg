JSDIR=		../public/js
JQSRC=          \
		${JSDIR}/bingo-county-select-g.js \
		${JSDIR}/bingos-index-g.js \
		${JSDIR}/entries-edit-g.js \
		${JSDIR}/hits-breakdown-g.js \
		${JSDIR}/hits-edit-g.js \
		${JSDIR}/masterStats-g.js

all:  ${JQSRC}

$(JSDIR)/%.js: %.go
	gopherjs build -m -o $@ $<
