vcl 4.1;

backend default {
    .host = "127.0.0.1";
    .port = "8080";
    .connect_timeout = 600s;
    .first_byte_timeout = 600s;
    .between_bytes_timeout = 600s;
}

sub vcl_recv {
    # Allow purging
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
    if (req.url ~ "\.(jpg|jpeg|gif|png|ico|css|zip|tgz|gz|rar|bz2|pdf|txt|tar|wav|bmp|rtf|js|flv|swf|html|htm)$") {
        unset req.http.Cookie;
    }

    return (hash);
}

sub vcl_backend_response {
    # Cache everything for 2 minutes
    set beresp.ttl = 2m;

    # Don't cache 50x responses
    if (beresp.status >= 500 && beresp.status < 600) {
        set beresp.uncacheable = true;
        return (deliver);
    }
}

sub vcl_deliver {
    # Add a header to see if it was a cache hit or miss
    if (obj.hits > 0) {
        set resp.http.X-Cache = "HIT";
    } else {
        set resp.http.X-Cache = "MISS";
    }
}

acl purge {
    "localhost";
    "127.0.0.1";
}
