FROM scratch
ADD irondb-prometheus-adapter /irondb-prometheus-adapter
CMD ["/irondb-prometheus-adapter", "-log", "debug", "-addr", ":8080"]
