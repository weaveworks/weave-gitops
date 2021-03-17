FROM node:lts

WORKDIR /usr/src/app

COPY package.json ./
COPY yarn.lock ./

RUN yarn install

EXPOSE 3000
CMD ["yarn", "run" ,"start", "--host", "0.0.0.0"]
