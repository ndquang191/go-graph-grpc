FROM postgres:10.3

COPY ./up.sql /docker-entrypoint-initdb.d/

CMD ["postgres"]