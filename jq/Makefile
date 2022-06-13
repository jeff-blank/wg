JSDIR=		../public/js
JQSRC=          \
		${JSDIR}/newWin-g.js \
		${JSDIR}/app-index-g.js \
		${JSDIR}/bingo-county-select-g.js \
		${JSDIR}/bingos-index-g.js \
		${JSDIR}/entries-edit-g.js \
		${JSDIR}/hits-breakdown-g.js \
		${JSDIR}/hits-edit-g.js \
		${JSDIR}/hits-index-g.js \
		${JSDIR}/masterStats-g.js \
		${JSDIR}/reports-firstHits-g.js

all:  ${JQSRC}

$(JSDIR)/%.js: %.go
	gopherjs build -m -o $@ $<
