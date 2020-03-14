RELEASEDEST=	${HOME}/revel/wg.new
APPDIR=		src/github.com/jeff-blank/wg
JSDIR=		public/js
JQSRC=		jq

js:
	(cd ${JQSRC} && gmake all)

release: js
	revel build -m prod github.com/jeff-blank/wg ${RELEASEDEST}
	sed -i '' '/runMode/s/^/exec /;s/wg\.new/wg/g' ${RELEASEDEST}/run.sh
