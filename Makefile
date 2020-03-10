RELEASEDEST=	${HOME}/revel/wg.new
APPDIR=         src/github.com/jeff-blank/wg
JSDIR=		public/js
JQPROD=		jq/prod
JQSRC=          ${JQPROD}/bingo-county-select.js ${JQPROD}/bingos-index.js ${JQPROD}/entries-edit.js ${JQPROD}/hits-breakdown.js ${JQPROD}/hits-edit.js

pre-build:  ${JQSRC}

release: pre-build
	revel build -m prod github.com/jeff-blank/wg ${RELEASEDEST}
	sed -i '' '/runMode/s/^/exec /;s/wg\.new/wg/g' ${RELEASEDEST}/run.sh
	cp ${JQPROD}/*.js* ${RELEASEDEST}/${APPDIR}/${JSDIR}
	rm -f ${RELEASEDEST}/${APPDIR}/${JSDIR}/*.go

${JQPROD}/bingo-county-select.js:	${JSDIR}/bingo-county-select.go
	sed 's@/rvd/@/rv/@g' ${.ALLSRC} > ${JQPROD}/`basename ${.ALLSRC}`
	(cd ${JQPROD} && gopherjs build -m `basename ${.ALLSRC}`)

${JQPROD}/bingos-index.js:	${JSDIR}/bingos-index.go
	sed 's@/rvd/@/rv/@g' ${.ALLSRC} > ${JQPROD}/`basename ${.ALLSRC}`
	(cd ${JQPROD} && gopherjs build -m `basename ${.ALLSRC}`)

${JQPROD}/entries-edit.js:	${JSDIR}/entries-edit.go
	sed 's@/rvd/@/rv/@g' ${.ALLSRC} > ${JQPROD}/`basename ${.ALLSRC}`
	(cd ${JQPROD} && gopherjs build -m `basename ${.ALLSRC}`)

${JQPROD}/hits-breakdown.js:	${JSDIR}/hits-breakdown.go
	sed 's@/rvd/@/rv/@g' ${.ALLSRC} > ${JQPROD}/`basename ${.ALLSRC}`
	(cd ${JQPROD} && gopherjs build -m `basename ${.ALLSRC}`)

${JQPROD}/hits-edit.js:	${JSDIR}/hits-edit.go
	sed 's@/rvd/@/rv/@g' ${.ALLSRC} > ${JQPROD}/`basename ${.ALLSRC}`
	(cd ${JQPROD} && gopherjs build -m `basename ${.ALLSRC}`)
