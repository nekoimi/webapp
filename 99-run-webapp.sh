#!/bin/sh

HTTP_PORT=${PORT:-80}
echo "Listen on $HTTP_PORT ..."

if [ -f /etc/nginx/conf.d/default.conf ]; then
    rm -rf /etc/nginx/conf.d/default.conf
fi
cp /etc/nginx/conf.d/default.conf.tpl /etc/nginx/conf.d/default.conf
sed -i "s/__HTTP_PORT__/$HTTP_PORT/g" /etc/nginx/conf.d/default.conf

echo "Start deploy..."
webapp

chown -R nginx:nginx /usr/share/nginx/html && chmod -R 755 /usr/share/nginx/html

echo "Done."
