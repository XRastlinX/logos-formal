FROM node:22-alpine

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci --ignore-scripts

COPY . .
RUN npm run build

ENV NODE_ENV=production
ENV ARK_PLACE=BEACON
ENV PORT=3000

EXPOSE 3000

CMD ["node", "dist/server.cjs"]
