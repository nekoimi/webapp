# /etc/nginx/conf.d/default.conf

server {
    listen          __HTTP_PORT__ default_server;
    listen          [::]:__HTTP_PORT__ default_server;
    server_name     _;
    root   /usr/share/nginx/html;

    # index.html fallback
    location / {
        try_files $uri $uri/ /index.html;
    }

    # favicon.ico
    location = /favicon.ico {
        log_not_found off;
    }

    # robots.txt
    location = /robots.txt {
        log_not_found off;
    }

    # assets, media
    location ~* \.(?:css(\.map)?|js(\.map)?|jpe?g|png|gif|ico|cur|heic|webp|tiff?|mp3|m4a|aac|ogg|midi?|wav|mp4|mov|webm|mpe?g|avi|ogv|flv|wmv)$ {
        expires 7d;
    }

    # svg, fonts
    location ~* \.(?:svgz?|ttf|ttc|otf|eot|woff2?)$ {
        add_header Access-Control-Allow-Origin "*";
        expires    7d;
    }

    # gzip
    gzip            on;
    gzip_vary       on;
    gzip_proxied    any;
    gzip_comp_level 6;
    gzip_types      text/plain text/css text/xml application/json application/javascript application/rss+xml application/atom+xml image/svg+xml;

    # . files
    location ~ /\.(?!well-known) {
        deny all;
    }
}