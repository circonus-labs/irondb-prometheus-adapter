FROM scratch
ADD irondb-prometheus-adapter /irondb-prometheus-adapter
CMD ["/irondb-prometheus-adapter", "-addr", ":8080"]
