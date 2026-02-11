# DevArch Varnish Configuration

vcl 4.1;

backend default {
    .host = "app";
    .port = "8080";
    .connect_timeout = 5s;
    .first_byte_timeout = 30s;
    .between_bytes_timeout = 10s;
    .probe = {
        .url = "/health";
        .interval = 10s;
        .timeout = 5s;
        .window = 5;
        .threshold = 3;
    }
}

acl purge {
    "localhost";
    "127.0.0.1";
    "172.16.0.0"/12;
}

sub vcl_recv {
    # Allow purging from trusted IPs
    if (req.method == "PURGE") {
        if (!client.ip ~ purge) {
            return(synth(405, "Not allowed."));
        }
        return (purge);
    }

    # Only cache GET and HEAD requests
    if (req.method != "GET" && req.method != "HEAD") {
        return (pass);
    }

    # Remove cookies for static files
    if (req.url ~ "\.(jpg|jpeg|gif|png|ico|css|zip|tgz|gz|rar|bz2|pdf|txt|tar|wav|bmp|rtf|js|flv|swf|html|htm|woff|woff2|svg|webp)$") {
        unset req.http.Cookie;
    }

    return (hash);
}

sub vcl_backend_response {
    # Cache static assets for longer
    if (bereq.url ~ "\.(jpg|jpeg|gif|png|ico|css|js|woff|woff2|svg|webp)$") {
        set beresp.ttl = 1h;
    } else {
        set beresp.ttl = 2m;
    }

    # Do not cache 50x responses
    if (beresp.status >= 500 && beresp.status < 600) {
        set beresp.uncacheable = true;
        return (deliver);
    }
}

sub vcl_deliver {
    # Add cache hit/miss header
    if (obj.hits > 0) {
        set resp.http.X-Cache = "HIT";
        set resp.http.X-Cache-Hits = obj.hits;
    } else {
        set resp.http.X-Cache = "MISS";
    }
}
