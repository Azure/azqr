FROM gcr.io/distroless/static:nonroot

ARG BIN_PATH
EXPOSE 8080/tcp

WORKDIR /app
ADD $BIN_PATH/azqr .
ENTRYPOINT ["./azqr"]
CMD ["serve"]