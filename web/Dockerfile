ARG NODE_VERSION=21.1.0
FROM node:${NODE_VERSION}

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm install

COPY . .
RUN npm run build

EXPOSE 5173

CMD [ "npm", "run", "preview"]
