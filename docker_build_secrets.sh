#!/bin/sh

if [ -z "$SECRETS" -o ! -f "$SECRETS" ]; then
	echo "Secrets file '$SECRETS' not found" 1>&2
	exit 1
fi

i=0
while [ ${#REVEL_SECRET} -lt 64 ]; do
	char=$(dd if=/dev/random bs=1 count=1 2> /dev/null)
	case "$char" in
		[A-Za-z]) REVEL_SECRET="${REVEL_SECRET}${char}";;
		*) continue;;
	esac
	i=$(expr $i + 1)
	[ $i -gt 1000 ] && exit 1
done

. $SECRETS
docker build \
	--tag wg:latest \
	--build-arg NR_LICENSE_B="$NR_LICENSE" \
	--build-arg REVEL_SECRET_B="$REVEL_SECRET" \
	--build-arg DB_CONNECT_PROD_B="$DB_CONNECT_PROD" \
	.
