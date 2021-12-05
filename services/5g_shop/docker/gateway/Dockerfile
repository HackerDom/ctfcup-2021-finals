FROM nginx:1.21.0-alpine

COPY ./nginx/conf.d/default.conf /etc/nginx/conf.d/

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]

