FROM node:22-alpine

RUN apk add --no-cache su-exec \
    && corepack enable

COPY app-entrypoint.sh /usr/local/bin/devarch-node-app
RUN chmod 755 /usr/local/bin/devarch-node-app

WORKDIR /app
ENV HOST=0.0.0.0 \
    HOSTNAME=0.0.0.0 \
    PORT=3000 \
    NUXT_HOST=0.0.0.0 \
    NUXT_PORT=3000

EXPOSE 3000
ENTRYPOINT ["devarch-node-app"]
