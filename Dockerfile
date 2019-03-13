FROM scratch
ADD irondb-prometheus-adapter /irondb-prometheus-adapter
ENTRYPOINT ["/irondb-prometheus-adapter"]
CMD ["-addr", ":8080"]
