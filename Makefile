RELEASEDEST=	${HOME}/revel/wg.new
APPDIR=		src/github.com/jeff-blank/wg
JSDIR=		public/js
JQSRC=		jq
REINPLACE=	sed -i ''
APPCONF=	conf/app.conf
APPCONF_IN=	${APPCONF}.in

js:
	(cd ${JQSRC} && gmake all)

release: js ${APPCONF}
	${GOPATH}/bin/revel build -m prod github.com/jeff-blank/wg \
		${RELEASEDEST}
	${REINPLACE} '/runMode/s/^/exec /;s/wg\.new/wg/g' ${RELEASEDEST}/run.sh

${APPCONF}:
	@echo 'Check for $$NR_LICENSE...'
	@[ -n "${NR_LICENSE}" ]
	@echo 'Check for $$REVEL_SECRET...'
	@[ -n "${REVEL_SECRET}" ]
	@echo 'Check for $$DB_CONNECT_PROD...'
	@[ -n "${DB_CONNECT_PROD}" ]
	cp ${APPCONF_IN} $@
	@echo Write secrets to $@
	@${REINPLACE} 's/%%NR_LICENSE%%/'"${NR_LICENSE}"'/' $@
	@${REINPLACE} 's/%%REVEL_SECRET%%/'"${REVEL_SECRET}"'/' $@
	@${REINPLACE} 's,%%DB_CONNECT_PROD%%,'"${DB_CONNECT_PROD}"',' $@
