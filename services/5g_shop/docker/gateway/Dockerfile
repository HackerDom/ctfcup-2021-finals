FROM nginx:1.21.0-alpine

COPY ./front/ /var/www
COPY ./nginx/conf.d/default.conf /etc/nginx/conf.d/

EXPOSE 4040

CMD ["nginx", "-g", "daemon off;"]

