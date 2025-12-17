FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY app .
EXPOSE 8080
USER nonroot
CMD ["./app"]
