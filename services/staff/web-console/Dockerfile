# syntax=docker/dockerfile:1.4
ARG NODE_IMAGE_VERSION=22-bookworm
ARG NGINX_IMAGE_VERSION=1.27-alpine
ARG NODE_IMAGE_REPOSITORY=mirror.gcr.io/library/node
ARG NGINX_IMAGE_REPOSITORY=mirror.gcr.io/library/nginx

FROM ${NODE_IMAGE_REPOSITORY}:${NODE_IMAGE_VERSION} AS deps

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

FROM deps AS build

COPY . .
RUN npm run build

FROM ${NGINX_IMAGE_REPOSITORY}:${NGINX_IMAGE_VERSION} AS prod

COPY docker/nginx/default.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]

FROM deps AS dev

COPY . .

ENV NODE_ENV=development

EXPOSE 5173

CMD ["sh", "-ec", "exec npm run dev -- --host 0.0.0.0 --port 5173"]
