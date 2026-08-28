FROM node:22-alpine

WORKDIR /opt/devarch-node-router
COPY --chown=node:node router.js ./router.js

ENV NODE_ENV=production \
    PORT=3000 \
    DEVARCH_NODE_TARGET_PORT=3000

USER node
EXPOSE 3000
CMD ["node", "router.js"]
